package main


import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestOpenFile(t *testing.T) {
    file := "test/CIMBL-0666-CERTS.csv"
    fh, err := openFile(file)
    defer fh.Close()

    assert.Nil(t, err, "no error")
    assert.NotNil(t, fh, "not nil")

    file = "foo.bar"
    fh, err = openFile(file)
    defer fh.Close()

    assert.Nil(t, fh, "nil")
    assert.Error(t, err, "got error")
}
