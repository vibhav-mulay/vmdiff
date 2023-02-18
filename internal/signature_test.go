package internal

import (
	"bytes"
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"testing"

	"vmdiff/chunker"

	"github.com/stretchr/testify/assert"
)

var TestData = []string{"This", "is", "test", "data"}
var ErrChunker = fmt.Errorf("Chunker Error")
var MarshalledSignature []byte

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

func TestEmptySignature(t *testing.T) {
	sign := EmptySignature()
	assert.Equal(t, len(sign.Entries), 0)
	assert.Equal(t, len(sign.SumList), 0)
}

func TestGenerateSignatureAndDump(t *testing.T) {
	testcases := []struct {
		name      string
		erroneous bool
		testdata  []string
	}{
		{
			name:      "Success",
			erroneous: false,
			testdata:  TestData,
		},
		{
			name:      "ChunkerError",
			erroneous: true,
			testdata:  TestData,
		},
	}

	for _, tc := range testcases {
		t.Run(fmt.Sprintf("%s", tc.name), func(t *testing.T) {
			chunker := NewTestChunker(tc.testdata, tc.erroneous)
			ctx := context.Background()

			sign, err := GenerateSignature(ctx, chunker)
			if tc.erroneous {
				assert.Equal(t, ErrChunker, err)
				return
			} else {
				assert.Nil(t, err)
			}

			ValidateSignatureEntries(t, sign, tc.testdata, chunker)

			buffer := bytes.NewBuffer(nil)
			sign.Dump(ctx, buffer)

			MarshalledSignature = buffer.Bytes()
			assert.NotEqual(t, 0, len(MarshalledSignature))

			reader := bytes.NewReader(MarshalledSignature)
			sign2, err := LoadSignature(ctx, reader)
			assert.Nil(t, err)

			ValidateSignatureEntries(t, sign2, tc.testdata, chunker)
		})
	}
}

func ValidateSignatureEntries(t *testing.T, sign *Signature, testdata []string, chunker chunker.Chunker) {
	assert.Equal(t, len(testdata), len(sign.Entries))
	assert.Equal(t, len(testdata), len(sign.SumList))
	assert.Equal(t, chunker.Name(), sign.Chunker)

	var offset int64
	var sum [16]byte
	for i, str := range testdata {
		sigEnt := sign.Entries[i]
		sum = md5.Sum([]byte(str))

		assert.Equal(t, fmt.Sprintf("%x", sum), sigEnt.Sum)
		assert.Equal(t, offset, sigEnt.Offset)
		assert.Equal(t, int64(len(str)), sigEnt.Size)

		offset += sigEnt.Size
	}
}
