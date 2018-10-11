package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path/filepath"
	"testing"
	"time"

	"github.com/h2non/gock"
	"github.com/keltia/proxy"
	"github.com/keltia/sandbox"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseCSVNone(t *testing.T) {
	file := "/noneexistent"
	ctx := &Context{}
	_, err := handleSingleFile(ctx, file)

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
	}

	fDebug = true
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

	res, err := handleSingleFile(ctx, file)
	assert.NoError(t, err)
	assert.Equal(t, realPaths, res.Paths)
	assert.Equal(t, realURLs, res.URLs)
	fDebug = false
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

	res, err := handleSingleFile(ctx, file)
	assert.NoError(t, err, "no error")
	assert.Equal(t, realPaths, res.Paths)
	assert.Equal(t, realURLs, res.URLs)
}

func TestExtractZipFrom_None(t *testing.T) {
	file := "/noneexistent"
	base, err := extractZipFrom(file)
	assert.Error(t, err)
	assert.Empty(t, base)
}

func TestExtractZipFrom(t *testing.T) {
	file := "/noneexistent"
	base, err := extractZipFrom(file)
	assert.Error(t, err)
	assert.Empty(t, base)
}

func TestReadFile_None(t *testing.T) {
	file := "nonexistent.txt"
	buf, err := readFile(file)
	assert.Error(t, err)
	assert.Empty(t, buf)
}

func TestReadFile_NoCSV(t *testing.T) {
	file := "testdata/CIMBL-0668-CERTS.zip"
	buf, err := readFile(file)
	assert.Error(t, err)
	assert.Empty(t, buf)
}

func TestReadFile_Good(t *testing.T) {
	file := "testdata/CIMBL-0666-CERTS.zip"
	buf, err := readFile(file)
	assert.NoError(t, err)
	assert.NotEmpty(t, buf)

	b, err := ioutil.ReadFile("testdata/CIMBL-0666-CERTS.csv")
	require.NoError(t, err)
	require.NotEmpty(t, b)

	assert.Equal(t, string(b), buf.String())
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
	res, err := handleSingleFile(ctx, file)
	assert.NoError(t, err)
	assert.NotEmpty(t, res.Paths)
	assert.NotEmpty(t, res.URLs)
	assert.Equal(t, realPaths, res.Paths)
	assert.Equal(t, realURLs, res.URLs)
}

func TestHandleSingleFile_Transport(t *testing.T) {
	baseDir = "testdata"
	config, err := loadConfig()
	assert.NoError(t, err)

	fVerbose = true

	realPaths := map[string]bool{
		"55fe62947f3860108e7798c4498618cb.rtf": true,
	}
	realURLs := map[string]bool{}

	snd, err := sandbox.New("test")
	require.NoError(t, err)
	defer snd.Cleanup()

	ctx := &Context{
		config:  config,
		tempdir: snd,
	}

	// Set up minimal client
	ctx.Client = &http.Client{Transport: nil, Timeout: 10 * time.Second}

	file := "testdata/CIMBL-0666-CERTS.csv"
	res, err := handleSingleFile(ctx, file)
	assert.NoError(t, err)
	assert.NotEmpty(t, res.Paths)
	assert.Empty(t, res.URLs)
	assert.EqualValues(t, realPaths, res.Paths)
	assert.EqualValues(t, realURLs, res.URLs)
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
		tempdir: snd,
	}

	file, _ := filepath.Abs("testdata/CIMBL-0667-CERTS.csv")
	res, err := handleSingleFile(ctx, file)
	assert.Error(t, err)
	require.NotNil(t, res)
	assert.Empty(t, res.Paths)
	assert.Empty(t, res.URLs)
}

func TestHandleAllFiles_None(t *testing.T) {
	ctx := &Context{}

	res, err := handleAllFiles(ctx, nil)
	assert.NoError(t, err)
	assert.Empty(t, res.URLs)
}

func TestHandleAllFiles_Null(t *testing.T) {
	ctx := &Context{}

	res, err := handleAllFiles(ctx, []string{"/nonexistent"})
	assert.NoError(t, err)
	assert.Empty(t, res.URLs)
}

func TestHandleAllFiles_SingleBad(t *testing.T) {
	snd, err := sandbox.New("test")
	require.NoError(t, err)
	defer snd.Cleanup()

	ctx := &Context{
		tempdir: snd,
	}

	res, err := handleAllFiles(ctx, []string{"testdata/CIMBL-0667-CERTS.csv"})
	assert.NoError(t, err)
	assert.Empty(t, res.URLs)
}

func TestHandleAllFiles_SingleBad2(t *testing.T) {
	snd, err := sandbox.New("test")
	require.NoError(t, err)
	defer snd.Cleanup()

	ctx := &Context{
		tempdir: snd,
	}

	res, err := handleAllFiles(ctx, []string{"http://localhost/foo.php"})
	assert.NoError(t, err)
	assert.Empty(t, res.URLs)
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

	res, err := handleAllFiles(ctx, []string{file})
	assert.NoError(t, err)

	assert.NotEmpty(t, res.Paths)
	assert.NotEmpty(t, res.URLs)
	assert.EqualValues(t, realPaths, res.Paths)
	assert.EqualValues(t, realURLs, res.URLs)
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

	res, err := handleAllFiles(ctx, []string{file})
	assert.NoError(t, err)

	assert.Empty(t, res.Paths)
	assert.NotEmpty(t, res.URLs)
	assert.EqualValues(t, realURLs, res.URLs)
}
