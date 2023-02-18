package internal

import (
	"context"
	"io"
	"testing"

	"vmdiff/internal/proto"

	"github.com/stretchr/testify/assert"
)

var SEND_COUNT = 100
var RECV_COUNT = 0

func DummyWriteEntry(w io.Writer, entry *proto.DeltaEntry) {
	RECV_COUNT++
}

func TestFileDeltaDumper(t *testing.T) {
	d := NewFileDeltaDumper(nil)
	ctx := context.Background()

	d.StartDump(ctx, DummyWriteEntry)
	for i := 0; i < SEND_COUNT; i++ {
		d.Dump(&proto.DeltaEntry{})
	}
	d.EndDump()

	assert.Equal(t, SEND_COUNT, RECV_COUNT)
}
