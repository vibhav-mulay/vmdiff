package internal

import "io"

type InputReader interface {
	io.Reader
	io.ReaderAt
}

type OutputWriter interface {
	io.Writer
	io.WriterAt
}
