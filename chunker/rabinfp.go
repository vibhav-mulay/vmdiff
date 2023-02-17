package chunker

import (
	"io"

	restic "github.com/restic/chunker"
)

const (
	rabinFPName = "rabinfp"
	MIN_CHUNK   = 256 << 10
	MAX_CHUNK   = 4 << 20
	POL         = restic.Pol(0x39fc525c05db35)
)

type RabinFPChunker struct {
	*restic.Chunker
}

var _ Chunker = &RabinFPChunker{}

func NewRabinFPChunker(r io.Reader) (Chunker, error) {
	rc := restic.NewWithBoundaries(r, POL, MIN_CHUNK, MAX_CHUNK)

	return &RabinFPChunker{
		Chunker: rc,
	}, nil
}

func (rc *RabinFPChunker) Name() string {
	return rabinFPName
}

func (rc *RabinFPChunker) Next() (*Chunk, error) {
	chunk, err := rc.Chunker.Next(nil)
	if err != nil {
		return nil, err
	}

	return &Chunk{
		Data:        chunk.Data,
		Size:        int64(chunk.Length),
		Offset:      int64(chunk.Start),
		Fingerprint: chunk.Cut,
	}, nil
}
