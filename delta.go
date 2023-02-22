package vmdiff

import (
	"context"
	"io"

	"github.com/vibhav-mulay/vmdiff/chunker"
	iproto "github.com/vibhav-mulay/vmdiff/proto"
	"github.com/vibhav-mulay/vmdiff/utils"

	"google.golang.org/protobuf/proto"
)

// Generates the delta/change in the two file signatures
// A delta entry is an instruction for the DeltaPatcher describing
// where and how to get the data for a particular offset
type DeltaGenerator struct {
	inFile InputReader
	dumper DeltaDumper
}

// A delta entry can have one of these two actions
// Add is new data. The "new" data is part of the delta entry
// Copy tells the DeltaPatcher to get the data for the particular chunk
// for the old file itself. It also tells where to look at within the old file
const (
	Add  string = "add"
	Copy string = "copy"
)

// Monkey patching for UT. I don't like this.
// Need to find a better way to do this.
var GetChunker = chunker.GetChunker

// Creates a DeltaGenerator with the provided input
func NewDeltaGenerator(infile InputReader, d DeltaDumper) *DeltaGenerator {
	return &DeltaGenerator{
		inFile: infile,
		dumper: d,
	}
}

// Generate the delta describing changes between the two signatures
func (d *DeltaGenerator) GenerateDelta(ctx context.Context, signature *Signature) error {
	logger.Infof("Initializing chunker: %s", signature.Chunker)
	chunker, err := GetChunker(signature.Chunker, d.inFile)
	if err != nil {
		logger.Errorf("GetChunker failed: %v", err)
		return err
	}

	newsignature, err := GenerateSignature(ctx, chunker)
	if err != nil {
		logger.Errorf("GenerateSignature failed: %v", err)
		return err
	}
	logger.Infof("New file signature generation complete")

	logger.Debugf("Starting delta write to file goroutine")
	d.dumper.StartDump(ctx, WriteEntry)

	logger.Debugf("Comparing signatures")
	err = d.CompareSignatures(ctx, signature, newsignature)
	if err != nil {
		logger.Errorf("CompareSignatures failed: %v", err)
		return err
	}

	return nil
}

// Compare two signatures and generate the resultant change in the form of Delta
func (d *DeltaGenerator) CompareSignatures(ctx context.Context, oldsig, newsig *Signature) error {
	var err error

	// Find all the common chunks in the two signatures using Longest Common Subsequence algorithm.
	lcs := utils.DetermineLCS(oldsig.SumList, newsig.SumList)
	logger.Tracef("LCS: %v", lcs)

	newSigIndex := 0
	newSigLen := len(newsig.Entries)

	var deltaEnt *iproto.DeltaEntry

	for _, item := range lcs {
		// All the chunks between two common chunks are new/changed chunks
		for ; ; newSigIndex++ {
			// If a new chunk is already present in the older signature at some different offset,
			// issue a copy command in the resultant delta. This allows to use the chunk from the older file,
			// during patch instead of including the data in the delta file. This helps in reducing the size of
			// the delta file
			if found, i := oldsig.SumExists(newsig.Entries[newSigIndex].Sum); found {
				deltaEnt = d.deltaCopyEntry(newsig.Entries[newSigIndex], oldsig.Entries[i])
			} else {
				deltaEnt = d.deltaAddEntry(newsig.Entries[newSigIndex])
			}

			// Send the delta entry to the dumper
			d.dumper.Dump(deltaEnt)

			if newsig.Entries[newSigIndex].Sum == item {
				newSigIndex++
				break
			}
		}

		if d.dumper.Err() != nil {
			logger.Errorf("Error while dumping: %v", d.dumper.Err())
			return d.dumper.Err()
		}
	}

	// After the last common chunk, treat everything as a new/changed chunks
	for ; newSigIndex < newSigLen; newSigIndex++ {
		if found, i := oldsig.SumExists(newsig.Entries[newSigIndex].Sum); found {
			deltaEnt = d.deltaCopyEntry(newsig.Entries[newSigIndex], oldsig.Entries[i])
		} else {
			deltaEnt = d.deltaAddEntry(newsig.Entries[newSigIndex])
		}

		// Send the delta entry to the dumper
		d.dumper.Dump(deltaEnt)
		if d.dumper.Err() != nil {
			logger.Errorf("Error while dumping: %v", d.dumper.Err())
			return d.dumper.Err()
		}
	}

	// End dumping
	d.dumper.EndDump()
	if d.dumper.Err() != nil {
		logger.Errorf("Error while dumping: %v", d.dumper.Err())
		return d.dumper.Err()
	}

	return err
}

func (d *DeltaGenerator) deltaAddEntry(sigEnt *iproto.SigEntry) *iproto.DeltaEntry {
	deltaEnt := &iproto.DeltaEntry{
		Action: Add,
		Offset: sigEnt.Offset,
		Size:   sigEnt.Size,
	}

	logger.Debugf("Added delta entry: %v", deltaEnt)

	deltaEnt.Data = d.dataAt(sigEnt.Offset, sigEnt.Size)
	return deltaEnt
}

func (d *DeltaGenerator) deltaCopyEntry(sigEnt, sigEntOld *iproto.SigEntry) *iproto.DeltaEntry {
	deltaEnt := &iproto.DeltaEntry{
		Action:    Copy,
		Offset:    sigEnt.Offset,
		Size:      sigEnt.Size,
		OldOffset: sigEntOld.Offset,
	}

	logger.Debugf("Added delta entry: %v", deltaEnt)
	return deltaEnt
}

func (d *DeltaGenerator) dataAt(offset, size int64) []byte {
	data := make([]byte, size)

	_, err := d.inFile.ReadAt(data, offset)
	if err != nil {
		logger.Errorf("dataAt ReadAt failed: %v", err)
		panic(err)
	}

	return data
}

// Serialize/Marshal the protobuf DeltaEntry and write to the io.Writer
// The object is dumped in the format -->Header-->Data-->Header-->Data-->
// Where header is a fixed size header containing size of the data to follow
func WriteEntry(w io.Writer, entry *iproto.DeltaEntry) error {
	data, err := proto.Marshal(entry)
	if err != nil {
		logger.Errorf("Data Marshal failed: %v", err)
		return err
	}

	dataLen := len(data)

	eheader := &iproto.EntryHeader{
		Size: uint64(dataLen),
	}

	header, err := proto.Marshal(eheader)
	if err != nil {
		logger.Errorf("Header Marshal failed: %v", err)
		return err
	}

	_, err = w.Write(header)
	if err != nil {
		logger.Errorf("Header Write failed: %v", err)
		return err
	}

	_, err = w.Write(data)
	if err != nil {
		logger.Errorf("Delta Write failed: %v", err)
		return err
	}

	return nil
}
