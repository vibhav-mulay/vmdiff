package internal

import (
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"log"

	"vmdiff/chunker"
	"vmdiff/utils"
)

type SigEntry struct {
	Sum    string `json:"checksum,omitempty"`
	Size   int64  `json:"size,omitempty"`
	Offset int64  `json:"offset"`
}

type Signature struct {
	Chunker string      `json:"chunker,omitempty"`
	Entries []*SigEntry `json:"entries,omitempty"`
	SumList []string    `json:"-"`
}

func EmptySignature() *Signature {
	return &Signature{}
}

func GenerateSignature(ctx context.Context, chunker chunker.Chunker) (*Signature, error) {
	sign := EmptySignature()
	sign.Chunker = chunker.Name()

	for {
		chunk, err := chunker.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		sigEnt := &SigEntry{
			Sum:    fmt.Sprintf("%x", md5.Sum(chunk.Data)),
			Size:   int64(chunk.Size),
			Offset: int64(chunk.Offset),
		}

		sign.Entries = append(sign.Entries, sigEnt)
		sign.SumList = append(sign.SumList, sigEnt.Sum)
	}

	log.Println("Signature generation complete")

	return sign, nil
}

func LoadSignature(ctx context.Context, r io.Reader) (*Signature, error) {
	s := EmptySignature()
	s.Load(ctx, r)

	s.SumList = func() []string {
		sumList := make([]string, len(s.Entries))
		for _, entry := range s.Entries {
			sumList = append(sumList, entry.Sum)
		}
		return sumList
	}()

	return s, nil
}

func (s *Signature) SumExists(sum string) bool {
	for _, item := range s.SumList {
		if item == sum {
			return true
		}
	}

	return false
}

func (s *Signature) Dump(ctx context.Context, w io.Writer) {
	utils.Dump(s, w)
}

func (s *Signature) Load(ctx context.Context, r io.Reader) {
	utils.Load(s, r)
}
