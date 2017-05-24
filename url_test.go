package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSetupProxyNoFile(t *testing.T) {
	file := ""
	user = ""
	password = ""

	err := setupProxy(file)
	assert.Error(t, err, "error")
	assert.EqualValues(t, "", user, "null user")
	assert.EqualValues(t, "", password, "null password")
}

func TestSetupProxyZero(t *testing.T) {
	file := "test/zero-dbrc"
	user = ""
	password = ""

	err := setupProxy(file)
	assert.Error(t, err, "error")
	assert.EqualValues(t, "", user, "test user")
	assert.EqualValues(t, "", password, "test password")
}

func TestSetupProxyGood(t *testing.T) {
	file := "test/test-dbrc"
	user = ""
	password = ""

	err := setupProxy(file)
	assert.NoError(t, err, "no error")
	assert.EqualValues(t, "test", user, "test user")
	assert.EqualValues(t, "test", password, "test password")
}
