package main

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"path/filepath"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/h2non/gock"
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
	assert.Equal(t, "zipin", base)
	assert.FileExists(t, filepath.Join(ctx.tempdir.Cwd(), "zipin"))
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

	ctx := &Context{config:  config, tempdir: snd, jobs: 1}

	c := resty.New()
	ctx.Client = c

	fDebug = true
	res, err := handleAllFiles(ctx, []string{"http://localhost/foo.php"})
	t.Logf("res/test=%#v", res)
	assert.NoError(t, err)
	assert.Empty(t, res.URLs)
	fDebug = false
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

	ctx := &Context{config:  config, jobs: 1}

	c := resty.New()

	ctx.Client = c
	testSite, err := url.Parse(TestSite)
	require.NoError(t, err)

	gock.New(testSite.Host).
		Head(testSite.Path).
		MatchHeader("User-Agent", fmt.Sprintf("%s/%s", MyName, MyVersion)).
		Reply(200)

	gock.InterceptClient(c.GetClient())
	defer gock.RestoreClient(c.GetClient())

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

	ctx := &Context{config:  config, jobs: 1}

	c := resty.New()

	ctx.Client = c
	testSite, err := url.Parse(TestSite)
	require.NoError(t, err)

	gock.New(testSite.Host).
		Head(testSite.Path).
		MatchHeader("User-Agent", fmt.Sprintf("%s/%s", MyName, MyVersion)).
		Reply(200)

	gock.InterceptClient(c.GetClient())
	defer gock.RestoreClient(c.GetClient())

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

	fDebug = true

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

	c := resty.New()

	ctx.Client = c
	testSite, err := url.Parse(TestSite)
	require.NoError(t, err)

	gock.New(testSite.Host).
		Head(testSite.Path).
		MatchHeader("User-Agent", fmt.Sprintf("%s/%s", MyName, MyVersion)).
		Reply(200)

	gock.InterceptClient(c.GetClient())
	defer gock.RestoreClient(c.GetClient())

	file := TestSite

	res, err := handleAllFiles(ctx, []string{file})
	assert.NoError(t, err)

	assert.Empty(t, res.Paths)
	assert.NotEmpty(t, res.URLs)
	assert.EqualValues(t, realURLs, res.URLs)

	fDebug = false
}

func TestRemoveExt(t *testing.T) {
	td := []struct{in, out string}{
		{"", ""},
		{"foobar", "foobar"},
		{"foobar.js", "foobar"},
		{"foobar.zip.asc", "foobar.zip"},
	}

	for _, d := range td {
		assert.Equal(t, d.out, RemoveExt(d.in))
	}
}

func BenchmarkRemoveExt(b *testing.B) {
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_ = RemoveExt("foobar.zip.asc")
	}
}
