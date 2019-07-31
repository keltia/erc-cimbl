package main

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/h2non/gock"
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
		{"http://example.onion", "http://example.onion", ErrHttpsSkip},
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

func TestHandleURLhttps(t *testing.T) {
	defer gock.Off()

	skipped = []string{}

	proxy := os.Getenv("http_proxy")
	c := resty.New().SetProxy(proxy)

	u, err := handleURL(c, "https://example.com")
	assert.NoError(t, err)
	assert.Empty(t, u)
	assert.EqualValues(t, []string{"https://example.com"}, skipped)
}

func TestHandleURLblocked(t *testing.T) {
	defer gock.Off()

	proxy := os.Getenv("http_proxy")
	c := resty.New().SetProxy(proxy)

	testSite, err := url.Parse(TestSite)
	require.NoError(t, err)

	gock.New(testSite.Host).
		Head(testSite.Path).
		MatchHeader("user-agent", fmt.Sprintf("%s/%s", MyName, MyVersion)).
		Reply(403)

	gock.InterceptClient(c.GetClient())
	defer gock.RestoreClient(c.GetClient())

	u, err := handleURL(c, TestSite)
	assert.NoError(t, err)
	require.EqualValues(t, ActionBlocked, u)
}

func TestHandleURLAuth(t *testing.T) {
	defer gock.Off()

	proxy := os.Getenv("http_proxy")
	c := resty.New().SetProxy(proxy)

	testSite, err := url.Parse(TestSite)
	require.NoError(t, err)

	gock.New(testSite.Host).
		Head(testSite.Path).
		MatchHeader("user-agent", fmt.Sprintf("%s/%s", MyName, MyVersion)).
		Reply(407)

	gock.InterceptClient(c.GetClient())
	defer gock.RestoreClient(c.GetClient())

	u, err := handleURL(c, TestSite)
	assert.NoError(t, err)
	require.EqualValues(t, ActionAuth, u)
}

func TestHandleURLNope(t *testing.T) {
	defer gock.Off()

	proxy := os.Getenv("http_proxy")
	c := resty.New().SetProxy(proxy)

	testSite, err := url.Parse(TestSite)
	require.NoError(t, err)

	gock.New(testSite.Host).
		Head(testSite.Path).
		MatchHeader("user-agent", fmt.Sprintf("%s/%s", MyName, MyVersion)).
		Reply(503)

	gock.InterceptClient(c.GetClient())
	defer gock.RestoreClient(c.GetClient())

	u, err := handleURL(c, TestSite)
	assert.NoError(t, err)
	require.EqualValues(t, ActionBlocked, u)
}

func TestHandleURLblock(t *testing.T) {
	defer gock.Off()

	proxy := os.Getenv("http_proxy")
	c := resty.New().SetProxy(proxy)

	testSite, err := url.Parse(TestSite)
	require.NoError(t, err)

	gock.New(testSite.Host).
		Head(testSite.Path).
		MatchHeader("user-agent", fmt.Sprintf("%s/%s", MyName, MyVersion)).
		Reply(200)

	gock.InterceptClient(c.GetClient())
	defer gock.RestoreClient(c.GetClient())

	u, err := handleURL(c, TestSite)
	assert.NoError(t, err)
	require.NotEmpty(t, u)
	assert.Equal(t, u, TestSite)
}

func TestHandleURLno(t *testing.T) {

	proxy := os.Getenv("http_proxy")
	c := resty.New().SetProxy(proxy)

	fNoURLs = true

	u, err := handleURL(c, TestSite)
	assert.NoError(t, err)
	require.Empty(t, u)

	fNoURLs = false
}
