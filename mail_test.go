package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateMailNilContext(t *testing.T) {
	txt, err := createMail(nil, nil)
	assert.Error(t, err)
	assert.Empty(t, txt)
}

func TestCreateMailNilConfig(t *testing.T) {
	ctx := &Context{}
	txt, err := createMail(ctx, nil)
	assert.Error(t, err)
	assert.Empty(t, txt)
}

func TestAddPaths(t *testing.T) {
	results := &Results{Paths: map[string]bool{"foo.docx": true}}

	res := fmt.Sprintf("%s  %s\n", pathsTmpl, "foo.docx")
	str := addPaths(results)
	assert.Equal(t, res, str, "should be equal")
}

func TestAddURLsBlock(t *testing.T) {
	results := &Results{URLs: map[string]bool{"http://example.com/malware": true}}

	res := fmt.Sprintf("%s  %s\n", urlsTmpl, "http://example.com/malware")
	str := addURLs(results)
	assert.Equal(t, res, str, "should be equal")

}

func TestDoSendMailNoMail(t *testing.T) {
	baseDir = "testdata"
	configName = "config.toml"
	fVerbose = false

	config, err := loadConfig()
	assert.NoError(t, err, "no error")
	ctx := &Context{
		config: config,
	}
	res := &Results{Paths: map[string]bool{"foo.docx": true}}

	err = doSendMail(ctx, res)
	assert.NoError(t, err, "no error")
}

func TestDoSendMailConfigError(t *testing.T) {
	ctx := &Context{config: nil}
	res := &Results{Paths: map[string]bool{"/dontcare": true}}

	err := doSendMail(ctx, res)
	assert.Error(t, err)
}

func TestDoSendMailNoWork(t *testing.T) {
	baseDir = "testdata"
	configName = "config.toml"
	fVerbose = false

	config, err := loadConfig()
	assert.NoError(t, err)
	ctx := &Context{config: config}
	res := &Results{Paths: map[string]bool{}}

	err = doSendMail(ctx, res)
	assert.NoError(t, err, "no error")
}

func TestDoSendMailWithMail(t *testing.T) {
	baseDir = "testdata"
	configName = "config.toml"
	fVerbose = false

	config, err := loadConfig()
	assert.NoError(t, err, "no error")
	ctx := &Context{
		config: config,
		mail:   NullMailer{},
	}
	res := &Results{Paths: map[string]bool{"foo.docx": true}}
	fDoMail = true

	err = doSendMail(ctx, res)
	assert.NoError(t, err, "no error")
}

func TestDoSendMailWithMailDebug(t *testing.T) {
	baseDir = "testdata"
	configName = "config.toml"
	fDebug = true

	config, err := loadConfig()
	assert.NoError(t, err, "no error")
	ctx := &Context{
		config: config,
		mail:   NullMailer{},
	}
	res := &Results{Paths: map[string]bool{"foo.docx": true}}
	fDoMail = true

	err = doSendMail(ctx, res)
	assert.NoError(t, err, "no error")
	fDebug = false
}

func TestSMTPMailSender_SendMail(t *testing.T) {
	m := &SMTPMailSender{}
	err := m.SendMail("", "", nil, nil)
	assert.Error(t, err)
}
