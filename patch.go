package vmdiff

import (
	"context"
	"io"

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
	logger.Debugf("Starting delta read from file goroutine")
	p.loader.StartLoad(ctx, LoadEntry)

	if p.dryRun {
		logger.Infof("Dry run")
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
		if p.loader.Err() != nil {
			logger.Errorf("Error while loading: %v", p.loader.Err())
			return p.loader.Err()
		}
	}

	return nil
}

func (p *DeltaPatcher) addBlock(entry *iproto.DeltaEntry) error {
	logger.Debugf("Adding block to new file: size=%d", entry.Size)
	logger.Tracef("Entry: %v", entry)

	_, err := p.outFile.WriteAt(entry.Data, entry.Offset)
	if err != nil {
		logger.Errorf("addBlock WriteAt failed: %v", err)
		return err
	}

	return nil
}

func (p *DeltaPatcher) copyBlock(entry *iproto.DeltaEntry) error {
	logger.Debugf("Copying block from old file to new file: size=%d", entry.Size)
	logger.Tracef("Entry: %v", entry)

	data := make([]byte, entry.Size)
	_, err := p.inFile.ReadAt(data, entry.OldOffset)
	if err != nil {
		logger.Errorf("copyBlock ReadAt failed: %v", err)
		return err
	}

	_, err = p.outFile.WriteAt(data, entry.Offset)
	if err != nil {
		logger.Errorf("copyBlock WriteAt failed: %v", err)
		return err
	}

	return nil
}

// Do not do the actual patching. Just print the patching instructions
// The patching instructions are written to the log at INFO level
func (p *DeltaPatcher) DryRun() {
	for entry := range p.loader.Next() {
		entry.Data = nil
		logger.Infof("%v", entry)
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
		logger.Errorf("Header Read failed: %v", err)
		return nil, err
	}

	err = proto.Unmarshal(headerData, header)
	if err != nil {
		logger.Errorf("Header Unmarshal failed: %v", err)
		return nil, err
	}

	deltaEntSize := header.Size
	deltaEntData := make([]byte, deltaEntSize)
	deltaEnt := &iproto.DeltaEntry{}

	_, err = io.ReadFull(r, deltaEntData)
	if err != nil {
		logger.Errorf("Delta Read failed: %v", err)
		return nil, err
	}

	err = proto.Unmarshal(deltaEntData, deltaEnt)
	if err != nil {
		logger.Errorf("Delta Unmarshal failed: %v", err)
		return nil, err
	}

	return deltaEnt, nil
}
