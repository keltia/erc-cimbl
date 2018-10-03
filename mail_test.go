package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateMailNilContext(t *testing.T) {
	txt, err := createMail(nil)
	assert.Error(t, err)
	assert.Empty(t, txt)
}

func TestCreateMailNilConfig(t *testing.T) {
	ctx := &Context{}
	txt, err := createMail(ctx)
	assert.Error(t, err)
	assert.Empty(t, txt)
}

func TestAddPaths(t *testing.T) {
	ctx := &Context{Paths: map[string]bool{"foo.docx": true}}

	res := fmt.Sprintf("%s  %s\n", pathsTmpl, "foo.docx")
	str := addPaths(ctx)
	assert.Equal(t, res, str, "should be equal")
}

func TestAddURLsBlock(t *testing.T) {
	ctx := &Context{URLs: map[string]string{"http://example.com/malware": "**BLOCK**"}}

	res := fmt.Sprintf("%s  %s\n", urlsTmpl, "http://example.com/malware")
	str := addURLs(ctx)
	assert.Equal(t, res, str, "should be equal")

}

func TestAddURLsUnknown(t *testing.T) {
	ctx := &Context{URLs: map[string]string{"http://example.com/malware": "UNKNOWN"}}

	str := addURLs(ctx)
	assert.Equal(t, urlsTmpl, str, "should be equal")
}

func TestDoSendMailNoMail(t *testing.T) {
	baseDir = "testdata"
	configName = "config.toml"
	fVerbose = false

	config, err := loadConfig()
	assert.NoError(t, err, "no error")
	ctx := &Context{
		config: config,
		Paths:  map[string]bool{"foo.docx": true},
	}

	err = doSendMail(ctx)
	assert.NoError(t, err, "no error")
}

func TestDoSendMailConfigError(t *testing.T) {
	ctx := &Context{
		config: nil,
	}

	err := doSendMail(ctx)
	assert.Error(t, err)
}

func TestDoSendMailNoWork(t *testing.T) {
	baseDir = "testdata"
	configName = "config.toml"
	fVerbose = false

	config, err := loadConfig()
	assert.NoError(t, err)
	ctx := &Context{
		config: config,
		Paths:  map[string]bool{},
	}

	err = doSendMail(ctx)
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
		Paths:  map[string]bool{"foo.docx": true},
		mail:   NullMailer{},
	}
	fDoMail = true

	err = doSendMail(ctx)
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
		Paths:  map[string]bool{"foo.docx": true},
		mail:   NullMailer{},
	}
	fDoMail = true

	err = doSendMail(ctx)
	assert.NoError(t, err, "no error")
	fDebug = false
}

func TestSMTPMailSender_SendMail(t *testing.T) {
	m := &SMTPMailSender{}
	err := m.SendMail("", "", nil, nil)
	assert.Error(t, err)
}
