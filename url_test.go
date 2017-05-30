package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
	"net/http"
	"os"
)

func TestSetupTransport(t *testing.T) {
	url := "foo.bar\\%%%%%%"
	ctx := &Context{
		URLs:   map[string]string{},
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
		URLs:   map[string]string{},
	}

	str := "http://pontonerywariva342.top/search.php"
	req, transport := setupTransport(ctx, str)
	assert.NotNil(t, req, "not nil")
	assert.NotNil(t, transport, "not nil")

	uri, err := getProxy(req)
	assert.Nil(t, uri, "should be nil")
	assert.NoError(t, err, "no error")
}

func TestDoCheck(t *testing.T) {
	// Check values
	ctx := &Context{
		URLs:   map[string]string{},
	}

	str := "http://pontonerywariva342.top/search.php"
	req, transport := setupTransport(ctx, str)
	assert.NotNil(t, req, "not nil")
	assert.NotNil(t, transport, "not nil")

	ctx.Client = &http.Client{Transport: transport, Timeout: 10 * time.Second}

	res := doCheck(ctx, req)
	assert.Equal(t, "**BLOCK**", res, "should be block")
}
