package main

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/h2non/gock"
	"github.com/keltia/proxy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	TestSite = "http://pontonerywariva342.top/search.php"
)

func TestSanitize(t *testing.T) {
	urls := []struct {
		url string
		res string
		err error
	}{
		{"https://example.com", "https://example.com", ErrHttpsSkip},
		{"http://example.com", "http://example.com", nil},
		{"ttp://example.com", "http://example.com", nil},
		{"://example.com", "http://://example.com", nil},
		{"example.com", "http://example.com", nil},
		{"example.com/foo.php", "http://example.com/foo.php", nil},
		{"http://[1.2.3.4]", "http://1.2.3.4", nil},
		{"[1.2.3.4]", "http://1.2.3.4", nil},
		{"103.15.234.152:80/index.php", "http://103.15.234.152:80/index.php", nil},
		{":/--%2Fexample.com", "http://:/--%2Fexample.com", nil},
		{"https://jtabserver.org/bins/jayct.vbs&amp#39;,&amp;#39;%ALLUSERSPROFILE%\\jayct.vbs&amp;quot;", "", ErrParseError},
	}
	for _, u := range urls {
		t.Logf("url=%s", u.url)
		n, err := sanitize(u.url)
		assert.Equal(t, u.res, n)
		assert.Equal(t, u.err, err)
	}
}

func TestCheckForIP(t *testing.T) {
	a := "1.2.3.4"
	b := checkForIP(a)
	c := net.ParseIP(a)
	assert.NotNil(t, b)
	assert.Equal(t, c, b)

	a = "[1.2.3.4]"
	d := checkForIP(a)
	assert.NotNil(t, d)
	assert.Equal(t, c, d)
	assert.Equal(t, b, d)

	a = ""
	e := checkForIP(a)
	assert.NotNil(t, e)
	assert.Empty(t, e)
}

func TestDoCheck403(t *testing.T) {
	defer gock.Off()

	fDebug = true

	// Check values
	ctx := &Context{}

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

	req, err := http.NewRequest("HEAD", TestSite, nil)
	require.NoError(t, err)

	res, err := doCheck(ctx, req)
	t.Logf("res=%#v", res)
	assert.NoError(t, err)
	assert.Equal(t, ActionBlocked, res)

	fDebug = false
}

func TestDoCheck200(t *testing.T) {
	defer gock.Off()

	// Check values
	ctx := &Context{}

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

	req, err := http.NewRequest("HEAD", TestSite, nil)
	require.NoError(t, err)

	res, err := doCheck(ctx, req)
	assert.NoError(t, err)
	assert.Equal(t, ActionBlock, res)
}

func TestDoCheck407(t *testing.T) {
	defer gock.Off()

	// Check values
	ctx := &Context{}

	_, transport := proxy.SetupTransport(TestSite)
	require.NotNil(t, transport)

	// Set up minimal client
	ctx.Client = &http.Client{Transport: transport, Timeout: 10 * time.Second}

	testSite, err := url.Parse(TestSite)
	require.NoError(t, err)

	gock.New(testSite.Host).
		Head(testSite.Path).
		Reply(407)

	gock.InterceptClient(ctx.Client)
	defer gock.RestoreClient(ctx.Client)

	req, err := http.NewRequest("HEAD", TestSite, nil)
	require.NoError(t, err)

	res, err := doCheck(ctx, req)
	assert.NoError(t, err)
	assert.Equal(t, ActionAuth, res)
}

func TestDoCheckError(t *testing.T) {
	// Check values
	ctx := &Context{}

	_, transport := proxy.SetupTransport(TestSite)
	require.NotNil(t, transport)

	// Set up minimal client
	ctx.Client = &http.Client{Transport: transport, Timeout: 10 * time.Second}

	req, err := http.NewRequest("HEAD", TestSite, nil)
	require.NoError(t, err)

	req.URL = nil
	_, err = doCheck(ctx, req)
	assert.Error(t, err)
}

func TestDoCheck_Transport(t *testing.T) {
	// Check values
	ctx := &Context{}

	// Set up minimal client
	ctx.Client = &http.Client{Transport: nil, Timeout: 10 * time.Second}

	req, err := http.NewRequest("HEAD", TestSite, nil)
	require.NoError(t, err)

	_, err = doCheck(ctx, req)
	assert.Error(t, err)
}

func TestHandleURLhttps(t *testing.T) {
	defer gock.Off()

	skipped = []string{}

	// Check values
	ctx := &Context{}

	u, err := handleURL(ctx, "https://example.com")
	assert.NoError(t, err)
	assert.Empty(t, u)
	assert.EqualValues(t, []string{"https://example.com"}, skipped)
}

func TestHandleURLblocked(t *testing.T) {
	defer gock.Off()

	// Check values
	ctx := &Context{}

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

	u, err := handleURL(ctx, TestSite)
	assert.NoError(t, err)
	require.Empty(t, u)
}

func TestHandleURLblock(t *testing.T) {
	defer gock.Off()

	// Check values
	ctx := &Context{}

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

	u, err := handleURL(ctx, TestSite)
	assert.NoError(t, err)
	require.NotEmpty(t, u)
	assert.Equal(t, u, TestSite)
}

func TestHandleURLno(t *testing.T) {
	// Check values
	ctx := &Context{}

	fNoURLs = true

	u, err := handleURL(ctx, TestSite)
	assert.NoError(t, err)
	require.Empty(t, u)

	fNoURLs = false
}
