package chunker

import (
	"fmt"
	"io"

	fastcdc "github.com/jotfs/fastcdc-go"
)

type Chunk struct {
	Data        []byte
	Size        uint64
	Offset      uint64
	Fingerprint uint64
}

type Chunker interface {
	Name() string
	Next() (fastcdc.Chunk, error)
}

func GetChunker(cStr string, reader io.Reader) (Chunker, error) {
	switch cStr {
	case "fastcdc":
		return NewFastCDCChunker(reader)
	default:
		return nil, fmt.Errorf("Invalid chunker string: %s", cStr)
	}
}
