package main

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"os"
	"testing"
	"time"
	"github.com/jarcoal/httpmock"
)

const (
	TestSite = "http://pontonerywariva342.top/search.php"
)

func TestSetupTransport(t *testing.T) {
	url := "foo.bar\\%%%%%%"
	ctx := &Context{
		URLs: map[string]string{},
	}

	r, tr := setupTransport(ctx, url)
	assert.Nil(t, r, "should be nil")
	assert.Nil(t, tr, "should be nil")
}

func TestGetProxy(t *testing.T) {
	// Cleanup
	for _, env := range []string{
		"http_proxy",
		"https_proxy",
		"HTTP_PROXY",
		"HTTPS_PROXY",
	} {
		os.Unsetenv(env)
	}
	ctx := &Context{
		URLs: map[string]string{},
	}

	str := TestSite
	req, transport := setupTransport(ctx, str)
	assert.NotNil(t, req, "not nil")
	assert.NotNil(t, transport, "not nil")

	uri, err := getProxy(req)
	assert.Nil(t, uri, "should be nil")
	assert.NoError(t, err, "no error")
}

func TestDoCheck(t *testing.T) {
	var testSite string

	// Check values
	ctx := &Context{
		URLs: map[string]string{},
	}

	err := setupProxyAuth(ctx, dbrcFile)
	if err != nil {
		t.Log("No dbrc file, no proxy auth.")
	}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	str := TestSite
	req, transport := setupTransport(ctx, str)
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
	assert.Equal(t, "**BLOCK**", res, "should be block")
}

