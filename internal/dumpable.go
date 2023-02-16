package internal

import (
	"io"
)

type Dumpable interface {
	Dump(io.Writer)
}
