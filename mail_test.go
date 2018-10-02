package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreateMail(t *testing.T) {

}

func TestAddPaths(t *testing.T) {
	ctx := &Context{Paths: map[string]bool{"foo.docx": true}}

	res := fmt.Sprintf("%s  %s\n", pathsTmpl, "foo.docx")
	str := addPaths(ctx)
	assert.Equal(t, res, str, "should be equal")
}

func TestAddURLsBlock(t *testing.T) {
	ctx := &Context{URLs: map[string]string{"http://example.com/malware": "**BLOCK**"}}

	res := fmt.Sprintf("%s  %s\n", urlsTmpl, "http://example.com/malware")
	str := addURLs(ctx)
	assert.Equal(t, res, str, "should be equal")

}

func TestAddURLsUnknown(t *testing.T) {
	ctx := &Context{URLs: map[string]string{"http://example.com/malware": "UNKNOWN"}}

	str := addURLs(ctx)
	assert.Equal(t, urlsTmpl, str, "should be equal")
}

func TestDoSendMailNoMail(t *testing.T) {
	baseDir = "testdata"
	configName = "config.toml"
	fVerbose = false

	config, err := loadConfig()
	assert.NoError(t, err, "no error")
	ctx := &Context{
		config: config,
		Paths:  map[string]bool{"foo.docx": true},
	}

	err = doSendMail(ctx)
	assert.NoError(t, err, "no error")
}

func TestDoSendMailWithMail(t *testing.T) {
	baseDir = "testdata"
	configName = "config.toml"
	fVerbose = false

	config, err := loadConfig()
	assert.NoError(t, err, "no error")
	ctx := &Context{
		config: config,
		Paths:  map[string]bool{"foo.docx": true},
	}
	fDoMail = true

	err = doSendMail(ctx)
	assert.NoError(t, err, "no error")
}
