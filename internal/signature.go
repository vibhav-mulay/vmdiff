package internal

import (
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"log"

	"vmdiff/chunker"
	iproto "vmdiff/internal/proto"

	"google.golang.org/protobuf/proto"
)

// A signature describes the file data by breaking the data in chunks
type Signature struct {
	iproto.SigProto
	SumList []string `json:"-"`
}

// Return an empty siganture
func EmptySignature() *Signature {
	return &Signature{}
}

// Generate signature for the input file with information of each chunk (checksum, size and offset)
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

		sigEnt := &iproto.SigEntry{
			Sum:    fmt.Sprintf("%x", md5.Sum(chunk.Data)),
			Size:   chunk.Size,
			Offset: chunk.Offset,
		}

		sign.Entries = append(sign.Entries, sigEnt)
		sign.SumList = append(sign.SumList, sigEnt.Sum)
	}

	log.Println("Signature generation complete")

	return sign, nil
}

// Load signature from file
func LoadSignature(ctx context.Context, r io.Reader) (*Signature, error) {
	s := EmptySignature()
	s.Load(ctx, r)

	s.SumList = func() []string {
		sumList := make([]string, 0, len(s.Entries))
		for _, entry := range s.Entries {
			sumList = append(sumList, entry.Sum)
		}
		return sumList
	}()

	return s, nil
}

// Check whether a checksum is part of this signature
func (s *Signature) SumExists(sum string) (bool, int) {
	for i, item := range s.SumList {
		if item == sum {
			return true, i
		}
	}

	return false, -1
}

// Dump the signature to a file
func (s *Signature) Dump(ctx context.Context, w io.Writer) {
	data, err := proto.Marshal(s)
	if err != nil {
		panic(err)
	}

	_, err = w.Write(data)
	if err != nil {
		panic(err)
	}
}

// Load the signature from a file
func (s *Signature) Load(ctx context.Context, r io.Reader) {
	data, err := io.ReadAll(r)
	if err != nil {
		panic(err)
	}

	err = proto.Unmarshal(data, s)
	if err != nil {
		panic(err)
	}
}
