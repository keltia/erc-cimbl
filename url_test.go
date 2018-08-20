package main

import (
	"github.com/jarcoal/httpmock"
	"github.com/keltia/proxy"
	"github.com/stretchr/testify/assert"
	"net"
	"net/http"
	"testing"
	"time"
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

	a = ""
	e := checkForIP(a)
	assert.NotNil(t, e)
	assert.Empty(t, e)
}

func TestDoCheck(t *testing.T) {
	var testSite string

	// Check values
	ctx := &Context{
		URLs: map[string]string{},
	}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	str := TestSite
	req, transport := proxy.SetupTransport(str)
	assert.NotNil(t, req, "not nil")
	assert.NotNil(t, transport, "not nil")

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

	ctx.Client = &http.Client{Transport: transport, Timeout: 10 * time.Second}

	res := doCheck(ctx, req)
	assert.Equal(t, "BLOCKED-EEC", res, "should be block")
}
