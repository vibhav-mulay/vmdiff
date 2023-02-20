package internal

import (
	"context"
	"io"
	"testing"

	"vmdiff/internal/proto"

	"github.com/stretchr/testify/assert"
)

var COUNT = 100
var COUNTER = 0

func DummyLoadEntry(r io.Reader) (*proto.DeltaEntry, error) {
	COUNTER++
	if COUNTER > COUNT {
		return nil, io.EOF
	}

	return nil, nil
}

func TestFileDeltaLoader(t *testing.T) {
	d := NewFileDeltaLoader(nil)
	ctx := context.Background()
	recvCnt := 0

	d.StartLoad(ctx, DummyLoadEntry)
	for range d.Next() {
		recvCnt++
	}

	assert.Equal(t, COUNT, recvCnt)
}
