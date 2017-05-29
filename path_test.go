package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestHandlePath(t *testing.T) {
	ctx := &Context{
		Paths: map[string]bool{},
		URLs:  map[string]string{},
	}
	Plen := len(ctx.Paths)

	path1 := "foo.exe"
	handlePath(ctx, path1)
	assert.EqualValues(t, Plen, len(ctx.Paths), "same value")

	path2 := "foo.docx"
	handlePath(ctx, path2)
	assert.EqualValues(t, 1, len(ctx.Paths), "plus one")
	assert.EqualValues(t, Plen+1, len(ctx.Paths), "plus one")
	assert.EqualValues(t, true, ctx.Paths[path2], "inserted")
}

func TestHandlePathVerbose(t *testing.T) {
	ctx := &Context{
		Paths: map[string]bool{},
		URLs:  map[string]string{},
	}
	Plen := len(ctx.Paths)
	fVerbose = true

	path1 := "foo.exe"
	handlePath(ctx, path1)
	assert.EqualValues(t, Plen, len(ctx.Paths), "same value")

	path2 := "foo.docx"
	handlePath(ctx, path2)
	assert.EqualValues(t, 1, len(ctx.Paths), "plus one")
	assert.EqualValues(t, Plen+1, len(ctx.Paths), "plus one")
	assert.EqualValues(t, true, ctx.Paths[path2], "inserted")
}

func TestEntryToPath(t *testing.T) {
	str := "foo.doc|1a9ceab8d9b2358b46f2c767ccfc1317"

	val := entryToPath(str)
	assert.NotNil(t, val, "not nil")
	assert.NotEmpty(t, val, "not empty")
	assert.EqualValues(t, "foo.doc", val, "should be equal")
}
