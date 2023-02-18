package chunker

import (
	"io"

	fastcdc "github.com/jotfs/fastcdc-go"
)

const (
	fastCDCName = "fastcdc"
	AVG_CHUNK   = 256 << 10
)

type FastCDCChunker struct {
	*fastcdc.Chunker
}

var _ Chunker = &FastCDCChunker{}

func NewFastCDCChunker(r io.Reader) (Chunker, error) {
	fc, err := fastcdc.NewChunker(r, fastcdc.Options{
		AverageSize: AVG_CHUNK,
	})
	if err != nil {
		return nil, err
	}

	return &FastCDCChunker{
		Chunker: fc,
	}, nil
}

func (fc *FastCDCChunker) Name() string {
	return fastCDCName
}

func (fc *FastCDCChunker) Next() (*Chunk, error) {
	chunk, err := fc.Chunker.Next()
	if err != nil {
		return nil, err
	}

	return &Chunk{
		Data:        chunk.Data,
		Size:        int64(chunk.Length),
		Offset:      int64(chunk.Offset),
		Fingerprint: chunk.Fingerprint,
	}, nil
}
