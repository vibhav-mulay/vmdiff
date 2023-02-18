package chunker

import (
	"fmt"
	"io"
)

type Chunk struct {
	Data        []byte
	Size        int64
	Offset      int64
	Fingerprint uint64
}

type Chunker interface {
	Name() string
	Next() (*Chunk, error)
}

var ErrInvalidChunker = fmt.Errorf("Invalid chunker string")

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
