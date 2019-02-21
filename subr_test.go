package main

import (
	"testing"

	"github.com/keltia/sandbox"
	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
)

func TestCheckFilename(t *testing.T) {
	baseDir = "testdata"
	config, err := loadConfig()
	assert.NoError(t, err)

	fVerbose = true

	snd, err := sandbox.New("test")
	require.NoError(t, err)
	defer snd.Cleanup()

	ctx := &Context{
		config:  config,
		tempdir: snd,
	}

	file := "foo.bar"
	res := checkFilename(ctx, file)
	assert.False(t, res)

	file = "CIMBL-0666-CERTS.csv"
	res = checkFilename(ctx, file)
	assert.True(t, res)

	file = "CIMBL-0666-EU.csv"
	res = checkFilename(ctx, file)
	assert.True(t, res)
}

func TestCheckOpenPGP(t *testing.T) {
	file := "foo.bar"
	res := checkOpenPGP(file)
	assert.False(t, res)

	file = "OpenPGP Encrypted File.asc"
	res = checkOpenPGP(file)
	assert.True(t, res)
}

func TestCheckMultipart(t *testing.T) {
	file := "foo.bar"
	res := checkMultipart(file)
	assert.False(t, res)

	file = "OpenPGP Encrypted File"
	res = checkMultipart(file)
	assert.True(t, res)
}
