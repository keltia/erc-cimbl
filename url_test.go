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
		{"://example.com", "://example.com", ErrParseError},
		{"http://[1.2.3.4]", "http://1.2.3.4", nil},
		{"[1.2.3.4]", "http://1.2.3.4", nil},
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
}

func TestDoCheck403(t *testing.T) {
	defer gock.Off()

	// Check values
	ctx := &Context{
		URLs: map[string]string{},
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

	req, err := http.NewRequest("HEAD", TestSite, nil)
	require.NoError(t, err)

	res, err := doCheck(ctx, req)
	assert.NoError(t, err)
	assert.Equal(t, ActionBlocked, res)
}

func TestDoCheck200(t *testing.T) {
	defer gock.Off()

	// Check values
	ctx := &Context{
		URLs: map[string]string{},
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

	req, err := http.NewRequest("HEAD", TestSite, nil)
	require.NoError(t, err)

	res, err := doCheck(ctx, req)
	assert.NoError(t, err)
	assert.Equal(t, ActionBlock, res)
}

func TestDoCheck407(t *testing.T) {
	defer gock.Off()

	// Check values
	ctx := &Context{
		URLs: map[string]string{},
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
		Reply(407)

	gock.InterceptClient(ctx.Client)
	defer gock.RestoreClient(ctx.Client)

	req, err := http.NewRequest("HEAD", TestSite, nil)
	require.NoError(t, err)

	res, err := doCheck(ctx, req)
	assert.NoError(t, err)
	assert.Equal(t, ActionAuth, res)
}

func TestHandleURLhttps(t *testing.T) {
	defer gock.Off()

	skipped = []string{}

	// Check values
	ctx := &Context{
		URLs: map[string]string{},
	}

	handleURL(ctx, "https://example.com")
	assert.Empty(t, ctx.URLs)
	assert.EqualValues(t, []string{"https://example.com"}, skipped)
}

func TestHandleURLblocked(t *testing.T) {
	defer gock.Off()

	// Check values
	ctx := &Context{
		URLs: map[string]string{},
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

	handleURL(ctx, TestSite)
	require.NotEmpty(t, ctx.URLs)
	assert.EqualValues(t, "", ctx.URLs[TestSite])
}

func TestHandleURLblock(t *testing.T) {
	defer gock.Off()

	// Check values
	ctx := &Context{
		URLs: map[string]string{},
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

	handleURL(ctx, TestSite)
	require.NotEmpty(t, ctx.URLs)
	assert.EqualValues(t, ActionBlock, ctx.URLs[TestSite])
}
