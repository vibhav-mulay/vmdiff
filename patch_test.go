package vmdiff

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"

	"github.com/vibhav-mulay/vmdiff/proto"
	"github.com/vibhav-mulay/vmdiff/testhelper"

	"github.com/stretchr/testify/assert"
	gproto "google.golang.org/protobuf/proto"
)

var PatchTestData = []string{"This", "is", "test", "data", "generated"}
var PatchNewData = []string{"Welcome!", "This", "is", "newly", "generated", "data", "Bye!"}

// action:"add"  size:8  data:"Welcome!"
// action:"copy"  offset:8  size:4
// action:"copy"  offset:12  size:2  old_offset:4
// action:"add"  offset:14  size:5  data:"newly"
// action:"add"  offset:19  size:9  data:"generated"
// action:"copy"  offset:28  size:4  old_offset:10
// action:"add"  offset:32  size:4  data:"Bye!"
var Entries = []*proto.DeltaEntry{
	{
		Action: Add, Size: 8,
		Data: []byte("Welcome!"), Offset: 0,
	},
	{
		Action: Copy, Offset: 8,
		Size: 4, OldOffset: 0,
	},
	{
		Action: Copy, Offset: 12,
		Size: 2, OldOffset: 4,
	},
	{
		Action: Add, Size: 5,
		Data: []byte("newly"), Offset: 14,
	},
	{
		Action: Copy, Size: 9,
		Offset: 19, OldOffset: 14,
	},
	{
		Action: Copy, Offset: 28,
		Size: 4, OldOffset: 10,
	},
	{
		Action: Add, Size: 4,
		Data: []byte("Bye!"), Offset: 32,
	},
}

func TestDeltaPatcher(t *testing.T) {
	loader := testhelper.NewTestDeltaLoader(Entries)

	input := strings.Join(PatchTestData, "")
	reader := strings.NewReader(input)
	writer := &testhelper.StringBuilder{}

	ctx := context.Background()
	patch := NewDeltaPatcher(reader, writer, loader, false)

	err := patch.PatchDelta(ctx)
	assert.Nil(t, err)

	output := writer.String()
	assert.Equal(t, strings.Join(PatchNewData, ""), output)
}

func TestDeltaPatcherDryRun(t *testing.T) {
	loader := testhelper.NewTestDeltaLoader(Entries)

	input := strings.Join(PatchTestData, "")
	reader := strings.NewReader(input)
	writer := &testhelper.StringBuilder{}

	ctx := context.Background()
	patch := NewDeltaPatcher(reader, writer, loader, true)

	err := patch.PatchDelta(ctx)
	assert.Nil(t, err)
}

func TestLoadEntry(t *testing.T) {
	buffer := bytes.NewBuffer(nil)

	entry := &proto.DeltaEntry{
		Action: Add, Offset: 1239,
		Size: 9201, Data: []byte("TESTDATA"),
	}

	data, err := gproto.Marshal(entry)
	assert.Nil(t, err)

	dataLen := len(data)
	eheader := &proto.EntryHeader{
		Size: uint64(dataLen),
	}

	header, err := gproto.Marshal(eheader)
	assert.Nil(t, err)

	_, err = buffer.Write(header)
	assert.Nil(t, err)

	_, err = buffer.Write(data)
	assert.Nil(t, err)

	reader := bytes.NewReader(buffer.Bytes())

	deltaEnt, err := LoadEntry(reader)
	assert.Nil(t, err)
	assert.Equal(t, true, gproto.Equal(entry, deltaEnt))

	_, err = LoadEntry(reader)
	assert.Equal(t, io.EOF, err)
}
