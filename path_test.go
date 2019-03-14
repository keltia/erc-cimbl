package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandlePath(t *testing.T) {
	ctx := &Context{}

	path1 := "foo.exe"
	r, err := handlePath(ctx, path1)
	assert.Empty(t, r)
	assert.Error(t, err)

	path2 := "foo.docx"
	r, err = handlePath(ctx, path2)
	assert.NotEmpty(t, r)
	assert.NoError(t, err)
}

func TestHandlePathVerbose(t *testing.T) {
	ctx := &Context{}
	fVerbose = true

	path1 := "foo.exe"
	r, err := handlePath(ctx, path1)
	assert.Empty(t, r)
	assert.Error(t, err)

	path2 := "foo.docx"
	r, err = handlePath(ctx, path2)
	assert.NotEmpty(t, r)
	assert.NoError(t, err)
	fVerbose = false
}

func TestEntryToPath(t *testing.T) {
	str := "foo.doc|1a9ceab8d9b2358b46f2c767ccfc1317"

	val := entryToPath(str)
	assert.NotNil(t, val, "not nil")
	assert.NotEmpty(t, val, "not empty")
	assert.EqualValues(t, "foo.doc", val, "should be equal")
}
