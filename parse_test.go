package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"github.com/jarcoal/httpmock"
	"net/http"
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
    var testSite string

	file := "test/CIMBL-0666-CERTS.csv"
	config, err := loadConfig()
	assert.NoError(t, err, "no error")

	ctx := &Context{
		config: config,
		Paths:  map[string]bool{},
		URLs:   map[string]string{},
	}

    err = setupProxyAuth(ctx, dbrcFile)
    if err != nil {
        t.Log("No dbrc file, no proxy auth.")
    }

    realPaths := map[string]bool{
		"55fe62947f3860108e7798c4498618cb.rtf": true,
	}
	realURLs := map[string]string{
		TestSite: "**BLOCK**",
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

	err = handleCSV(ctx, file)
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

    err = setupProxyAuth(ctx, dbrcFile)
    if err != nil {
        t.Log("No dbrc file, no proxy auth.")
    }

	realPaths := map[string]bool{
		"55fe62947f3860108e7798c4498618cb.rtf": true,
	}
	realURLs := map[string]string{
		TestSite: "**BLOCK**",
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

	err = handleCSV(ctx, file)
	assert.NoError(t, err, "no error")
	assert.Equal(t, realPaths, ctx.Paths, "should be equal")
	assert.Equal(t, realURLs, ctx.URLs, "should be equal")
}
