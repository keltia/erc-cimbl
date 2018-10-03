package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/keltia/sandbox"
	"github.com/proglottis/gpgme"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type NullGPGError struct{}

var ErrFakeGPGError = fmt.Errorf("fake error")

func (NullGPGError) Decrypt(r io.Reader) (*gpgme.Data, error) {
	return &gpgme.Data{}, ErrFakeGPGError
}

func TestNullGPG_Decrypt(t *testing.T) {
	baseDir = "testdata"
	config, err := loadConfig()
	assert.NoError(t, err)

	fVerbose = true

	snd, err := sandbox.New("test")
	require.NoError(t, err)
	defer snd.Cleanup()

	ctx := &Context{
		config:  config,
		Paths:   map[string]bool{},
		URLs:    map[string]string{},
		tempdir: snd,
		gpg:     NullGPG{},
	}

	file := "testdata/CIMBL-0666-CERTS.zip.asc"
	fh, err := os.Open(file)
	require.NoError(t, err)
	require.NotNil(t, fh)

	plain, err := ctx.gpg.Decrypt(fh)
	assert.NoError(t, err)
	assert.Empty(t, plain)
}

func TestNullGPGError_Decrypt(t *testing.T) {
	baseDir = "testdata"
	config, err := loadConfig()
	assert.NoError(t, err)

	fVerbose = true

	snd, err := sandbox.New("test")
	require.NoError(t, err)
	defer snd.Cleanup()

	ctx := &Context{
		config:  config,
		Paths:   map[string]bool{},
		URLs:    map[string]string{},
		tempdir: snd,
		gpg:     NullGPGError{},
	}

	file := "testdata/CIMBL-0666-CERTS.zip.asc"
	fh, err := os.Open(file)
	require.NoError(t, err)
	require.NotNil(t, fh)

	plain, err := ctx.gpg.Decrypt(fh)
	assert.Error(t, err)
	assert.Equal(t, ErrFakeGPGError, err)
	assert.Empty(t, plain)
}

func TestDecryptFileNull(t *testing.T) {
	baseDir = "testdata"
	config, err := loadConfig()
	assert.NoError(t, err)

	fVerbose = true

	snd, err := sandbox.New("test")
	require.NoError(t, err)
	defer snd.Cleanup()

	ctx := &Context{
		config:  config,
		Paths:   map[string]bool{},
		URLs:    map[string]string{},
		tempdir: snd,
		gpg:     NullGPG{},
	}

	file, err := filepath.Abs("testdata/CIMBL-0666-CERTS.zip.asc")
	require.NoError(t, err)

	fh, err := os.Open(file)
	require.NoError(t, err)
	require.NotNil(t, fh)

	// create our fake zip file
	dfh, err := os.Create(filepath.Join(snd.Cwd(), "CIMBL-0666-CERTS.zip"))
	require.NoError(t, err)
	defer dfh.Close()

	// copy into sandbox
	_, err = io.Copy(dfh, fh)
	require.NoError(t, err)

	snd.Enter()
	plain, err := decryptFile(ctx, file)
	snd.Exit()
	assert.Error(t, err)
	assert.Empty(t, plain)
}

func TestDecryptFileNone(t *testing.T) {
	baseDir = "testdata"
	config, err := loadConfig()
	assert.NoError(t, err)

	fVerbose = true

	snd, err := sandbox.New("test")
	require.NoError(t, err)
	defer snd.Cleanup()

	ctx := &Context{
		config:  config,
		Paths:   map[string]bool{},
		URLs:    map[string]string{},
		tempdir: snd,
		gpg:     NullGPG{},
	}

	snd.Enter()
	plain, err := decryptFile(ctx, "/nonexistent")
	snd.Exit()
	assert.Error(t, err)
	assert.Empty(t, plain)
}

func TestGpgme_Decrypt(t *testing.T) {
	baseDir = "testdata"
	config, err := loadConfig()
	assert.NoError(t, err)

	fVerbose = true

	snd, err := sandbox.New("test")
	require.NoError(t, err)
	defer snd.Cleanup()

	ctx := &Context{
		config:  config,
		Paths:   map[string]bool{},
		URLs:    map[string]string{},
		tempdir: snd,
		gpg:     Gpgme{},
	}

	file := "testdata/CIMBL-0666-CERTS.zip.asc"
	fh, err := os.Open(file)
	require.NoError(t, err)
	require.NotNil(t, fh)

	plain, err := ctx.gpg.Decrypt(fh)
	assert.Error(t, err)
	assert.NotEmpty(t, plain)
}
