package main

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/h2non/gock"
	"github.com/keltia/proxy"
	"github.com/keltia/sandbox"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenFileBad(t *testing.T) {
	file := "foo.bar"
	ctx := &Context{}
	fn, err := openFile(ctx, file)

	assert.Empty(t, fn)
	assert.Error(t, err)
	assert.Nil(t, fn)
}

func TestOpenFileGood(t *testing.T) {
	file := "test/CIMBL-0666-CERTS.csv"
	ctx := &Context{}

	fn, err := openFile(ctx, file)

	assert.NoError(t, err)
	assert.NotNil(t, fn)
}

func TestOpenZIPFileGood(t *testing.T) {
	file := "test/CIMBL-0666-CERTS.zip"

	snd, err := sandbox.New("test")
	require.NoError(t, err)
	defer snd.Cleanup()

	ctx := &Context{
		tempdir: snd,
	}

	fn, err := openFile(ctx, file)

	assert.NoError(t, err)
	assert.NotNil(t, fn)
	assert.IsType(t, (*os.File)(nil), fn)
}

func TestParseCSVNone(t *testing.T) {
	file := "/noneexistent"
	ctx := &Context{}
	err := handleSingleFile(ctx, file)

	assert.Error(t, err)
}

func TestHandleCSV(t *testing.T) {
	defer gock.Off()

	baseDir = "test"
	file := "test/CIMBL-0666-CERTS.csv"
	config, err := loadConfig()
	assert.NoError(t, err)
	assert.NotNil(t, config)

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

	_, transport := proxy.SetupTransport(TestSite)
	require.NotNil(t, transport)

	// Set up minimal client
	ctx.Client = &http.Client{Transport: transport, Timeout: 10 * time.Second}

	testSite, err := url.Parse(TestSite)
	require.NoError(t, err)

	gock.New(testSite.Host).
		Head(testSite.Path).
		Reply(403)

	gock.InterceptClient(ctx.Client)
	defer gock.RestoreClient(ctx.Client)

	err = handleSingleFile(ctx, file)
	assert.NoError(t, err)
	assert.Equal(t, realPaths, ctx.Paths)
	assert.Equal(t, realURLs, ctx.URLs)
}

func TestHandleCSVVerbose(t *testing.T) {
	defer gock.Off()

	baseDir = "test"
	file := "test/CIMBL-0666-CERTS.csv"
	config, err := loadConfig()
	assert.NoError(t, err)

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

	_, transport := proxy.SetupTransport(TestSite)
	require.NotNil(t, transport)

	// Set up minimal client
	ctx.Client = &http.Client{Transport: transport, Timeout: 10 * time.Second}

	testSite, err := url.Parse(TestSite)
	require.NoError(t, err)

	gock.New(testSite.Host).
		Head(testSite.Path).
		MatchHeader("user-agent", fmt.Sprintf("%s/%s", MyName, MyVersion)).
		Reply(403)

	gock.InterceptClient(ctx.Client)
	defer gock.RestoreClient(ctx.Client)

	err = handleSingleFile(ctx, file)
	assert.NoError(t, err, "no error")
	assert.Equal(t, realPaths, ctx.Paths, "should be equal")
	assert.Equal(t, realURLs, ctx.URLs, "should be equal")
}

func TestOpenZIPFile(t *testing.T) {
	baseDir = "test"
	config, err := loadConfig()
	assert.NoError(t, err)

	fVerbose = true

	snd, err := sandbox.New("test")
	require.NoError(t, err)
	defer snd.Cleanup()

	ctx := &Context{
		config:  config,
		tempdir: snd,
	}

	file := "test/CIMBL-0666-CERTS.zip"
	fn := openZipfile(ctx, file)
	assert.Equal(t, snd.Cwd()+"/CIMBL-0666-CERTS.csv", fn)
}

func TestHandleSingleFile(t *testing.T) {
	baseDir = "test"
	config, err := loadConfig()
	assert.NoError(t, err)

	fVerbose = true

	realPaths := map[string]bool{
		"55fe62947f3860108e7798c4498618cb.rtf": true,
	}
	realURLs := map[string]string{
		TestSite: ActionBlock,
	}

	snd, err := sandbox.New("test")
	require.NoError(t, err)
	defer snd.Cleanup()

	ctx := &Context{
		config:  config,
		Paths:   map[string]bool{},
		URLs:    map[string]string{},
		tempdir: snd,
	}

	_, transport := proxy.SetupTransport(TestSite)
	require.NotNil(t, transport)

	// Set up minimal client
	ctx.Client = &http.Client{Transport: transport, Timeout: 10 * time.Second}

	testSite, err := url.Parse(TestSite)
	require.NoError(t, err)

	gock.New(testSite.Host).
		Head(testSite.Path).
		MatchHeader("user-agent", fmt.Sprintf("%s/%s", MyName, MyVersion)).
		Reply(200)

	gock.InterceptClient(ctx.Client)
	defer gock.RestoreClient(ctx.Client)

	file := "test/CIMBL-0666-CERTS.csv"
	err = handleSingleFile(ctx, file)
	assert.NoError(t, err)
	assert.NotEmpty(t, ctx.Paths)
	assert.NotEmpty(t, ctx.URLs)
	assert.Equal(t, realPaths, ctx.Paths, "should be equal")
	assert.Equal(t, realURLs, ctx.URLs, "should be equal")
}
