package chunker

import (
	//	"fmt"
	"io"

	fastcdc "github.com/jotfs/fastcdc-go"
)

const name = "fastcdc"

type FastCDCChunker struct {
	fastcdc.Chunker
}

var _ Chunker = &FastCDCChunker{}

func NewFastCDCChunker(r io.Reader) (Chunker, error) {
	fc, err := fastcdc.NewChunker(r, fastcdc.Options{
		AverageSize: 1024 * 1024, // target 1 MiB average chunk size
	})
	if err != nil {
		return nil, err
	}

	return &FastCDCChunker{
		Chunker: *fc,
	}, nil
}

func (fc *FastCDCChunker) Name() string {
	return name
}

//func (fc *FastCDCChunker) Next() (*Chunk, error) {
//	return &Chunk{}, nil
//}
