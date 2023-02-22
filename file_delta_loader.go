package vmdiff

import (
	"context"
	"io"

	iproto "github.com/vibhav-mulay/vmdiff/proto"
)

// Loads the delta entries by reading from io.Reader and
// making them available to the caller
type FileDeltaLoader struct {
	deltaFile io.Reader
	readCh    chan *iproto.DeltaEntry
	err       error
}

// Creates FileDeltaLoader
func NewFileDeltaLoader(deltafile io.Reader) *FileDeltaLoader {
	return &FileDeltaLoader{
		deltaFile: deltafile,
		readCh:    make(chan *iproto.DeltaEntry, 20),
	}
}

var _ DeltaLoader = &FileDeltaLoader{}

// Start reading the delta entries
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
			l.err = err
			close(l.readCh)
			break
		}

		l.readCh <- deltaEnt
	}
}

// The caller calls this to get a stream of delta entries
func (l *FileDeltaLoader) Next() <-chan *iproto.DeltaEntry {
	return l.readCh
}

func (l *FileDeltaLoader) Err() error {
	return l.err
}
