package main

import (
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfigNone(t *testing.T) {
	baseDir = "test"
	configName = "/nonexistant"
	empty := &Config{}

	c, err := loadConfig()
	assert.Equal(t, empty, c, "empty")
	assert.Error(t, err, "should be in error")
}

func TestLoadConfigBad(t *testing.T) {
	baseDir = "test"
	configName = "bad.toml"

	c, err := loadConfig()
	assert.Nil(t, c, "nil value")
	assert.Error(t, err, "should be in error")
}

func TestLoadConfigPerms(t *testing.T) {
	baseDir = "test"
	configName = "config.toml"

	file := filepath.Join(baseDir, configName)
	err := os.Chmod(file, 0000)
	assert.NoError(t, err, "should be fine")

	c, err := loadConfig()
	assert.Nil(t, c, "nil value")
	assert.Error(t, err, "should be in error")

	err = os.Chmod(file, 0644)
	assert.NoError(t, err, "should be fine")
}

func TestLoadConfigGood(t *testing.T) {
	baseDir = "test"
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
	}
	assert.Equal(t, cnf, c, "should be equal")
}

func TestLoadConfigGoodVerbose(t *testing.T) {
	baseDir = "test"
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
	}
	assert.Equal(t, cnf, c, "should be equal")
	fVerbose = false
}

// ---
func TestSetupProxyAuth(t *testing.T) {
	dbrc := "test/no-dbrc"
	ctx := &Context{}

	err := setupProxyAuth(ctx, dbrc)
	assert.Error(t, err, "should be an error")

	dbrc = "test/test-dbrc"
	err = setupProxyAuth(ctx, dbrc)
	assert.NoError(t, err, "no error")
}

func TestSetupProxyAuthVerbose(t *testing.T) {
	fVerbose = true
	ctx := &Context{}

	dbrc := "test/no-dbrc"
	err := setupProxyAuth(ctx, dbrc)
	assert.Error(t, err, "should be an error")

	dbrc = "test/test-dbrc"
	err = setupProxyAuth(ctx, dbrc)
	assert.NoError(t, err, "no error")
}

func TestLoadDbrcNoFile(t *testing.T) {
	file := ""
	user = ""
	password = ""

	err := loadDbrc(file)
	assert.Error(t, err, "error")
	assert.EqualValues(t, "", user, "null user")
	assert.EqualValues(t, "", password, "null password")
}

func TestLoadDbrcZero(t *testing.T) {
	file := "test/zero-dbrc"
	user = ""
	password = ""

	err := loadDbrc(file)
	assert.Error(t, err, "error")
	assert.EqualValues(t, "", user, "test user")
	assert.EqualValues(t, "", password, "test password")
}

func TestLoadDbrcGood(t *testing.T) {
	file := "test/test-dbrc"
	user = ""
	password = ""

	err := loadDbrc(file)
	assert.NoError(t, err, "no error")
	assert.EqualValues(t, "test", user, "test user")
	assert.EqualValues(t, "test", password, "test password")
}
