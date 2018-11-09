package main

import (
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
