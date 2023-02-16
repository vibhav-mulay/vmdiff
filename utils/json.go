package utils

import (
	"encoding/json"
	"io"
)

func JSONDump(v any, w io.Writer) {
	err := json.NewEncoder(w).Encode(v)
	if err != nil {
		panic(err)
	}
}

func JSONLoad(v any, r io.Reader) {
	err := json.NewDecoder(r).Decode(v)
	if err != nil {
		panic(err)
	}
}
