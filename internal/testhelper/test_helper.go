package testhelper

import (
	"context"
	"fmt"
	"io"
	"testing"

	"vmdiff/chunker"
	"vmdiff/internal/proto"

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
