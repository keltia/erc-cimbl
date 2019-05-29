package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Filename

func TestNewFilename(t *testing.T) {
	fn := NewFilename("")
	assert.IsType(t, (*Filename)(nil), fn)
	assert.Empty(t, fn.Name)
}

func TestNewFilename2(t *testing.T) {
	fn := NewFilename("foo")
	assert.IsType(t, (*Filename)(nil), fn)
	assert.Equal(t, "foo", fn.Name)
}

func TestFilename_AddTo(t *testing.T) {
	td := map[string]bool{"example.docx":true}

	fn := NewFilename("example.docx")

	r := NewResults()
	fn.AddTo(r)
	assert.NotEmpty(t, r.Paths)
	assert.Equal(t, td, r.Paths)
}

func TestFilename_Check(t *testing.T) {
	fn := NewFilename("example.docx")
	assert.True(t, fn.Check())
}

// URL

func TestNewURL(t *testing.T) {
	u := NewURL("")
	assert.IsType(t, (*URL)(nil), u)
	assert.Empty(t, u.H)
}

func TestURL_AddTo(t *testing.T) {
	td := map[string]bool{"http://www.example.net/":true}

	u := NewURL("http://www.example.net/")
	r := NewResults()
	u.AddTo(r)
	assert.NotEmpty(t, r.URLs)
	assert.Equal(t, td, r.URLs)
}

func TestURL_Check(t *testing.T) {
	u := NewURL("http://www.example.net/")
	assert.True(t, u.Check())
}

// List

func TestNewList(t *testing.T) {
	ctx := &Context{}
	l := NewList(ctx, nil)
	assert.Empty(t, l)
}

func TestNewList2(t *testing.T) {
	ctx := &Context{}
	l := NewList(ctx, []string{})
	assert.Empty(t, l)
}

func TestNewList3(t *testing.T) {
	td := NewURL("http://www.example.net/")

	ctx := &Context{
		config:&Config{REFile:`(?i:CIMBL-\\d+-(CERTS|EU)\\.(csv|zip)(\\.asc|))`},
	}

	l := NewList(ctx, []string{"http://www.example.net/"})
	assert.NotEmpty(t, l)
	assert.Equal(t, td, l.s[0])
}

func TestNewList4(t *testing.T) {
	ctx := &Context{
		config:&Config{REFile:`(?i:CIMBL-\d+-(CERTS|EU)\.(csv|zip)(\.asc|))`},
	}

	l := NewList(ctx, []string{"testdata/CIMBL-0666-CERTS.csv"})
	assert.NotEmpty(t, l)
	t.Log(l.s)
	//assert.Equal(t, "", l.s[0])
}

func TestList_Add(t *testing.T) {
	td := []Sourcer{NewFilename("exemple.docx")}
	ctx := &Context{}
	l := NewList(ctx, nil)
	l.Add(NewFilename("exemple.docx"))
	assert.NotEmpty(t, l.s)
	assert.Equal(t, td, l.s)
}

func TestList_AddFromFile(t *testing.T) {

}

func TestList_Check(t *testing.T) {

}

