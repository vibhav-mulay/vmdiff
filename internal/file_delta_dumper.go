package internal

import (
	"context"
	"io"

	iproto "vmdiff/internal/proto"
)

// Implements DeltaDumper, writes the entry to the io.Writer
type FileDeltaDumper struct {
	deltaFile    io.Writer
	writeCh      chan *iproto.DeltaEntry
	dumpComplete chan struct{}
}

// Creates a new FileDeltaDumper
func NewFileDeltaDumper(deltafile io.Writer) *FileDeltaDumper {
	return &FileDeltaDumper{
		deltaFile:    deltafile,
		writeCh:      make(chan *iproto.DeltaEntry, 20),
		dumpComplete: make(chan struct{}),
	}
}

var _ DeltaDumper = &FileDeltaDumper{}

// The caller hints that dumping is about to start by calling this method
func (d *FileDeltaDumper) StartDump(ctx context.Context, entryWriter func(io.Writer, *iproto.DeltaEntry)) {
	go d.startDump(ctx, entryWriter)
}

func (d *FileDeltaDumper) startDump(ctx context.Context, entryWriter func(io.Writer, *iproto.DeltaEntry)) {
	defer func() {
		d.dumpComplete <- struct{}{}
	}()

	for entry := range d.writeCh {
		entryWriter(d.deltaFile, entry)
	}
}

// The caller hints that dumping has completed by calling this method
// This blocks till all the data in the channel is processed
func (d *FileDeltaDumper) EndDump() {
	close(d.writeCh)
	<-d.dumpComplete
}

// Dump a DeltaEntry
func (d *FileDeltaDumper) Dump(entry *iproto.DeltaEntry) {
	d.writeCh <- entry
}
