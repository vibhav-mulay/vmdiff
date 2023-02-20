package chunker

import (
	"fmt"
	"io"
)

// Defines a chunk of data
type Chunk struct {
	Data        []byte
	Size        int64
	Offset      int64
	Fingerprint uint64
}

// Chunker interface wraps the Next() method
// Next() gives the next chunk
type Chunker interface {
	Name() string
	Next() (*Chunk, error)
}

// Invalid chunker string error
var ErrInvalidChunker = fmt.Errorf("Invalid chunker string")

// Returns the appropriate chunker implementation based on the input string
func GetChunker(cStr string, reader io.Reader) (Chunker, error) {
	switch cStr {
	case "fastcdc":
		return NewFastCDCChunker(reader)
	case "rabinfp":
		return NewRabinFPChunker(reader)
	default:
		return nil, ErrInvalidChunker
	}
}
