package internal

import (
	"context"
	"io"

	"vmdiff/internal/proto"
)

type InputReader interface {
	io.Reader
	io.ReaderAt
}

type OutputWriter interface {
	io.Writer
	io.WriterAt
}

type DeltaDumper interface {
	StartDump(context.Context, func(io.Writer, *proto.DeltaEntry))
	Dump(*proto.DeltaEntry)
	EndDump()
}
