package chunker

import (
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var TestData string = "TESTSTRING"

func TestGetChunker(t *testing.T) {
	testcases := []struct {
		name string
		err  error
	}{
		{
			name: "fastcdc",
			err:  nil,
		},
		{
			name: "rabinfp",
			err:  nil,
		},
		{
			name: "other",
			err:  ErrInvalidChunker,
		},
	}

	for i, tc := range testcases {
		t.Run(fmt.Sprintf("TC %d", i+1), func(t *testing.T) {
			chunker, err := GetChunker(tc.name, strings.NewReader(TestData))
			if tc.err == nil {
				assert.Nil(t, err)
			} else {
				assert.Equal(t, err, tc.err)
				return
			}
			assert.Equal(t, chunker.Name(), tc.name)

			chunk, _ := chunker.Next()
			assert.Equal(t, TestData, string(chunk.Data))
			assert.Equal(t, int64(0), chunk.Offset)
			assert.Equal(t, int64(len(TestData)), chunk.Size)

			chunk, err = chunker.Next()
			assert.Nil(t, chunk)
			assert.Equal(t, err, io.EOF)
		})
	}
}
