package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfigNone(t *testing.T) {
	baseDir = "testdata"
	configName = "/nonexistant"

	c, err := loadConfig()
	assert.NotNil(t, c)
	assert.Empty(t, c)
	assert.Error(t, err)
}

func TestLoadConfigBad(t *testing.T) {
	baseDir = "testdata"
	configName = "bad.toml"

	c, err := loadConfig()
	assert.Empty(t, c)
	assert.Error(t, err)
}

func TestLoadConfigPerms(t *testing.T) {
	baseDir = "testdata"
	configName = "config.toml"

	file := filepath.Join(baseDir, configName)
	err := os.Chmod(file, 0000)
	assert.NoError(t, err)

	c, err := loadConfig()
	assert.Empty(t, c)
	assert.Error(t, err)

	err = os.Chmod(file, 0644)
	assert.NoError(t, err)
}

func TestLoadConfigGood(t *testing.T) {
	baseDir = "testdata"
	configName = "config.toml"

	c, err := loadConfig()
	assert.NotEmpty(t, c)
	assert.NoError(t, err)

	// Check values
	cnf := &Config{
		From:    "foo@example.com",
		To:      "security@example.com",
		Cc:      "root@example.com",
		Subject: "CRQ: New URLs/files to be BLOCKED",
		Server:  "SMTP:PORT",
		REFile:  `(?i:CIMBL-\d+-(CERTS|EU)\.(csv|zip)(\.asc|))`,
	}
	assert.EqualValues(t, cnf, c)
}

func TestLoadConfigGood_NoRE(t *testing.T) {
	baseDir = "testdata"
	configName = "config-nore.toml"

	c, err := loadConfig()
	assert.NotEmpty(t, c, "not empty")
	assert.NoError(t, err)

	// Check values
	cnf := &Config{
		From:    "foo@example.com",
		To:      "security@example.com",
		Cc:      "root@example.com",
		Subject: "CRQ: New URLs/files to be BLOCKED",
		Server:  "SMTP:PORT",
	}
	assert.EqualValues(t, cnf, c)
}

func TestLoadConfigGoodVerbose(t *testing.T) {
	baseDir = "testdata"
	configName = "config.toml"
	fVerbose = true

	c, err := loadConfig()
	assert.NotEmpty(t, c, "not empty")
	assert.NoError(t, err)

	// Check values
	cnf := &Config{
		From:    "foo@example.com",
		To:      "security@example.com",
		Cc:      "root@example.com",
		Subject: "CRQ: New URLs/files to be BLOCKED",
		Server:  "SMTP:PORT",
		REFile:  `(?i:CIMBL-\d+-(CERTS|EU)\.(csv|zip)(\.asc|))`,
	}
	assert.EqualValues(t, cnf, c)
	fVerbose = false
}
