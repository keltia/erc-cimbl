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
	file := "testdata/CIMBL-0666-CERTS.csv"
	ctx := &Context{}

	fn, err := openFile(ctx, file)

	assert.NoError(t, err)
	assert.NotNil(t, fn)
}

func TestOpenZIPFileGood(t *testing.T) {
	file := "testdata/CIMBL-0666-CERTS.zip"

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

func TestOpenZIPFileBad(t *testing.T) {
	file := "testdata/CIMBL-0666-CERTS.zip"

	require.NoError(t, os.Chmod(file, 000))

	snd, err := sandbox.New("test")
	require.NoError(t, err)
	defer snd.Cleanup()

	ctx := &Context{
		tempdir: snd,
	}

	fn, err := openFile(ctx, file)

	assert.Error(t, err)
	assert.Empty(t, fn)

	require.NoError(t, os.Chmod(file, 0644))
}

func TestOpenASCFileBad(t *testing.T) {
	file := "testdata/CIMBL-0666-CERTS.zip.asc"

	require.NoError(t, os.Chmod(file, 000))

	snd, err := sandbox.New("test")
	require.NoError(t, err)
	defer snd.Cleanup()

	ctx := &Context{
		tempdir: snd,
	}

	fn, err := openFile(ctx, file)

	assert.Error(t, err)
	assert.Empty(t, fn)

	require.NoError(t, os.Chmod(file, 0644))
}

func TestReadCSVNone(t *testing.T) {
	snd, err := sandbox.New("test")
	require.NoError(t, err)
	defer snd.Cleanup()

	ctx := &Context{
		tempdir: snd,
	}

	path := readCSV(ctx, nil)
	assert.Empty(t, path)
}

func TestParseCSVNone(t *testing.T) {
	file := "/noneexistent"
	ctx := &Context{}
	err := handleSingleFile(ctx, file)

	assert.Error(t, err)
}

func TestHandleCSV(t *testing.T) {
	defer gock.Off()

	baseDir = "testdata"
	file := "testdata/CIMBL-0666-CERTS.csv"
	config, err := loadConfig()
	assert.NoError(t, err)
	assert.NotNil(t, config)

	ctx := &Context{
		config: config,
		Paths:  map[string]bool{},
		URLs:   map[string]bool{},
	}

	realPaths := map[string]bool{
		"55fe62947f3860108e7798c4498618cb.rtf": true,
	}
	realURLs := map[string]bool{
		TestSite: true,
	}

	_, transport := proxy.SetupTransport(TestSite)
	require.NotNil(t, transport)

	// Set up minimal client
	ctx.Client = &http.Client{Transport: transport, Timeout: 10 * time.Second}

	testSite, err := url.Parse(TestSite)
	require.NoError(t, err)

	gock.New(testSite.Host).
		Head(testSite.Path).
		Reply(200)

	gock.InterceptClient(ctx.Client)
	defer gock.RestoreClient(ctx.Client)

	err = handleSingleFile(ctx, file)
	assert.NoError(t, err)
	assert.Equal(t, realPaths, ctx.Paths)
	assert.Equal(t, realURLs, ctx.URLs)
}

func TestHandleCSVVerbose(t *testing.T) {
	defer gock.Off()

	baseDir = "testdata"
	file := "testdata/CIMBL-0666-CERTS.csv"
	config, err := loadConfig()
	assert.NoError(t, err)

	fVerbose = true

	ctx := &Context{
		config: config,
		Paths:  map[string]bool{},
		URLs:   map[string]bool{},
	}

	realPaths := map[string]bool{
		"55fe62947f3860108e7798c4498618cb.rtf": true,
	}
	realURLs := map[string]bool{
		TestSite: true,
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

	err = handleSingleFile(ctx, file)
	assert.NoError(t, err, "no error")
	assert.Equal(t, realPaths, ctx.Paths, "should be equal")
	assert.Equal(t, realURLs, ctx.URLs, "should be equal")
}

func TestOpenZIPFile(t *testing.T) {
	baseDir = "testdata"
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

	file := "testdata/CIMBL-0666-CERTS.zip"
	fn, err := openZipfile(ctx, file)
	assert.NoError(t, err)
	assert.Equal(t, snd.Cwd()+"/CIMBL-0666-CERTS.csv", fn)
}

func TestOpenZIPFile_None(t *testing.T) {
	baseDir = "testdata"
	config, err := loadConfig()
	assert.NoError(t, err)

	ctx := &Context{
		config: config,
	}

	file := "/nonexistent"
	fn, err := openZipfile(ctx, file)
	assert.Error(t, err)
	assert.Empty(t, fn)
}

func TestHandleSingleFile(t *testing.T) {
	baseDir = "testdata"
	config, err := loadConfig()
	assert.NoError(t, err)

	fVerbose = true

	realPaths := map[string]bool{
		"55fe62947f3860108e7798c4498618cb.rtf": true,
	}
	realURLs := map[string]bool{
		TestSite: true,
	}

	snd, err := sandbox.New("test")
	require.NoError(t, err)
	defer snd.Cleanup()

	ctx := &Context{
		config:  config,
		Paths:   map[string]bool{},
		URLs:    map[string]bool{},
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

	file := "testdata/CIMBL-0666-CERTS.csv"
	err = handleSingleFile(ctx, file)
	assert.NoError(t, err)
	assert.NotEmpty(t, ctx.Paths)
	assert.NotEmpty(t, ctx.URLs)
	assert.Equal(t, realPaths, ctx.Paths, "should be equal")
	assert.Equal(t, realURLs, ctx.URLs, "should be equal")
}

func TestHandleSingleFile_None(t *testing.T) {
	baseDir = "testdata"
	config, err := loadConfig()
	assert.NoError(t, err)

	fVerbose = true

	snd, err := sandbox.New("test")
	require.NoError(t, err)
	defer snd.Cleanup()

	ctx := &Context{
		config:  config,
		Paths:   map[string]bool{},
		URLs:    map[string]bool{},
		tempdir: snd,
	}

	file := "testdata/CIMBL-0667-CERTS.csv"
	err = handleSingleFile(ctx, file)
	assert.Error(t, err)
	assert.Empty(t, ctx.Paths)
	assert.Empty(t, ctx.URLs)
}

func TestHandleAllFiles_None(t *testing.T) {
	ctx := &Context{
		Paths: map[string]bool{},
		URLs:  map[string]bool{},
	}

	err := handleAllFiles(ctx, nil)
	assert.NoError(t, err)
}

func TestHandleAllFiles_Null(t *testing.T) {
	ctx := &Context{
		Paths: map[string]bool{},
		URLs:  map[string]bool{},
	}

	err := handleAllFiles(ctx, []string{"/nonexistent"})
	assert.NoError(t, err)
}

func TestHandleAllFiles_OneFile(t *testing.T) {
	baseDir = "testdata"
	config, err := loadConfig()
	assert.NoError(t, err)

	fVerbose = true

	realPaths := map[string]bool{
		"55fe62947f3860108e7798c4498618cb.rtf": true,
	}

	realURLs := map[string]bool{
		TestSite: true,
	}

	snd, err := sandbox.New("test")
	require.NoError(t, err)
	defer snd.Cleanup()

	ctx := &Context{
		config:  config,
		Paths:   map[string]bool{},
		URLs:    map[string]bool{},
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
		Reply(200)

	gock.InterceptClient(ctx.Client)
	defer gock.RestoreClient(ctx.Client)

	file := "testdata/CIMBL-0666-CERTS.csv"

	err = handleAllFiles(ctx, []string{file})
	assert.NoError(t, err)

	assert.NotEmpty(t, ctx.Paths)
	assert.NotEmpty(t, ctx.URLs)
	assert.Equal(t, realPaths, ctx.Paths)
	assert.Equal(t, realURLs, ctx.URLs)
}

func TestHandleAllFiles_OneURL(t *testing.T) {
	baseDir = "testdata"
	config, err := loadConfig()
	assert.NoError(t, err)

	fVerbose = true

	realURLs := map[string]bool{
		TestSite: true,
	}

	snd, err := sandbox.New("test")
	require.NoError(t, err)
	defer snd.Cleanup()

	ctx := &Context{
		config:  config,
		URLs:    map[string]bool{},
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
		Reply(200)

	gock.InterceptClient(ctx.Client)
	defer gock.RestoreClient(ctx.Client)

	file := TestSite

	err = handleAllFiles(ctx, []string{file})
	assert.NoError(t, err)

	assert.Empty(t, ctx.Paths)
	assert.NotEmpty(t, ctx.URLs)
	assert.Equal(t, realURLs, ctx.URLs)
}
