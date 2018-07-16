package main

import (
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestOpenFileBad(t *testing.T) {
	file := "foo.bar"
	ctx := &Context{}
	fn, err := openFile(ctx, file)

	assert.Empty(t, fn, "empty")
	assert.Error(t, err, "got error")
}

func TestOpenFileGood(t *testing.T) {
	file := "test/CIMBL-0666-CERTS.csv"
	ctx := &Context{}
	fn, err := openFile(ctx, file)

	assert.NoError(t, err, "no error")
	assert.NotEmpty(t, fn, "not empty")
}

func TestParseCSVNone(t *testing.T) {
	file := "/noneexistent"
	ctx := &Context{}
	err := handleSingleFile(ctx, file)

	assert.Error(t, err, "should be in error")
}

func TestHandleCSV(t *testing.T) {
	var testSite string

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
		TestSite: ActionBlocked,
	}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	if proxyURL != nil {
		testSite = proxyURL.Host
	} else {
		testSite = TestSite
	}

	// mock to add a new measurement
	httpmock.RegisterResponder("HEAD", testSite,
		func(req *http.Request) (*http.Response, error) {

			if req.Method != "HEAD" {
				return httpmock.NewStringResponse(400, "Bad method"), nil
			}

			if req.RequestURI != TestSite {
				return httpmock.NewStringResponse(400, "Bad URL"), nil
			}

			return httpmock.NewStringResponse(200, "To be blocked"), nil
		},
	)

	err = handleSingleFile(ctx, file)
	assert.NoError(t, err, "no error")
	assert.Equal(t, realPaths, ctx.Paths, "should be equal")
	assert.Equal(t, realURLs, ctx.URLs, "should be equal")
}

func TestHandleCSVVerbose(t *testing.T) {
	var testSite string

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
		TestSite: "BLOCKED-EEC",
	}
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	if proxyURL != nil {
		testSite = proxyURL.Host
	} else {
		testSite = TestSite
	}

	// mock to add a new measurement
	httpmock.RegisterResponder("HEAD", testSite,
		func(req *http.Request) (*http.Response, error) {

			if req.Method != "HEAD" {
				return httpmock.NewStringResponse(400, "Bad method"), nil
			}

			if req.RequestURI != TestSite {
				return httpmock.NewStringResponse(400, "Bad URL"), nil
			}

			return httpmock.NewStringResponse(200, "To be blocked"), nil
		},
	)

	err = handleSingleFile(ctx, file)
	assert.NoError(t, err, "no error")
	assert.Equal(t, realPaths, ctx.Paths, "should be equal")
	assert.Equal(t, realURLs, ctx.URLs, "should be equal")
}
