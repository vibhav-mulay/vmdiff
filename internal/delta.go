package internal

import (
	"io"
	"log"

	"vmdiff/chunker"
	"vmdiff/utils"
)

type DeltaEntry struct {
	Action string `json:"action"`
	Offset uint64 `json:"offset"`
	Size   uint64 `json:"size,omitempty"`
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

func GenerateDelta(infile io.Reader, signature *Signature) (*Delta, error) {
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
	delta.CompareSignatures(signature, newsignature)

	return delta, nil
}

func (d *Delta) CompareSignatures(oldsig, newsig *Signature) {
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
				// TODO: Add data
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
			// TODO: Add data
		}

		d.Entries = append(d.Entries, deltaEnt)
	}
}

func (d *Delta) Dump(w io.Writer) {
	utils.JSONDump(d, w)
}
