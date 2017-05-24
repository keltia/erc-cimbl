package main


import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLoadConfigNone(t *testing.T) {
	baseDir = "test"
    configName = "/nonexistant"

	c, err := loadConfig()
    assert.Empty(t, c, "empty")
	assert.Error(t, err, "should be in error")
}

func TestLoadConfigGood(t *testing.T) {
    baseDir = "test"

    c, err := loadConfig()
    assert.NotEmpty(t, c, "not empty")
    assert.NoError(t, err, "should be fine")
}
