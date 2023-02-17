package internal

import (
	"context"
	"log"
	"os"

	"vmdiff/chunker"
	"vmdiff/utils"

	"google.golang.org/protobuf/proto"
)

type DeltaGenerator struct {
	inFile       *os.File
	deltaFile    *os.File
	writeCh      chan *DeltaEntry
	dumpComplete chan struct{}
}

const (
	Add  string = "add"
	Copy string = "copy"
)

func NewDeltaGenerator(infile, deltafile *os.File) *DeltaGenerator {
	return &DeltaGenerator{
		inFile:       infile,
		deltaFile:    deltafile,
		writeCh:      make(chan *DeltaEntry, 20),
		dumpComplete: make(chan struct{}),
	}
}

func (d *DeltaGenerator) GenerateDelta(ctx context.Context, signature *Signature) error {
	log.Printf("Initializing chunker: %s", signature.Chunker)
	chunker, err := chunker.GetChunker(signature.Chunker, d.inFile)
	if err != nil {
		return err
	}

	newsignature, err := GenerateSignature(ctx, chunker)
	if err != nil {
		return err
	}

	log.Println("Starting delta write to file goroutine")
	go d.StartDump(ctx)

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

	var deltaEnt *DeltaEntry

	for _, item := range lcs {
		for ; ; newSigIndex++ {
			if found, i := oldsig.SumExists(newsig.Entries[newSigIndex].Sum); found {
				deltaEnt = d.DeltaCopyEntry(newsig.Entries[newSigIndex], oldsig.Entries[i])
			} else {
				deltaEnt = d.DeltaAddEntry(newsig.Entries[newSigIndex])
			}

			d.writeCh <- deltaEnt

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

		d.writeCh <- deltaEnt
	}

	close(d.writeCh)
	<-d.dumpComplete
	return err
}

func (d *DeltaGenerator) StartDump(ctx context.Context) {
	defer func() {
		d.dumpComplete <- struct{}{}
	}()

	for entry := range d.writeCh {
		d.WriteEntry(entry)
	}
}

func (d *DeltaGenerator) WriteEntry(entry *DeltaEntry) {
	data, err := proto.Marshal(entry)
	if err != nil {
		panic(err)
	}

	dataLen := len(data)

	eheader := &EntryHeader{
		Size: uint64(dataLen),
	}

	header, err := proto.Marshal(eheader)
	if err != nil {
		panic(err)
	}

	_, err = d.deltaFile.Write(header)
	if err != nil {
		panic(err)
	}

	_, err = d.deltaFile.Write(data)
	if err != nil {
		panic(err)
	}
}

func (d *DeltaGenerator) DeltaAddEntry(sigEnt *SigEntry) *DeltaEntry {
	deltaEnt := &DeltaEntry{
		Action: Add,
		Offset: sigEnt.Offset,
		Size:   sigEnt.Size,
		Data:   d.DataAt(sigEnt.Offset, sigEnt.Size),
	}

	return deltaEnt
}

func (d *DeltaGenerator) DeltaCopyEntry(sigEnt, sigEntOld *SigEntry) *DeltaEntry {
	deltaEnt := &DeltaEntry{
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
