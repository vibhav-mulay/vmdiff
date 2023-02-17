package internal

import (
	"context"
	"log"
	"os"

	"vmdiff/chunker"
	"vmdiff/utils"

	"google.golang.org/protobuf/proto"
)

type Delta struct {
	infile       *os.File
	deltafile    *os.File
	writeCh      chan *DeltaEntry
	dumpComplete chan struct{}
}

const (
	Add    string = "add"
	Remove string = "remove"
	Copy   string = "copy"
)

func EmptyDelta(infile, deltafile *os.File) *Delta {
	return &Delta{
		infile:       infile,
		deltafile:    deltafile,
		writeCh:      make(chan *DeltaEntry, 20),
		dumpComplete: make(chan struct{}),
	}
}

func GenerateDelta(ctx context.Context, infile *os.File, signature *Signature, deltafile *os.File) (*Delta, error) {
	log.Printf("Initializing chunker: %s", signature.Chunker)
	chunker, err := chunker.GetChunker(signature.Chunker, infile)
	if err != nil {
		return nil, err
	}

	newsignature, err := GenerateSignature(ctx, chunker)
	if err != nil {
		return nil, err
	}

	delta := EmptyDelta(infile, deltafile)

	log.Println("Starting delta write to file goroutine")
	go delta.Dump(ctx)

	err = delta.CompareSignatures(ctx, signature, newsignature)
	if err != nil {
		return nil, err
	}

	return delta, nil
}

func (d *Delta) CompareSignatures(ctx context.Context, oldsig, newsig *Signature) error {
	var err error

	lcs := utils.DetermineLCS(oldsig.SumList, newsig.SumList)
	log.Printf("LCS: %v", lcs)

	oldSigIndex := 0
	oldSigLen := len(oldsig.Entries)
	newSigIndex := 0
	newSigLen := len(newsig.Entries)

	var deltaEnt *DeltaEntry

	for _, item := range lcs {
		for ; ; oldSigIndex++ {
			if oldsig.Entries[oldSigIndex].Sum == item {
				oldSigIndex++
				break
			}

			deltaEnt = d.DeltaRemoveEntry(oldsig.Entries[oldSigIndex])

			d.writeCh <- deltaEnt
		}

		for ; ; newSigIndex++ {
			if newsig.Entries[newSigIndex].Sum == item {
				newSigIndex++
				break
			}

			if oldsig.SumExists(newsig.Entries[newSigIndex].Sum) {
				deltaEnt = d.DeltaCopyEntry(newsig.Entries[newSigIndex])
			} else {
				deltaEnt = d.DeltaAddEntry(newsig.Entries[newSigIndex])
			}

			d.writeCh <- deltaEnt
		}
	}

	for ; oldSigIndex < oldSigLen; oldSigIndex++ {
		deltaEnt = d.DeltaRemoveEntry(oldsig.Entries[oldSigIndex])

		d.writeCh <- deltaEnt
	}

	for ; newSigIndex < newSigLen; newSigIndex++ {
		if oldsig.SumExists(newsig.Entries[newSigIndex].Sum) {
			deltaEnt = d.DeltaCopyEntry(newsig.Entries[newSigIndex])
		} else {
			deltaEnt = d.DeltaAddEntry(newsig.Entries[newSigIndex])
		}

		d.writeCh <- deltaEnt
	}

	close(d.writeCh)
	<-d.dumpComplete
	return err
}

func (d *Delta) Dump(ctx context.Context) {
	defer func() {
		d.dumpComplete <- struct{}{}
	}()

	for entry := range d.writeCh {
		d.WriteEntry(entry)
	}
}

func (d *Delta) WriteEntry(entry *DeltaEntry) {
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

	_, err = d.deltafile.Write(header)
	if err != nil {
		panic(err)
	}

	_, err = d.deltafile.Write(data)
	if err != nil {
		panic(err)
	}
}

func (d *Delta) DeltaRemoveEntry(sigEnt *SigEntry) *DeltaEntry {
	deltaEnt := &DeltaEntry{
		Action: Remove,
		Offset: sigEnt.Offset,
		Size:   sigEnt.Size,
	}

	return deltaEnt
}

func (d *Delta) DeltaAddEntry(sigEnt *SigEntry) *DeltaEntry {
	deltaEnt := &DeltaEntry{
		Action: Add,
		Offset: sigEnt.Offset,
		Size:   sigEnt.Size,
	}

	deltaEnt.Data = d.DataAt(deltaEnt.Offset, deltaEnt.Size)

	return deltaEnt
}

func (d *Delta) DeltaCopyEntry(sigEnt *SigEntry) *DeltaEntry {
	deltaEnt := &DeltaEntry{
		Action: Copy,
		Offset: sigEnt.Offset,
		Size:   sigEnt.Size,
	}

	deltaEnt.Sum = sigEnt.Sum

	return deltaEnt
}

func (d *Delta) DataAt(offset, size int64) []byte {
	data := make([]byte, size)

	_, err := d.infile.ReadAt(data, offset)
	if err != nil {
		panic(err)
	}
	log.Printf("Name=%s Fd=%d Offset=%d Size=%d", d.infile.Name(), d.infile.Fd(), offset, size)
	return data
}
