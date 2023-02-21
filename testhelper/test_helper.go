package testhelper

import (
	"context"
	"fmt"
	"io"
	"testing"

	"github.com/vibhav-mulay/vmdiff/chunker"
	"github.com/vibhav-mulay/vmdiff/proto"

	"github.com/stretchr/testify/assert"
)

var ErrChunker = fmt.Errorf("Chunker Error")

type TestChunker struct {
	erroneous bool
	data      []string
	dataIndex int
	offset    int64
}

func NewTestChunker(d []string, e bool) chunker.Chunker {
	return &TestChunker{
		data:      d,
		erroneous: e,
	}
}

func (t *TestChunker) Name() string {
	return "testchunker"
}

func (t *TestChunker) Next() (chunk *chunker.Chunk, err error) {
	if t.erroneous {
		return nil, ErrChunker
	}

	if t.dataIndex == len(t.data) {
		return nil, io.EOF
	}

	chunk, err = &chunker.Chunk{
		Data:        []byte(t.data[t.dataIndex]),
		Offset:      t.offset,
		Size:        int64(len(t.data[t.dataIndex])),
		Fingerprint: uint64(12345),
	}, nil

	t.dataIndex++
	t.offset = t.offset + chunk.Size
	return
}

type DeltaValidator struct {
	t           *testing.T
	validations []*proto.DeltaEntry
	index       int
}

func NewDeltaValidator(t *testing.T, v []*proto.DeltaEntry) *DeltaValidator {
	return &DeltaValidator{
		t:           t,
		validations: v,
	}
}

func (d *DeltaValidator) StartDump(ctx context.Context, entryWriter func(io.Writer, *proto.DeltaEntry)) {
}

func (d *DeltaValidator) Dump(entry *proto.DeltaEntry) {
	assert.NotEqual(d.t, 0, len(d.validations)-d.index)
	assert.Equal(d.t, d.validations[d.index], entry)
	d.index++
}

func (d *DeltaValidator) EndDump() {
}

func GetChunker(input []string) func(string, io.Reader) (chunker.Chunker, error) {
	return func(cStr string, r io.Reader) (chunker.Chunker, error) {
		switch cStr {
		case "testchunker":
			fmt.Printf("XXX: %v\n", input)
			return NewTestChunker(input, false), nil
		default:
			return nil, chunker.ErrInvalidChunker
		}
	}
}

type TestDeltaLoader struct {
	entries []*proto.DeltaEntry
	readCh  chan *proto.DeltaEntry
}

func NewTestDeltaLoader(entries []*proto.DeltaEntry) *TestDeltaLoader {
	return &TestDeltaLoader{
		entries: entries,
		readCh:  make(chan *proto.DeltaEntry, 20),
	}
}

func (l *TestDeltaLoader) StartLoad(ctx context.Context, entryLoader func(io.Reader) (*proto.DeltaEntry, error)) {
	go l.startLoad(ctx, entryLoader)
}

func (l *TestDeltaLoader) startLoad(ctx context.Context, entryLoader func(io.Reader) (*proto.DeltaEntry, error)) {
	for _, entry := range l.entries {
		l.readCh <- entry
	}
	close(l.readCh)
}

func (l *TestDeltaLoader) Next() <-chan *proto.DeltaEntry {
	return l.readCh
}

type StringBuilder struct {
	s string
}

func (sb *StringBuilder) Write(p []byte) (int, error) {
	str := sb.s + string(p)
	sb.s = str
	return len(p), nil
}

func (sb *StringBuilder) WriteAt(p []byte, off int64) (n int, err error) {
	str := sb.s[:off] + string(p) + sb.s[off:]
	sb.s = str
	return len(p), nil
}

func (sb *StringBuilder) String() string {
	return sb.s
}
