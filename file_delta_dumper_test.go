package vmdiff

import (
	"context"
	"io"
	"testing"

	"github.com/vibhav-mulay/vmdiff/proto"

	"github.com/stretchr/testify/assert"
)

var SEND_COUNT = 100
var RECV_COUNT = 0

func DummyWriteEntry(w io.Writer, entry *proto.DeltaEntry) error {
	RECV_COUNT++
	return nil
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
	assert.Nil(t, d.Err())
}
