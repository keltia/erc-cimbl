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
	assert.Error(t, err, "should be in error")
}

func TestLoadConfigBad(t *testing.T) {
	baseDir = "testdata"
	configName = "bad.toml"

	c, err := loadConfig()
	assert.Empty(t, c)
	assert.Error(t, err, "should be in error")
}

func TestLoadConfigPerms(t *testing.T) {
	baseDir = "testdata"
	configName = "config.toml"

	file := filepath.Join(baseDir, configName)
	err := os.Chmod(file, 0000)
	assert.NoError(t, err, "should be fine")

	c, err := loadConfig()
	assert.Empty(t, c)
	assert.Error(t, err, "should be in error")

	err = os.Chmod(file, 0644)
	assert.NoError(t, err, "should be fine")
}

func TestLoadConfigGood(t *testing.T) {
	baseDir = "testdata"
	configName = "config.toml"

	c, err := loadConfig()
	assert.NotEmpty(t, c, "not empty")
	assert.NoError(t, err, "should be fine")

	// Check values
	cnf := &Config{
		From:    "foo@example.com",
		To:      "security@example.com",
		Cc:      "root@example.com",
		Subject: "CRQ: New URLs/files to be BLOCKED",
		Server:  "SMTP:PORT",
		REFile:  `(?i:CIMBL-\d+-(CERTS|EU)\.(csv|zip)(\.asc|))`,
	}
	assert.EqualValues(t, cnf, c, "should be equal")
}

func TestLoadConfigGoodVerbose(t *testing.T) {
	baseDir = "testdata"
	configName = "config.toml"
	fVerbose = true

	c, err := loadConfig()
	assert.NotEmpty(t, c, "not empty")
	assert.NoError(t, err, "should be fine")

	// Check values
	cnf := &Config{
		From:    "foo@example.com",
		To:      "security@example.com",
		Cc:      "root@example.com",
		Subject: "CRQ: New URLs/files to be BLOCKED",
		Server:  "SMTP:PORT",
		REFile:  `(?i:CIMBL-\d+-(CERTS|EU)\.(csv|zip)(\.asc|))`,
	}
	assert.EqualValues(t, cnf, c, "should be equal")
	fVerbose = false
}
