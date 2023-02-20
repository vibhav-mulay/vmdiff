package internal

import (
	"context"
	"io"

	iproto "vmdiff/internal/proto"
)

type FileDeltaLoader struct {
	deltaFile io.Reader
	readCh    chan *iproto.DeltaEntry
}

func NewFileDeltaLoader(deltafile io.Reader) *FileDeltaLoader {
	return &FileDeltaLoader{
		deltaFile: deltafile,
		readCh:    make(chan *iproto.DeltaEntry, 20),
	}
}

var _ DeltaLoader = &FileDeltaLoader{}

func (l *FileDeltaLoader) StartLoad(ctx context.Context, entryLoader func(io.Reader) (*iproto.DeltaEntry, error)) {
	go l.startLoad(ctx, entryLoader)
}

func (l *FileDeltaLoader) startLoad(ctx context.Context, entryLoader func(io.Reader) (*iproto.DeltaEntry, error)) {
	for {
		deltaEnt, err := entryLoader(l.deltaFile)
		if err == io.EOF {
			close(l.readCh)
			break
		}
		if err != nil {
			panic(err)
		}

		l.readCh <- deltaEnt
	}
}

func (l *FileDeltaLoader) Next() <-chan *iproto.DeltaEntry {
	return l.readCh
}
