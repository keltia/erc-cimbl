package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestOpenFileBad(t *testing.T) {
	file := "foo.bar"
	fh, err := openFile(file)
	defer fh.Close()

	assert.Nil(t, fh, "nil")
	assert.Error(t, err, "got error")
}

func TestOpenFileGood(t *testing.T) {
	file := "test/CIMBL-0666-CERTS.csv"
	fh, err := openFile(file)
	defer fh.Close()

	assert.NoError(t, err, "no error")
	assert.NotNil(t, fh, "not nil")
}

func TestParseCSVNone(t *testing.T) {
	file := "/noneexistent"
	ctx := &Context{}
	err := handleCSV(ctx, file)

	assert.Error(t, err, "should be in error")
}

func TestHandleCSV(t *testing.T) {
	file := "test/CIMBL-0666-CERTS.csv"
	config, err := loadConfig()
	assert.NoError(t, err, "no error")

	ctx := &Context{
		config: config,
		Paths:  map[string]bool{},
		URLs:   map[string]string{},
	}

	realPaths := map[string]bool{
		"55fe62947f3860108e7798c4498618cb.rtf": true,
	}
	realURLs := map[string]string{
		"http://pontonerywariva342.top/search.php": "**BLOCK**",
	}
	err = handleCSV(ctx, file)
	assert.NoError(t, err, "no error")
	assert.Equal(t, realPaths, ctx.Paths, "should be equal")
	assert.Equal(t, realURLs, ctx.URLs, "should be equal")
}

func TestHandleCSVVerbose(t *testing.T) {
	file := "test/CIMBL-0666-CERTS.csv"
	config, err := loadConfig()
	assert.NoError(t, err, "no error")

	fVerbose = true
	ctx := &Context{
		config: config,
		Paths:  map[string]bool{},
		URLs:   map[string]string{},
	}

	realPaths := map[string]bool{
		"55fe62947f3860108e7798c4498618cb.rtf": true,
	}
	realURLs := map[string]string{
		"http://pontonerywariva342.top/search.php": "**BLOCK**",
	}
	err = handleCSV(ctx, file)
	assert.NoError(t, err, "no error")
	assert.Equal(t, realPaths, ctx.Paths, "should be equal")
	assert.Equal(t, realURLs, ctx.URLs, "should be equal")
}
