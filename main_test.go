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

	ctx, err := setup()
	assert.NotNil(t, ctx)
	assert.NoError(t, err)

	assert.NotNil(t, ctx.config)
	assert.NotNil(t, ctx.tempdir)
}

func TestSetupNone(t *testing.T) {
	baseDir = "testx"

	ctx, err := setup()
	assert.NotNil(t, ctx)
	assert.NoError(t, err)

	assert.Empty(t, ctx.config)
	assert.NotNil(t, ctx.tempdir)
}

func TestSetupNoneDebug(t *testing.T) {
	baseDir = "testx"

	fDebug = true
	ctx, err := setup()
	assert.NotNil(t, ctx)
	assert.NoError(t, err)

	assert.Empty(t, ctx.config)
	assert.NotNil(t, ctx.tempdir)
	assert.True(t, fVerbose)

	fDebug = false
}

func TestSetupNoneDebugSandboxInvalid(t *testing.T) {
	baseDir = "testdata"

	prev := os.Getenv("TMPDIR")
	if prev == "" {
		prev = "/tmp"
	}
	require.NoError(t, os.Setenv("TMPDIR", "/nonexistent"))

	fDebug = true
	ctx, err := setup()
	assert.Nil(t, ctx)
	assert.Error(t, err)

	require.NoError(t, os.Setenv("TMPDIR", prev))

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

	ctx, err := setup()
	assert.NotNil(t, ctx)
	assert.NoError(t, err)

	assert.NotNil(t, ctx.config)
	assert.NotNil(t, ctx.tempdir)
	assert.NotEmpty(t, ctx.proxyauth)
	unsetvars(t)
}

func TestSetupProxyAuth(t *testing.T) {
	setvars(t)

	baseDir = "testdata"
	netrc := filepath.Join(".", "testdata", "test-netrc")
	require.NoError(t, os.Chmod(netrc, 0600))
	require.NoError(t, os.Setenv("NETRC", netrc))

	ctx, err := setup()
	assert.NotNil(t, ctx)
	assert.NoError(t, err)

	assert.NotNil(t, ctx.config)
	assert.NotNil(t, ctx.tempdir)
	assert.NotEmpty(t, ctx.proxyauth)
	unsetvars(t)
}

func TestSetupServer(t *testing.T) {
	baseDir = "testdata"
	configName = "config-smtp.toml"
	os.Setenv("NETRC", filepath.Join(".", "test", "test-netrc"))

	fDebug = true
	ctx, err := setup()
	assert.NotNil(t, ctx)
	assert.NoError(t, err)

	assert.NotNil(t, ctx.config)
	assert.NotNil(t, ctx.tempdir)
	assert.NotNil(t, ctx.proxyauth)
	assert.NotEmpty(t, ctx.config.Server)
	fDebug = false
}

func TestRealMain_Noarg(t *testing.T) {
	err := realmain([]string{})
	assert.NoError(t, err)
}

func TestRealMain_Badtemp(t *testing.T) {
	old := os.Getenv("TMPDIR")
	require.NoError(t, os.Setenv("TMPDIR", "/nonexistent"))

	err := realmain([]string{})
	assert.Error(t, err)

	require.NoError(t, os.Setenv("TMPDIR", old))
}

func TestRealMain_Onebadarg(t *testing.T) {
	err := realmain([]string{"/foo.bar"})
	assert.NoError(t, err)
}

func TestRealMain_Onearg_Empty(t *testing.T) {
	err := realmain([]string{"testdata/CIMBL-0667-CERTS.csv"})
	assert.NoError(t, err)
}

func TestRealMain_Onearg_Good(t *testing.T) {
	err := realmain([]string{"testdata/CIMBL-0666-CERTS.csv"})
	assert.NoError(t, err)
}

func TestRealMain_Onearg_GoodZip(t *testing.T) {
	err := realmain([]string{"testdata/CIMBL-0666-CERTS.zip"})
	assert.NoError(t, err)
}

func TestUsage(t *testing.T) {
	Usage()
}
