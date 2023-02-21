package vmdiff

import (
	"context"
	"io"
	"log"

	iproto "github.com/vibhav-mulay/vmdiff/proto"

	"google.golang.org/protobuf/proto"
)

// Carry out the instructions mentioned in the delta file
type DeltaPatcher struct {
	inFile  InputReader
	outFile OutputWriter
	loader  DeltaLoader
	dryRun  bool
}

// Creates a DeltaPatcher with the provided input
func NewDeltaPatcher(infile InputReader, outfile OutputWriter, loader DeltaLoader, dryRun bool) *DeltaPatcher {
	return &DeltaPatcher{
		inFile:  infile,
		outFile: outfile,
		loader:  loader,
		dryRun:  dryRun,
	}
}

// From the information given in the delta file, generate the updated file
// with the help of the old file
func (p *DeltaPatcher) PatchDelta(ctx context.Context) error {
	log.Println("Starting delta read from file goroutine")
	p.loader.StartLoad(ctx, LoadEntry)

	if p.dryRun {
		p.DryRun()
		return nil
	}

	var err error
	for entry := range p.loader.Next() {
		switch entry.Action {
		case Add:
			if err = p.addBlock(entry); err != nil {
				return err
			}
		case Copy:
			if err = p.copyBlock(entry); err != nil {
				return err
			}
		}
	}

	return nil
}

func (p *DeltaPatcher) addBlock(entry *iproto.DeltaEntry) error {
	_, err := p.outFile.WriteAt(entry.Data, entry.Offset)
	if err != nil {
		return err
	}

	return nil
}

func (p *DeltaPatcher) copyBlock(entry *iproto.DeltaEntry) error {
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

// Do not do the actual patching. Just print the patching instructions
func (p *DeltaPatcher) DryRun() {
	for entry := range p.loader.Next() {
		entry.Data = nil
		log.Printf("%v", entry)
	}
}

// Deserialize/Unmarshal the protobuf DeltaEntry after reading from the io.Reader
func LoadEntry(r io.Reader) (*iproto.DeltaEntry, error) {
	header := &iproto.EntryHeader{Size: 2}
	headerLen := proto.Size(header)
	headerData := make([]byte, headerLen)

	_, err := io.ReadFull(r, headerData)
	if err == io.EOF || err == io.ErrUnexpectedEOF {
		return nil, io.EOF
	}
	if err != nil {
		return nil, err
	}

	err = proto.Unmarshal(headerData, header)
	if err != nil {
		return nil, err
	}

	deltaEntSize := header.Size
	deltaEntData := make([]byte, deltaEntSize)
	deltaEnt := &iproto.DeltaEntry{}

	_, err = io.ReadFull(r, deltaEntData)
	if err != nil {
		return nil, err
	}

	err = proto.Unmarshal(deltaEntData, deltaEnt)
	if err != nil {
		return nil, err
	}

	return deltaEnt, nil
}