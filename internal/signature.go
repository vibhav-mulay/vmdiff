package internal

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"

	"vmdiff/chunker"
)

type SigEntry struct {
	Sum    string `json:"sum,omitempty"`
	Size   uint64 `json:"size,omitempty"`
	Offset uint64 `json:"offset"`
}

type Signature struct {
	Chunker string      `json:"chunker,omitempty"`
	Entries []*SigEntry `json:"entries,omitempty"`
}

var _ Dumpable = &Signature{}

func GenerateSignature(chunker chunker.Chunker) (*Signature, error) {
	sign := &Signature{}
	sign.Chunker = chunker.Name()

	for {
		chunk, err := chunker.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		sigent := &SigEntry{
			Sum:    fmt.Sprintf("%x", md5.Sum(chunk.Data)),
			Size:   uint64(chunk.Length),
			Offset: uint64(chunk.Offset),
		}

		sign.Entries = append(sign.Entries, sigent)
	}

	return sign, nil
}

func (s *Signature) Dump(w io.Writer) {
	err := json.NewEncoder(w).Encode(s)
	if err != nil {
		panic(err)
	}
}
