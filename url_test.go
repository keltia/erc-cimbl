package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
	"net/http"
)

func TestCheckSetup(t *testing.T) {
	url := "foo.bar\\%%%%%%"
	r, tr := setupCheck(url)
	assert.Nil(t, r, "should be nil")
	assert.Nil(t, tr, "should be nil")
}

func TestDoCheck(t *testing.T) {
	// Check values
	cnf := &Config{
		From:    "foo@example.com",
		To:      "security@example.com",
		Cc:      "root@example.com",
		Subject: "CRQ: New URLs/files to be BLOCKED",
		Server:  "SMTP",
	}

	ctx := &Context{
		config: cnf,
		Paths:  map[string]bool{},
		URLs:   map[string]string{},
	}

	str := "http://pontonerywariva342.top/search.php"
	req, transport := setupCheck(str)
	assert.NotNil(t, req, "not nil")
	assert.NotNil(t, transport, "not nil")

	ctx.Client = &http.Client{Transport: transport, Timeout: 10 * time.Second}

	res := doCheck(ctx, req)
	assert.Equal(t, "**BLOCK**", res, "should be block")
}
