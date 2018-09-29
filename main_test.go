package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckFilename(t *testing.T) {
	file := "foo.bar"
	res := checkFilename(file)
	assert.False(t, res, "should be false")

	file = "CIMBL-0666-CERTS.csv"
	res = checkFilename(file)
	assert.True(t, res, "should be true")
}

func TestSetup(t *testing.T) {
	baseDir = "test"

	ctx := setup()
	assert.NotNil(t, ctx)

	assert.NotNil(t, ctx.config)
	assert.Empty(t, ctx.URLs)
	assert.Empty(t, ctx.Paths)
	assert.Nil(t, ctx.tempdir)
}

func TestSetupNone(t *testing.T) {
	baseDir = "testx"

	ctx := setup()
	assert.NotNil(t, ctx)

	assert.Empty(t, ctx.config)
	assert.Empty(t, ctx.URLs)
	assert.Empty(t, ctx.Paths)
	assert.Nil(t, ctx.tempdir)
}

func TestSetupProxy(t *testing.T) {
	baseDir = "test"
	os.Setenv("NETRC", filepath.Join(".", "test", "test-netrc"))

	ctx := setup()
	assert.NotNil(t, ctx)

	assert.NotNil(t, ctx.config)
	assert.Empty(t, ctx.URLs)
	assert.Empty(t, ctx.Paths)
	assert.Nil(t, ctx.tempdir)
	assert.NotNil(t, ctx.proxyauth)
}
