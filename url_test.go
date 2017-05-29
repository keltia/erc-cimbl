package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCheckSetup(t *testing.T) {
	url := "foo.bar\\%%%%%%"
	r, tr := setupCheck(url)
	assert.Nil(t, r, "should be nil")
	assert.Nil(t, tr, "should be nil")
}

func TestDoCheck(t *testing.T) {

}

