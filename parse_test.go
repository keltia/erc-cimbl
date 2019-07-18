package main

import (
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

func TestExtractZipEmpty(t *testing.T) {
	file := "testdata/empty.zip"
	base, err := extractZipFrom(file)
	assert.Error(t, err)
	assert.Empty(t, base)
}

func TestExtractZipOne(t *testing.T) {
	file := "testdata/one.zip"
	base, err := extractZipFrom(file)
	assert.Error(t, err)
	assert.Empty(t, base)
}

func TestExtractZipZipin(t *testing.T) {
	baseDir = "testdata"
	config, err := loadConfig()
	assert.NoError(t, err)

	snd, err := sandbox.New("test")
	require.NoError(t, err)
	defer snd.Cleanup()

	ctx := &Context{
		config:  config,
		tempdir: snd,
		jobs:    1,
	}

	var base string

	file, err := filepath.Abs("testdata/zipin.zip")
	require.NoError(t, err)

	err = snd.Run(func() error {
		var err error

		base, err = extractZipFrom(file)
		assert.NotEmpty(t, base)
		return err
	})

	assert.NoError(t, err)
	assert.NotEmpty(t, base)
	assert.Equal(t, "zipin.zip.zip", base)
	assert.FileExists(t, filepath.Join(ctx.tempdir.Cwd(), "zipin.zip.zip"))
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

func TestHandleAllFiles_None(t *testing.T) {
	baseDir = "testdata"
	config, err := loadConfig()
	assert.NoError(t, err)
	require.NotEmpty(t, config.REFile)

	snd, err := sandbox.New("test")
	require.NoError(t, err)
	defer snd.Cleanup()

	ctx := &Context{
		config:  config,
		tempdir: snd,
		jobs:    1,
	}

	res, err := handleAllFiles(ctx, nil)
	assert.NoError(t, err)
	assert.Empty(t, res.URLs)
}

func TestHandleAllFiles_Null(t *testing.T) {
	baseDir = "testdata"
	config, err := loadConfig()
	assert.NoError(t, err)
	require.NotEmpty(t, config.REFile)

	snd, err := sandbox.New("test")
	require.NoError(t, err)
	defer snd.Cleanup()

	ctx := &Context{
		config:  config,
		tempdir: snd,
		jobs:    1,
	}

	res, err := handleAllFiles(ctx, []string{"/nonexistent"})
	assert.NoError(t, err)
	assert.Empty(t, res.URLs)
}

func TestHandleAllFiles_SingleBad(t *testing.T) {
	baseDir = "testdata"
	config, err := loadConfig()
	assert.NoError(t, err)
	require.NotEmpty(t, config.REFile)

	snd, err := sandbox.New("test")
	require.NoError(t, err)
	defer snd.Cleanup()

	ctx := &Context{
		config:  config,
		tempdir: snd,
		jobs:    1,
	}

	res, err := handleAllFiles(ctx, []string{"testdata/CIMBL-0667-CERTS.csv"})
	assert.NoError(t, err)
	assert.Empty(t, res.URLs)
}

func TestHandleAllFiles_SingleBad2(t *testing.T) {
	baseDir = "testdata"
	config, err := loadConfig()
	assert.NoError(t, err)
	require.NotEmpty(t, config.REFile)

	snd, err := sandbox.New("test")
	require.NoError(t, err)
	defer snd.Cleanup()

	ctx := &Context{
		config:  config,
		tempdir: snd,
		jobs:    1,
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
		jobs:    1,
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

func TestHandleAllFiles_OneFile1(t *testing.T) {
	baseDir = "testdata"
	config, err := loadConfig()
	assert.NoError(t, err)

	fDebug = true

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
		jobs:    1,
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

	fDebug = false
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
		jobs:    1,
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

	fVerbose = false
}
