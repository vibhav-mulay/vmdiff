package internal

import (
	"crypto/md5"
	"fmt"
	"io"
	"log"

	"vmdiff/chunker"
	"vmdiff/utils"
)

type SigEntry struct {
	Sum    string `json:"checksum,omitempty"`
	Size   uint64 `json:"size,omitempty"`
	Offset uint64 `json:"offset"`
}

type Signature struct {
	Chunker string      `json:"chunker,omitempty"`
	Entries []*SigEntry `json:"entries,omitempty"`
	SumList []string    `json:"-"`
}

var _ Dumpable = &Signature{}
var _ Loadable = &Signature{}

func EmptySignature() *Signature {
	return &Signature{}
}

func GenerateSignature(chunker chunker.Chunker) (*Signature, error) {
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
			Size:   uint64(chunk.Length),
			Offset: uint64(chunk.Offset),
		}

		sign.Entries = append(sign.Entries, sigEnt)
		sign.SumList = append(sign.SumList, sigEnt.Sum)
	}

	log.Println("Signature generation complete")

	return sign, nil
}

func LoadSignature(r io.Reader) (*Signature, error) {
	s := EmptySignature()
	s.Load(r)

	s.SumList = func() []string {
		sumList := make([]string, len(s.Entries))
		for _, entry := range s.Entries {
			sumList = append(sumList, entry.Sum)
		}
		return sumList
	}()

	return s, nil
}

func (s *Signature) Dump(w io.Writer) {
	utils.JSONDump(s, w)
}

func (s *Signature) Load(r io.Reader) {
	utils.JSONLoad(s, r)
}
