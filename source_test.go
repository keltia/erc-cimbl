package main

import (
	"fmt"
	"net/url"
	"os"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/h2non/gock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Filename

func TestNewFilename(t *testing.T) {
	fn := NewFilename("")
	assert.IsType(t, (*Filename)(nil), fn)
	assert.Empty(t, fn.Name)
}

func TestNewFilename2(t *testing.T) {
	fn := NewFilename("foo")
	assert.IsType(t, (*Filename)(nil), fn)
	assert.Equal(t, "foo", fn.Name)
}

func TestFilename_AddTo(t *testing.T) {
	td := map[string]bool{"example.docx": true}

	fn := NewFilename("example.docx")

	r := NewResults()
	fn.AddTo(r)
	assert.NotEmpty(t, r.Paths)
	assert.Equal(t, td, r.Paths)
}

func TestFilename_Check(t *testing.T) {
	proxy := os.Getenv("http_proxy")
	c := resty.New().SetProxy(proxy)

	fn := NewFilename("example.docx")
	assert.True(t, fn.Check(c))
}

// URL

func TestNewURL(t *testing.T) {
	u := NewURL("")
	assert.IsType(t, (*URL)(nil), u)
	assert.Empty(t, u.H)
}

func TestURL_AddTo(t *testing.T) {
	td := map[string]bool{"http://www.example.net/": true}

	u := NewURL("http://www.example.net/")
	r := NewResults()
	u.AddTo(r)
	assert.NotEmpty(t, r.URLs)
	assert.Equal(t, td, r.URLs)
}

func TestList_Check(t *testing.T) {
	defer gock.Off()

	baseDir = "testdata"
	config, err := loadConfig()
	assert.NoError(t, err)
	assert.NotNil(t, config)

	l := NewList([]string{})
	require.Empty(t, l)

	proxy := os.Getenv("http_proxy")
	c := resty.New().SetProxy(proxy)

	testSite, err := url.Parse(TestSite)
	require.NoError(t, err)

	gock.New(testSite.Host).
		Head(testSite.Path).
		Reply(200)

	gock.InterceptClient(c.GetClient())
	defer gock.RestoreClient(c.GetClient())

	u := NewURL(TestSite)
	assert.True(t, u.Check(c))
}

func TestList_Check2(t *testing.T) {
	defer gock.Off()

	baseDir = "testdata"
	file := "testdata/CIMBL-0669-CERTS.csv"
	config, err := loadConfig()
	assert.NoError(t, err)
	assert.NotNil(t, config)

	ctx := &Context{
		config: config,
		jobs:   1,
	}

	l := NewList([]string{file})
	require.NotEmpty(t, l)

	realPaths := map[string]bool{
		"55fe62947f3860108e7798c4498618cb.rtf": true,
	}
	realURLs := map[string]bool{
		TestSite: true,
	}

	proxy := os.Getenv("http_proxy")
	c := resty.New().SetProxy(proxy)

	ctx.Client = c
	testSite, err := url.Parse(TestSite)
	require.NoError(t, err)

	gock.New(testSite.Host).
		Head(testSite.Path).
		Reply(200)

	gock.InterceptClient(c.GetClient())
	defer gock.RestoreClient(c.GetClient())

	res := l.Check(ctx)
	t.Logf("res=%#v", res)
	assert.NoError(t, err)
	assert.EqualValues(t, realPaths, res.Paths)
	assert.EqualValues(t, realURLs, res.URLs)
}

func TestList_Check3(t *testing.T) {
	defer gock.Off()

	baseDir = "testdata"
	file := "testdata/CIMBL-0666-CERTS.csv"
	config, err := loadConfig()
	assert.NoError(t, err)

	fDebug = true

	ctx := &Context{
		config: config,
		jobs:   1,
	}

	l := NewList([]string{file})

	require.NotEmpty(t, l)

	realPaths := map[string]bool{
		"55fe62947f3860108e7798c4498618cb.rtf": true,
	}
	realURLs := map[string]bool{
		TestSite: true,
	}

	proxy := os.Getenv("http_proxy")
	c := resty.New().SetProxy(proxy)

	ctx.Client = c
	testSite, err := url.Parse(TestSite)
	require.NoError(t, err)

	gock.New(testSite.Host).
		Head(testSite.Path).
		MatchHeader("user-agent", fmt.Sprintf("%s/%s", MyName, MyVersion)).
		Reply(200)

	gock.InterceptClient(c.GetClient())
	defer gock.RestoreClient(c.GetClient())

	res := l.Check(ctx)
	t.Logf("res=%#v", res)
	assert.NoError(t, err, "no error")
	assert.EqualValues(t, realPaths, res.Paths)
	assert.EqualValues(t, realURLs, res.URLs)
	fDebug = false
}

func TestList_Check41(t *testing.T) {
	defer gock.Off()

	baseDir = "testdata"
	file := "testdata/CIMBL-0669-CERTS.csv"
	config, err := loadConfig()
	assert.NoError(t, err)

	ctx := &Context{
		config: config,
		jobs:   1,
	}

	l := NewList([]string{file})
	require.NotEmpty(t, l)

	t.Logf("site=%#v", TestSite)
	realPaths := map[string]bool{
		"55fe62947f3860108e7798c4498618cb.rtf": true,
	}
	realURLs := map[string]bool{
		TestSite: true,
	}

	proxy := os.Getenv("http_proxy")
	c := resty.New().SetProxy(proxy)

	ctx.Client = c
	testSite, err := url.Parse(TestSite)
	require.NoError(t, err)

	gock.New(testSite.Host).
		Head(testSite.Path).
		Reply(200)

	gock.InterceptClient(c.GetClient())
	defer gock.RestoreClient(c.GetClient())

	res := l.Check(ctx)
	assert.NoError(t, err)
	assert.EqualValues(t, realPaths, res.Paths)
	assert.EqualValues(t, realURLs, res.URLs)
}

func TestList_Check43(t *testing.T) {
	defer gock.Off()

	baseDir = "testdata"
	file := "testdata/CIMBL-0669-CERTS.csv"
	config, err := loadConfig()
	assert.NoError(t, err)

	ctx := &Context{
		config: config,
		jobs:   3,
	}

	l := NewList([]string{file})
	require.NotEmpty(t, l)

	realPaths := map[string]bool{
		"55fe62947f3860108e7798c4498618cb.rtf": true,
	}
	realURLs := map[string]bool{
		TestSite: true,
	}

	proxy := os.Getenv("http_proxy")
	c := resty.New().SetProxy(proxy)

	ctx.Client = c
	testSite, err := url.Parse(TestSite)
	require.NoError(t, err)

	gock.New(testSite.Host).
		Head(testSite.Path).
		Reply(200)

	gock.InterceptClient(c.GetClient())
	defer gock.RestoreClient(c.GetClient())

	res := l.Check(ctx)
	assert.NoError(t, err)
	assert.EqualValues(t, realPaths, res.Paths)
	assert.EqualValues(t, realURLs, res.URLs)
}

func TestList_Check44(t *testing.T) {
	defer gock.Off()

	baseDir = "testdata"
	file := "testdata/CIMBL-0669-CERTS.csv"
	config, err := loadConfig()
	assert.NoError(t, err)

	ctx := &Context{
		config: config,
		jobs:   4,
	}

	l := NewList([]string{file})
	require.NotEmpty(t, l)

	realPaths := map[string]bool{
		"55fe62947f3860108e7798c4498618cb.rtf": true,
	}
	realURLs := map[string]bool{
		TestSite: true,
	}

	proxy := os.Getenv("http_proxy")
	c := resty.New().SetProxy(proxy)

	ctx.Client = c
	testSite, err := url.Parse(TestSite)
	require.NoError(t, err)

	gock.New(testSite.Host).
		Head(testSite.Path).
		Reply(200)

	gock.InterceptClient(c.GetClient())
	defer gock.RestoreClient(c.GetClient())

	res := l.Check(ctx)
	assert.NoError(t, err)
	assert.EqualValues(t, realPaths, res.Paths)
	assert.EqualValues(t, realURLs, res.URLs)
}

func TestList_Check48(t *testing.T) {
	defer gock.Off()

	baseDir = "testdata"
	file := "testdata/CIMBL-0669-CERTS.csv"
	config, err := loadConfig()
	assert.NoError(t, err)

	ctx := &Context{
		config: config,
		jobs:   8,
	}

	l := NewList([]string{file})
	require.NotEmpty(t, l)

	realPaths := map[string]bool{
		"55fe62947f3860108e7798c4498618cb.rtf": true,
	}
	realURLs := map[string]bool{
		TestSite: true,
	}

	proxy := os.Getenv("http_proxy")
	c := resty.New().SetProxy(proxy)

	ctx.Client = c
	testSite, err := url.Parse(TestSite)
	require.NoError(t, err)

	gock.New(testSite.Host).
		Head(testSite.Path).
		Reply(200)

	gock.InterceptClient(c.GetClient())
	defer gock.RestoreClient(c.GetClient())

	res := l.Check(ctx)
	assert.NoError(t, err)
	assert.EqualValues(t, realPaths, res.Paths)
	assert.EqualValues(t, realURLs, res.URLs)
}

// Variation with Check1

func TestList_Check1(t *testing.T) {
	defer gock.Off()

	baseDir = "testdata"
	config, err := loadConfig()
	assert.NoError(t, err)
	assert.NotNil(t, config)

	l := NewList([]string{})
	require.Empty(t, l)

	proxy := os.Getenv("http_proxy")
	c := resty.New().SetProxy(proxy)

	testSite, err := url.Parse(TestSite)
	require.NoError(t, err)

	gock.New(testSite.Host).
		Head(testSite.Path).
		Reply(200)

	gock.InterceptClient(c.GetClient())
	defer gock.RestoreClient(c.GetClient())

	u := NewURL(TestSite)
	assert.True(t, u.Check(c))
}

func TestList_Check12(t *testing.T) {
	defer gock.Off()

	baseDir = "testdata"
	file := "testdata/CIMBL-0669-CERTS.csv"
	config, err := loadConfig()
	assert.NoError(t, err)
	assert.NotNil(t, config)

	ctx := &Context{
		config: config,
		jobs:   1,
	}

	l := NewList([]string{file})
	require.NotEmpty(t, l)

	realPaths := map[string]bool{
		"55fe62947f3860108e7798c4498618cb.rtf": true,
	}
	realURLs := map[string]bool{
		TestSite: true,
	}

	proxy := os.Getenv("http_proxy")
	c := resty.New().SetProxy(proxy)

	ctx.Client = c
	testSite, err := url.Parse(TestSite)
	require.NoError(t, err)

	gock.New(testSite.Host).
		Head(testSite.Path).
		Reply(200)

	gock.InterceptClient(c.GetClient())
	defer gock.RestoreClient(c.GetClient())

	res := l.Check1(ctx)
	t.Logf("res=%#v", res)
	assert.NoError(t, err)
	assert.EqualValues(t, realPaths, res.Paths)
	assert.EqualValues(t, realURLs, res.URLs)
}

func TestList_Check13(t *testing.T) {
	defer gock.Off()

	baseDir = "testdata"
	file := "testdata/CIMBL-0666-CERTS.csv"
	config, err := loadConfig()
	assert.NoError(t, err)

	fDebug = true

	ctx := &Context{
		config: config,
		jobs:   1,
	}

	l := NewList([]string{file})

	require.NotEmpty(t, l)

	realPaths := map[string]bool{
		"55fe62947f3860108e7798c4498618cb.rtf": true,
	}
	realURLs := map[string]bool{
		TestSite: true,
	}

	proxy := os.Getenv("http_proxy")
	c := resty.New().SetProxy(proxy)

	ctx.Client = c
	testSite, err := url.Parse(TestSite)
	require.NoError(t, err)

	gock.New(testSite.Host).
		Head(testSite.Path).
		MatchHeader("user-agent", fmt.Sprintf("%s/%s", MyName, MyVersion)).
		Reply(200)

	gock.InterceptClient(c.GetClient())
	defer gock.RestoreClient(c.GetClient())

	res := l.Check1(ctx)
	t.Logf("res=%#v", res)
	assert.NoError(t, err, "no error")
	assert.EqualValues(t, realPaths, res.Paths)
	assert.EqualValues(t, realURLs, res.URLs)
	fDebug = false
}

func TestList_Check141(t *testing.T) {
	defer gock.Off()

	baseDir = "testdata"
	file := "testdata/CIMBL-0669-CERTS.csv"
	config, err := loadConfig()
	assert.NoError(t, err)

	ctx := &Context{
		config: config,
		jobs:   1,
	}

	l := NewList([]string{file})
	require.NotEmpty(t, l)

	t.Logf("site=%#v", TestSite)
	realPaths := map[string]bool{
		"55fe62947f3860108e7798c4498618cb.rtf": true,
	}
	realURLs := map[string]bool{
		TestSite: true,
	}

	proxy := os.Getenv("http_proxy")
	c := resty.New().SetProxy(proxy)

	ctx.Client = c
	testSite, err := url.Parse(TestSite)
	require.NoError(t, err)

	gock.New(testSite.Host).
		Head(testSite.Path).
		Reply(200)

	gock.InterceptClient(c.GetClient())
	defer gock.RestoreClient(c.GetClient())

	res := l.Check1(ctx)
	assert.NoError(t, err)
	assert.EqualValues(t, realPaths, res.Paths)
	assert.EqualValues(t, realURLs, res.URLs)
}

func TestList_Check143(t *testing.T) {
	defer gock.Off()

	baseDir = "testdata"
	file := "testdata/CIMBL-0669-CERTS.csv"
	config, err := loadConfig()
	assert.NoError(t, err)

	ctx := &Context{
		config: config,
		jobs:   3,
	}

	l := NewList([]string{file})
	require.NotEmpty(t, l)

	realPaths := map[string]bool{
		"55fe62947f3860108e7798c4498618cb.rtf": true,
	}
	realURLs := map[string]bool{
		TestSite: true,
	}

	proxy := os.Getenv("http_proxy")
	c := resty.New().SetProxy(proxy)

	ctx.Client = c
	testSite, err := url.Parse(TestSite)
	require.NoError(t, err)

	gock.New(testSite.Host).
		Head(testSite.Path).
		Reply(200)

	gock.InterceptClient(c.GetClient())
	defer gock.RestoreClient(c.GetClient())

	res := l.Check1(ctx)
	assert.NoError(t, err)
	assert.EqualValues(t, realPaths, res.Paths)
	assert.EqualValues(t, realURLs, res.URLs)
}

func TestList_Check144(t *testing.T) {
	defer gock.Off()

	baseDir = "testdata"
	file := "testdata/CIMBL-0669-CERTS.csv"
	config, err := loadConfig()
	assert.NoError(t, err)

	ctx := &Context{
		config: config,
		jobs:   4,
	}

	l := NewList([]string{file})
	require.NotEmpty(t, l)

	realPaths := map[string]bool{
		"55fe62947f3860108e7798c4498618cb.rtf": true,
	}
	realURLs := map[string]bool{
		TestSite: true,
	}

	proxy := os.Getenv("http_proxy")
	c := resty.New().SetProxy(proxy)

	ctx.Client = c
	testSite, err := url.Parse(TestSite)
	require.NoError(t, err)

	gock.New(testSite.Host).
		Head(testSite.Path).
		Reply(200)

	gock.InterceptClient(c.GetClient())
	defer gock.RestoreClient(c.GetClient())

	res := l.Check1(ctx)
	assert.NoError(t, err)
	assert.EqualValues(t, realPaths, res.Paths)
	assert.EqualValues(t, realURLs, res.URLs)
}

func TestList_Check148(t *testing.T) {
	defer gock.Off()

	baseDir = "testdata"
	file := "testdata/CIMBL-0669-CERTS.csv"
	config, err := loadConfig()
	assert.NoError(t, err)

	ctx := &Context{
		config: config,
		jobs:   8,
	}

	l := NewList([]string{file})
	require.NotEmpty(t, l)

	realPaths := map[string]bool{
		"55fe62947f3860108e7798c4498618cb.rtf": true,
	}
	realURLs := map[string]bool{
		TestSite: true,
	}

	proxy := os.Getenv("http_proxy")
	c := resty.New().SetProxy(proxy)

	ctx.Client = c
	testSite, err := url.Parse(TestSite)
	require.NoError(t, err)

	gock.New(testSite.Host).
		Head(testSite.Path).
		Reply(200)

	gock.InterceptClient(c.GetClient())
	defer gock.RestoreClient(c.GetClient())

	res := l.Check1(ctx)
	assert.NoError(t, err)
	assert.EqualValues(t, realPaths, res.Paths)
	assert.EqualValues(t, realURLs, res.URLs)
}

// List

func TestNewList(t *testing.T) {
	l := NewList(nil)
	assert.Empty(t, l)
}

func TestNewList2(t *testing.T) {
	l := NewList([]string{})
	assert.Empty(t, l)
}

func TestNewList3(t *testing.T) {
	td := NewURL("http://www.example.net/")

	l := NewList([]string{"http://www.example.net/"})
	require.NotEmpty(t, l)
	assert.Equal(t, td, l.s[0])
}

func TestNewList4(t *testing.T) {
	td := []Sourcer{
		NewFilename("55fe62947f3860108e7798c4498618cb.rtf"),
		NewURL(TestSite),
		NewURL("http://www.example.net/"),
	}

	l := NewList([]string{"testdata/CIMBL-0666-CERTS.csv", "exemple.docx", "http://www.example.net/"})
	require.NotEmpty(t, l)
	assert.EqualValues(t, td, l.s)
}

func TestNewList_IP(t *testing.T) {
	td := []Sourcer{
		NewURL("http://10.1.1.1/"),
		NewURL("http://172.16.1.1/"),
		NewURL("http://192.168.1.1/"),
	}

	l := NewList([]string{"testdata/iplist.txt"})
	require.NotEmpty(t, l)
	assert.EqualValues(t, td, l.s)
}

func TestNewList_IPBad(t *testing.T) {
	l := NewList([]string{"testdata/nonexistent.txt"})
	assert.Empty(t, l)
}

func TestList_Add(t *testing.T) {
	td := []Sourcer{NewFilename("exemple.docx")}
	l := NewList(nil)
	l.Add(NewFilename("exemple.docx"))
	assert.NotEmpty(t, l.s)
	assert.Equal(t, td, l.s)
}

func TestList_Add2(t *testing.T) {
	td := []Sourcer{
		NewFilename("exemple.docx"),
		NewURL("http://www.example.net/"),
	}

	l := NewList(nil)
	l.Add(NewFilename("exemple.docx"))
	l.Add(NewURL("http://www.example.net/"))
	assert.NotEmpty(t, l.s)
	assert.Equal(t, td, l.s)
}

func TestList_AddFromIP_None(t *testing.T) {
	fn := "nonexistent"
	l := NewList(nil)
	l1, err := l.AddFromIP(fn)
	require.Error(t, err)
	assert.Empty(t, l1)
}

func TestList_AddFromIP_Good(t *testing.T) {
	td := []Sourcer{
		NewURL("http://10.1.1.1/"),
		NewURL("http://172.16.1.1/"),
		NewURL("http://192.168.1.1/"),
	}

	fn := "testdata/iplist.txt"
	l := NewList(nil)
	l1, err := l.AddFromIP(fn)
	require.NoError(t, err)
	assert.NotEmpty(t, l1)
	assert.EqualValues(t, td, l1.s)
}

func TestList_AddFromFile(t *testing.T) {
	td := []Sourcer{
		NewFilename("55fe62947f3860108e7798c4498618cb.rtf"),
		NewURL(TestSite),
	}

	l := NewList(nil)
	l1, err := l.AddFromFile("testdata/CIMBL-0666-CERTS.csv")
	require.NoError(t, err)
	require.NotEmpty(t, l)
	assert.EqualValues(t, td, l.s)
	assert.EqualValues(t, l1, l)
}

func TestList_AddFromFile_None(t *testing.T) {
	l := NewList(nil)
	l1, err := l.AddFromFile("/nonexistent")
	require.Error(t, err)
	require.Empty(t, l)
	assert.EqualValues(t, l1, l)
}

func TestList_AddFromFile_Perms(t *testing.T) {
	file := "testdata/CIMBL-0666-CERTS.csv"

	l := NewList(nil)

	assert.NoError(t, os.Chmod(file, 0000), "should be fine")

	l1, err := l.AddFromFile(file)

	require.Error(t, err)
	require.Empty(t, l)
	assert.EqualValues(t, l1, l)

	assert.NoError(t, os.Chmod(file, 0644), "should be fine")
}

func TestList_AddFromFile2(t *testing.T) {
	file := "testdata/CIMBL-0666-CERTS.zip"

	l := NewList(nil)

	l1, err := l.AddFromFile(file)
	assert.NotEmpty(t, l1)

	assert.NoError(t, err)
	assert.NotEmpty(t, l)
}

func TestList_AddFromFile_Gpg(t *testing.T) {
	file := "testdata/CIMBL-0666-CERTS.zip.asc"
	l := NewList(nil)

	l1, err := l.AddFromFile(file)
	assert.Empty(t, l1)
	assert.Error(t, err)

	assert.Empty(t, l)
}

func TestList_AddFromFile_Badcsv(t *testing.T) {

	l := NewList(nil)
	l1, err := l.AddFromFile("testdata/bad.csv")
	require.Error(t, err)
	assert.NotEmpty(t, l)
	assert.NotEmpty(t, l1.files)
	assert.EqualValues(t, []string{"bad.csv"}, l1.files)
}

func TestList_Merge(t *testing.T) {
	td2 := []string{"http://example.net/"}

	tdm := []Sourcer{
		NewFilename("55fe62947f3860108e7798c4498618cb.rtf"),
		NewURL(TestSite),
		NewURL("http://example.net/"),
	}

	l := NewList(nil)
	_, err := l.AddFromFile("testdata/CIMBL-0666-CERTS.csv")
	require.NoError(t, err)
	require.NotEmpty(t, l)

	l1 := NewList(td2)
	l2 := l.Merge(l1)

	assert.Equal(t, 3, len(l.s))
	assert.EqualValues(t, tdm, l2.s)
	assert.EqualValues(t, tdm, l.s)
}
