package internal

import (
	"io"
	"log"
	"os"

	"vmdiff/chunker"
	"vmdiff/utils"
)

type DeltaEntry struct {
	Action string `json:"action"`
	Offset int64  `json:"offset"`
	Size   int64  `json:"size,omitempty"`
	Data   []byte `json:"data"`
}

type Delta struct {
	Entries []*DeltaEntry `json:"entries,omitempty"`
}

const (
	Add    string = "add"
	Remove string = "remove"
)

var _ Dumpable = &Delta{}

func EmptyDelta() *Delta {
	return &Delta{}
}

func GenerateDelta(infile *os.File, signature *Signature) (*Delta, error) {
	log.Printf("Initializing chunker: %s", signature.Chunker)
	chunker, err := chunker.GetChunker(signature.Chunker, infile)
	if err != nil {
		return nil, err
	}

	newsignature, err := GenerateSignature(chunker)
	if err != nil {
		return nil, err
	}

	delta := EmptyDelta()
	err = delta.CompareSignatures(infile, signature, newsignature)
	if err != nil {
		return nil, err
	}

	return delta, nil
}

func (d *Delta) CompareSignatures(infile *os.File, oldsig, newsig *Signature) error {
	var err error

	lcs := utils.DetermineLCS(oldsig.SumList, newsig.SumList)
	log.Printf("LCS: %v", lcs)

	oldSigIndex := 0
	oldSigLen := len(oldsig.Entries)
	newSigIndex := 0
	newSigLen := len(newsig.Entries)

	for _, item := range lcs {
		for ; ; oldSigIndex++ {
			if oldsig.Entries[oldSigIndex].Sum == item {
				oldSigIndex++
				break
			}

			deltaEnt := &DeltaEntry{
				Action: Remove,
				Offset: oldsig.Entries[oldSigIndex].Offset,
				Size:   oldsig.Entries[oldSigIndex].Size,
			}

			d.Entries = append(d.Entries, deltaEnt)
		}

		for ; ; newSigIndex++ {
			if newsig.Entries[newSigIndex].Sum == item {
				newSigIndex++
				break
			}

			deltaEnt := &DeltaEntry{
				Action: Add,
				Offset: newsig.Entries[newSigIndex].Offset,
				Size:   newsig.Entries[newSigIndex].Size,
			}

			deltaEnt.Data, err = d.DataAt(infile,
				deltaEnt.Offset,
				deltaEnt.Size)
			if err != nil {
				return err
			}

			d.Entries = append(d.Entries, deltaEnt)
		}
	}

	for ; oldSigIndex < oldSigLen; oldSigIndex++ {
		deltaEnt := &DeltaEntry{
			Action: Remove,
			Offset: oldsig.Entries[oldSigIndex].Offset,
			Size:   oldsig.Entries[oldSigIndex].Size,
		}

		d.Entries = append(d.Entries, deltaEnt)
	}

	for ; newSigIndex < newSigLen; newSigIndex++ {
		deltaEnt := &DeltaEntry{
			Action: Add,
			Offset: newsig.Entries[newSigIndex].Offset,
			Size:   newsig.Entries[newSigIndex].Size,
		}

		deltaEnt.Data, err = d.DataAt(infile,
			deltaEnt.Offset,
			deltaEnt.Size)
		if err != nil {
			return err
		}

		d.Entries = append(d.Entries, deltaEnt)
	}

	return err
}

func (d *Delta) Dump(w io.Writer) {
	utils.JSONDump(d, w)
}

func (d *Delta) DataAt(infile *os.File, offset, size int64) ([]byte, error) {
	data := make([]byte, size)

	_, err := infile.ReadAt(data, offset)
	if err != nil {
		return nil, err
	}
	log.Printf("Name=%s Fd=%d Offset=%d Size=%d", infile.Name(), infile.Fd(), offset, size)
	return data, nil
}
