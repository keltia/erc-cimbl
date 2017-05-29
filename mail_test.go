package main

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "fmt"
)

func TestCreateMail(t *testing.T) {

}

func TestAddPaths(t *testing.T) {
    ctx := &Context{Paths: map[string]bool{"foo.docx": true, }}


    res := fmt.Sprintf("%s  %s\n", PathsTmpl, "foo.docx")
    str := addPaths(ctx)
    assert.Equal(t, res, str, "should be equal")
}

func TestAddURLsBlock(t *testing.T) {
    ctx := &Context{URLs: map[string]string{"http://example.com/malware": "**BLOCK**", }}

    res := fmt.Sprintf("%s  %s\n", URLsTmpl, "http://example.com/malware")
    str := addURLs(ctx)
    assert.Equal(t, res, str, "should be equal")


}

func TestAddURLsUnknown(t *testing.T) {
    ctx := &Context{URLs: map[string]string{"http://example.com/malware": "UNKNOWN", }}

    str := addURLs(ctx)
    assert.Equal(t, URLsTmpl, str, "should be equal")
}

func TestDoSendMail(t *testing.T) {

}
