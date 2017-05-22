package main

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestHandlePath(t *testing.T) {
    Plen := len(Paths)
    cntPaths = 0

    path1 := "foo.exe"
    handlePath(path1)
    assert.EqualValues(t, Plen, len(Paths), "same value")
    assert.Panics(t, func() {
            path1 = Paths[0]
        }, "should panic")

    path2 := "foo.docx"
    handlePath(path2)
    assert.EqualValues(t, 1, cntPaths, "plus one")
    assert.EqualValues(t, Plen + 1, len(Paths), "plus one")
    assert.EqualValues(t, path2, Paths[0], "inserted")
}

func TestEntryToPath(t *testing.T) {
    str := "foo.doc|1a9ceab8d9b2358b46f2c767ccfc1317"

    val := entryToPath(str)
    assert.NotNil(t, val, "not nil")
    assert.NotEmpty(t, val, "not empty")
    assert.EqualValues(t, "foo.doc", val, "should be equal")
}
