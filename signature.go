package vmdiff

import (
	"context"
	"crypto/md5"
	"fmt"
	"io"

	"github.com/vibhav-mulay/vmdiff/chunker"
	iproto "github.com/vibhav-mulay/vmdiff/proto"

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
	logger.Debugf("Using chunker %s", sign.Chunker)

	for {
		chunk, err := chunker.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			logger.Errorf("Chunker Next failed: %v", err)
			return nil, err
		}

		sigEnt := &iproto.SigEntry{
			Sum:    fmt.Sprintf("%x", md5.Sum(chunk.Data)),
			Size:   chunk.Size,
			Offset: chunk.Offset,
		}

		logger.Tracef("Adding signature entry %v", sigEnt)
		sign.Entries = append(sign.Entries, sigEnt)
		sign.SumList = append(sign.SumList, sigEnt.Sum)
	}

	logger.Debugf("Added %d signature entries", len(sign.Entries))
	logger.Infof("Signature generation complete")

	return sign, nil
}

// Load signature from file
func LoadSignature(ctx context.Context, r io.Reader) (*Signature, error) {
	s := EmptySignature()
	err := s.Load(ctx, r)
	if err != nil {
		return nil, err
	}

	s.SumList = func() []string {
		sumList := make([]string, 0, len(s.Entries))
		for _, entry := range s.Entries {
			sumList = append(sumList, entry.Sum)
		}
		return sumList
	}()

	logger.Infof("Signature load complete")
	logger.Debugf("Loaded %d entries", len(s.Entries))
	logger.Tracef("Entries: %v", s.Entries)

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
func (s *Signature) Dump(ctx context.Context, w io.Writer) error {
	data, err := proto.Marshal(s)
	if err != nil {
		logger.Errorf("Signature Dump Marshal failed: %v", err)
		return err
	}

	_, err = w.Write(data)
	if err != nil {
		logger.Errorf("Signature Dump Write failed: %v", err)
		return err
	}

	return nil
}

// Load the signature from a file
func (s *Signature) Load(ctx context.Context, r io.Reader) error {
	data, err := io.ReadAll(r)
	if err != nil {
		logger.Errorf("Signature Load ReadAll failed: %v", err)
		return err
	}

	err = proto.Unmarshal(data, s)
	if err != nil {
		logger.Errorf("Signature Load Unmarshal failed: %v", err)
		return err
	}

	return nil
}
