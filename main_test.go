package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCheckFilename(t *testing.T) {
	file := "foo.bar"
	res := checkFilename(file)
	assert.False(t, res, "should be false")

	file = "CIMBL-0666-CERTS.csv"
	res = checkFilename(file)
	assert.True(t, res, "should be true")
}
