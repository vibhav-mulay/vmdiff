package vmdiff

import (
	"context"
	"io"

	"github.com/vibhav-mulay/vmdiff/proto"
)

// InputReader interface groups io.Reader and io.ReaderAt
type InputReader interface {
	io.Reader
	io.ReaderAt
}

// OutputReader interface groups io.Writer and io.WriterAt
type OutputWriter interface {
	io.Writer
	io.WriterAt
}

// DeltaDumper handles the DeltaEntries as they are generated
type DeltaDumper interface {
	StartDump(context.Context, func(io.Writer, *proto.DeltaEntry))
	Dump(*proto.DeltaEntry)
	EndDump()
}

// Loads the delta entries and makes them available to the caller
type DeltaLoader interface {
	StartLoad(context.Context, func(io.Reader) (*proto.DeltaEntry, error))
	Next() <-chan *proto.DeltaEntry
}
