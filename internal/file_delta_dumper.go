package internal

import (
	"context"
	"io"

	iproto "vmdiff/internal/proto"
)

type FileDeltaDumper struct {
	deltaFile    io.Writer
	writeCh      chan *iproto.DeltaEntry
	dumpComplete chan struct{}
}

func NewFileDeltaDumper(deltafile io.Writer) *FileDeltaDumper {
	return &FileDeltaDumper{
		deltaFile:    deltafile,
		writeCh:      make(chan *iproto.DeltaEntry, 20),
		dumpComplete: make(chan struct{}),
	}
}

var _ DeltaDumper = &FileDeltaDumper{}

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

func (d *FileDeltaDumper) EndDump() {
	close(d.writeCh)
	<-d.dumpComplete
}

func (d *FileDeltaDumper) Dump(entry *iproto.DeltaEntry) {
	d.writeCh <- entry
}
