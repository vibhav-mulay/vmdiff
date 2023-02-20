package internal

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"

	"vmdiff/internal/proto"
	"vmdiff/internal/testhelper"

	"github.com/stretchr/testify/assert"
	gproto "google.golang.org/protobuf/proto"
)

var DeltaTestData = []string{"This", "is", "test", "data", "generated"}
var NewData = []string{"Welcome!", "This", "is", "newly", "generated", "data", "Bye!"}

// action:"add"  size:8  data:"Welcome!"
// action:"copy"  offset:8  size:4
// action:"copy"  offset:12  size:2  old_offset:4
// action:"add"  offset:14  size:5  data:"newly"
// action:"add"  offset:19  size:9  data:"generated"
// action:"copy"  offset:28  size:4  old_offset:10
// action:"add"  offset:32  size:4  data:"Bye!"
var Change = []*proto.DeltaEntry{
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

func TestDeltaGenerator(t *testing.T) {
	chunker := testhelper.NewTestChunker(DeltaTestData, false)
	ctx := context.Background()

	sign, err := GenerateSignature(ctx, chunker)
	assert.Nil(t, err)

	validator := testhelper.NewDeltaValidator(t, Change)

	input := strings.Join(NewData, "")
	reader := strings.NewReader(input)
	delta := NewDeltaGenerator(reader, validator)

	GetChunker = testhelper.GetChunker(NewData)
	err = delta.GenerateDelta(ctx, sign)
	assert.Nil(t, err)
}

func TestWriteEntry(t *testing.T) {
	buffer := bytes.NewBuffer(nil)

	entry := &proto.DeltaEntry{
		Action: Add, Offset: 1239,
		Size: 9201, Data: []byte("TESTDATA"),
	}

	WriteEntry(buffer, entry)

	reader := bytes.NewReader(buffer.Bytes())

	header := &proto.EntryHeader{Size: 2}
	headerLen := gproto.Size(header)
	headerData := make([]byte, headerLen)
	_, err := io.ReadFull(reader, headerData)
	assert.Nil(t, err)

	err = gproto.Unmarshal(headerData, header)
	assert.Nil(t, err)

	deltaEntSize := header.Size
	deltaEntData := make([]byte, deltaEntSize)
	deltaEnt := &proto.DeltaEntry{}

	_, err = io.ReadFull(reader, deltaEntData)
	assert.Nil(t, err)

	err = gproto.Unmarshal(deltaEntData, deltaEnt)
	assert.Nil(t, err)

	assert.Equal(t, true, gproto.Equal(entry, deltaEnt))
}
