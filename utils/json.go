package utils

import (
	//	"compress/gzip"
	"encoding/json"
	"io"
)

func Dump(v any, w io.Writer) {
	//	gz := gzip.NewWriter(w)
	//	defer gz.Close()

	err := json.NewEncoder(w).Encode(v)
	if err != nil {
		panic(err)
	}
}

func Load(v any, r io.Reader) {
	//	gz, err := gzip.NewReader(r)
	//	if err != nil {
	//		panic(err)
	//	}
	//	defer gz.Close()

	err := json.NewDecoder(r).Decode(v)
	if err != nil {
		panic(err)
	}
}
