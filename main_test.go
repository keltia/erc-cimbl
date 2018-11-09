package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetup(t *testing.T) {
	baseDir = "testdata"

	ctx := setup()
	assert.NotNil(t, ctx)

	assert.NotNil(t, ctx.config)
	assert.Nil(t, ctx.tempdir)
}

func TestSetupNone(t *testing.T) {
	baseDir = "testx"

	ctx := setup()
	assert.NotNil(t, ctx)

	assert.Empty(t, ctx.config)
	assert.Nil(t, ctx.tempdir)
}

func TestSetupNoneDebug(t *testing.T) {
	baseDir = "testx"

	fDebug = true
	ctx := setup()
	assert.NotNil(t, ctx)

	assert.Empty(t, ctx.config)
	assert.Nil(t, ctx.tempdir)
	assert.True(t, fVerbose)

	fDebug = false
}

func setvars(t *testing.T) {
	// Insert our values
	require.NoError(t, os.Setenv("HTTP_PROXY", "http://proxy:8080/"))
	require.NoError(t, os.Setenv("HTTPS_PROXY", "http://proxy:8080/"))
	require.NoError(t, os.Setenv("http_proxy", "http://proxy:8080/"))
	require.NoError(t, os.Setenv("https_proxy", "http://proxy:8080/"))
}

func unsetvars(t *testing.T) {
	// Remove our values
	require.NoError(t, os.Unsetenv("HTTP_PROXY"))
	require.NoError(t, os.Unsetenv("HTTPS_PROXY"))
	require.NoError(t, os.Unsetenv("http_proxy"))
	require.NoError(t, os.Unsetenv("https_proxy"))
}

func TestSetupProxyError(t *testing.T) {
	setvars(t)

	baseDir = "testdata"
	netrc := filepath.Join(".", "testdata", "test-netrc")
	require.NoError(t, os.Chmod(netrc, 0600))
	require.NoError(t, os.Setenv("NETRC", netrc))

	ctx := setup()
	assert.NotNil(t, ctx)

	assert.NotNil(t, ctx.config)
	assert.Nil(t, ctx.tempdir)
	assert.NotEmpty(t, ctx.proxyauth)
	unsetvars(t)
}

func TestSetupProxyAuth(t *testing.T) {
	setvars(t)

	baseDir = "testdata"
	netrc := filepath.Join(".", "testdata", "test-netrc")
	require.NoError(t, os.Chmod(netrc, 0600))
	require.NoError(t, os.Setenv("NETRC", netrc))

	ctx := setup()
	assert.NotNil(t, ctx)

	assert.NotNil(t, ctx.config)
	assert.Nil(t, ctx.tempdir)
	assert.NotEmpty(t, ctx.proxyauth)
	unsetvars(t)
}

func TestSetupServer(t *testing.T) {
	baseDir = "testdata"
	configName = "config-smtp.toml"
	os.Setenv("NETRC", filepath.Join(".", "test", "test-netrc"))

	fDebug = true
	ctx := setup()
	assert.NotNil(t, ctx)

	assert.NotNil(t, ctx.config)
	assert.Nil(t, ctx.tempdir)
	assert.NotNil(t, ctx.proxyauth)
	assert.NotEmpty(t, ctx.config.Server)
	fDebug = false
}
