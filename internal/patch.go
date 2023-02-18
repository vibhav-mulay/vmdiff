package internal

import (
	"context"
	"io"
	"log"

	iproto "vmdiff/internal/proto"

	"google.golang.org/protobuf/proto"
)

type DeltaPatcher struct {
	inFile    InputReader
	outFile   OutputWriter
	deltaFile io.Reader
	dryRun    bool

	readCh chan *iproto.DeltaEntry
}

func NewDeltaPatcher(infile InputReader, outfile OutputWriter, deltafile io.Reader, dryRun bool) *DeltaPatcher {
	return &DeltaPatcher{
		inFile:    infile,
		outFile:   outfile,
		deltaFile: deltafile,
		dryRun:    dryRun,
		readCh:    make(chan *iproto.DeltaEntry, 20),
	}
}

func (p *DeltaPatcher) PatchDelta(ctx context.Context) error {
	log.Println("Starting delta read from file goroutine")
	go p.StartLoad(ctx)

	if p.dryRun {
		p.DryRun()
		return nil
	}

	var err error
	for entry := range p.readCh {
		switch entry.Action {
		case Add:
			if err = p.AddBlock(entry); err != nil {
				return err
			}
		case Copy:
			if err = p.CopyBlock(entry); err != nil {
				return err
			}
		}
	}

	return nil
}

func (p *DeltaPatcher) AddBlock(entry *iproto.DeltaEntry) error {
	_, err := p.outFile.WriteAt(entry.Data, entry.Offset)
	if err != nil {
		return err
	}

	return nil
}

func (p *DeltaPatcher) CopyBlock(entry *iproto.DeltaEntry) error {
	data := make([]byte, entry.Size)
	_, err := p.inFile.ReadAt(data, entry.OldOffset)
	if err != nil {
		return err
	}

	_, err = p.outFile.WriteAt(data, entry.Offset)
	if err != nil {
		return err
	}

	return nil
}

func (p *DeltaPatcher) StartLoad(ctx context.Context) {
	for {
		header := &iproto.EntryHeader{Size: 2}
		headerLen := proto.Size(header)
		headerData := make([]byte, headerLen)

		_, err := io.ReadFull(p.deltaFile, headerData)
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			close(p.readCh)
			break
		}
		if err != nil {
			panic(err)
		}

		err = proto.Unmarshal(headerData, header)
		if err != nil {
			panic(err)
		}

		deltaEntSize := header.Size
		deltaEntData := make([]byte, deltaEntSize)
		deltaEnt := &iproto.DeltaEntry{}

		_, err = io.ReadFull(p.deltaFile, deltaEntData)
		if err != nil {
			panic(err)
		}

		err = proto.Unmarshal(deltaEntData, deltaEnt)
		if err != nil {
			panic(err)
		}

		p.readCh <- deltaEnt
	}
}

func (p *DeltaPatcher) DryRun() {
	for entry := range p.readCh {
		entry.Data = nil
		log.Printf("%v", entry)
	}
}
