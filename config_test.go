package main


import (
	"github.com/stretchr/testify/assert"
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
        Subject: "CIMBL",
        Server:  "SMTP",
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
        Subject: "CIMBL",
        Server:  "SMTP",
    }
    assert.Equal(t, cnf, c, "should be equal")
    fVerbose = false
}
