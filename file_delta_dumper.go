package vmdiff

import (
	"context"
	"io"

	iproto "github.com/vibhav-mulay/vmdiff/proto"
)

// Implements DeltaDumper, writes the entry to the io.Writer
type FileDeltaDumper struct {
	deltaFile    io.Writer
	writeCh      chan *iproto.DeltaEntry
	dumpComplete chan struct{}

	err error
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
func (d *FileDeltaDumper) StartDump(ctx context.Context, entryWriter func(io.Writer, *iproto.DeltaEntry) error) {
	go d.startDump(ctx, entryWriter)
}

func (d *FileDeltaDumper) startDump(ctx context.Context, entryWriter func(io.Writer, *iproto.DeltaEntry) error) {
	defer func() {
		d.dumpComplete <- struct{}{}
	}()

	for entry := range d.writeCh {
		err := entryWriter(d.deltaFile, entry)
		if err != nil {
			d.err = err
		}
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
	if d.Err() != nil {
		return
	}

	d.writeCh <- entry
}

// Returns if dumping faced any error
func (d *FileDeltaDumper) Err() error {
	return d.err
}
