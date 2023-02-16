package internal

import (
	"io"
)

type Dumpable interface {
	Dump(io.Writer)
}

type Loadable interface {
	Load(io.Reader)
}
