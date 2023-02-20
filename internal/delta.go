package internal

import (
	"context"
	"io"
	"log"

	"vmdiff/chunker"
	iproto "vmdiff/internal/proto"
	"vmdiff/utils"

	"google.golang.org/protobuf/proto"
)

type DeltaGenerator struct {
	inFile InputReader
	dumper DeltaDumper
}

const (
	Add  string = "add"
	Copy string = "copy"
)

// Monkey patching for UT. I don't like this.
// Need to find a better way to do this.
var GetChunker = chunker.GetChunker

func NewDeltaGenerator(infile InputReader, d DeltaDumper) *DeltaGenerator {
	return &DeltaGenerator{
		inFile: infile,
		dumper: d,
	}
}

func (d *DeltaGenerator) GenerateDelta(ctx context.Context, signature *Signature) error {
	log.Printf("Initializing chunker: %s", signature.Chunker)
	chunker, err := GetChunker(signature.Chunker, d.inFile)
	if err != nil {
		return err
	}

	newsignature, err := GenerateSignature(ctx, chunker)
	if err != nil {
		return err
	}

	log.Println("Starting delta write to file goroutine")
	d.dumper.StartDump(ctx, WriteEntry)

	err = d.CompareSignatures(ctx, signature, newsignature)
	if err != nil {
		return err
	}

	return nil
}

func (d *DeltaGenerator) CompareSignatures(ctx context.Context, oldsig, newsig *Signature) error {
	var err error

	lcs := utils.DetermineLCS(oldsig.SumList, newsig.SumList)
	log.Printf("LCS: %v", lcs)

	newSigIndex := 0
	newSigLen := len(newsig.Entries)

	var deltaEnt *iproto.DeltaEntry

	for _, item := range lcs {
		for ; ; newSigIndex++ {
			if found, i := oldsig.SumExists(newsig.Entries[newSigIndex].Sum); found {
				deltaEnt = d.DeltaCopyEntry(newsig.Entries[newSigIndex], oldsig.Entries[i])
			} else {
				deltaEnt = d.DeltaAddEntry(newsig.Entries[newSigIndex])
			}

			d.dumper.Dump(deltaEnt)

			if newsig.Entries[newSigIndex].Sum == item {
				newSigIndex++
				break
			}
		}
	}

	for ; newSigIndex < newSigLen; newSigIndex++ {
		if found, i := oldsig.SumExists(newsig.Entries[newSigIndex].Sum); found {
			deltaEnt = d.DeltaCopyEntry(newsig.Entries[newSigIndex], oldsig.Entries[i])
		} else {
			deltaEnt = d.DeltaAddEntry(newsig.Entries[newSigIndex])
		}

		d.dumper.Dump(deltaEnt)
	}

	d.dumper.EndDump()
	return err
}

func (d *DeltaGenerator) DeltaAddEntry(sigEnt *iproto.SigEntry) *iproto.DeltaEntry {
	deltaEnt := &iproto.DeltaEntry{
		Action: Add,
		Offset: sigEnt.Offset,
		Size:   sigEnt.Size,
		Data:   d.DataAt(sigEnt.Offset, sigEnt.Size),
	}

	return deltaEnt
}

func (d *DeltaGenerator) DeltaCopyEntry(sigEnt, sigEntOld *iproto.SigEntry) *iproto.DeltaEntry {
	deltaEnt := &iproto.DeltaEntry{
		Action:    Copy,
		Offset:    sigEnt.Offset,
		Size:      sigEnt.Size,
		OldOffset: sigEntOld.Offset,
	}

	return deltaEnt
}

func (d *DeltaGenerator) DataAt(offset, size int64) []byte {
	data := make([]byte, size)

	_, err := d.inFile.ReadAt(data, offset)
	if err != nil {
		panic(err)
	}

	return data
}

func WriteEntry(w io.Writer, entry *iproto.DeltaEntry) {
	data, err := proto.Marshal(entry)
	if err != nil {
		panic(err)
	}

	dataLen := len(data)

	eheader := &iproto.EntryHeader{
		Size: uint64(dataLen),
	}

	header, err := proto.Marshal(eheader)
	if err != nil {
		panic(err)
	}

	_, err = w.Write(header)
	if err != nil {
		panic(err)
	}

	_, err = w.Write(data)
	if err != nil {
		panic(err)
	}
}
