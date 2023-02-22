package vmdiff

import (
	"bytes"
	"context"
	"crypto/md5"
	"fmt"
	"testing"

	"github.com/vibhav-mulay/vmdiff/chunker"
	"github.com/vibhav-mulay/vmdiff/testhelper"

	"github.com/stretchr/testify/assert"
)

var TestData = []string{"This", "is", "test", "data"}
var MarshalledSignature []byte

func TestEmptySignature(t *testing.T) {
	sign := EmptySignature()
	assert.Equal(t, len(sign.Entries), 0)
	assert.Equal(t, len(sign.SumList), 0)
}

func TestSignature(t *testing.T) {
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
		t.Run(tc.name, func(t *testing.T) {
			chunker := testhelper.NewTestChunker(tc.testdata, tc.erroneous)
			ctx := context.Background()

			sign, err := GenerateSignature(ctx, chunker)
			if tc.erroneous {
				assert.Equal(t, testhelper.ErrChunker, err)
				return
			} else {
				assert.Nil(t, err)
			}

			ValidateSignatureEntries(t, sign, tc.testdata, chunker)

			buffer := bytes.NewBuffer(nil)
			err = sign.Dump(ctx, buffer)
			assert.Nil(t, err)

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
