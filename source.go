package main

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

type Sourcer interface {
	Check() bool
	AddTo(r *Results)
}

type URL struct {
	H string
}

func NewURL(u string) *URL {
	return &URL{H: u}
}

func (u *URL) Check() bool {
	return true
}

func (u *URL) AddTo(r *Results) {
	r.Add("url", u.H)
}

type Filename struct {
	Name string
}

func NewFilename(s string) *Filename {
	return &Filename{Name: s}
}

func (f *Filename) Check() bool {
	return true
}

func (f *Filename) AddTo(r *Results) {
	r.Add("filename", f.Name)
}

type List struct {
	s []Sourcer
}

func NewList(ctx *Context, files []string) *List {
	l := new(List)
	for _, e := range files {
		if checkFilename(ctx, e) {
			l.AddFromFile(e)
		} else if strings.HasPrefix(e, "http:") {
			l.Add(NewURL(e))
		}
	}
	return l
}

func (l *List) Add(s Sourcer) *List {
	l.s = append(l.s, s)
	return l
}

func (l *List) AddFromFile(fn string) (*List, error) {
	var (
		base string
		err  error
	)

	if _, err := os.Stat(fn); err != nil {
		return l, errors.Wrapf(err, "unknown fn %s", fn)
	}

	base = fn

	// Special case for .zip.asc
	if strings.HasSuffix(base, ".zip.asc") || strings.HasSuffix(base, ".zip.gpg") {
		rbase, err := extractZipFrom(fn)
		if err != nil {
			return l, errors.Wrap(err, "extractzip")
		}
		base, err = filepath.Abs(rbase)
		if err != nil {
			return l, errors.Wrap(err, "basename")
		}
	}

	debug("opening %s", base)

	buf, err := readFile(base)
	if err != nil {
		return l, errors.Wrap(err, "single/readfile")
	}

	return l.ReadFromCSV(buf), nil
}

func (l *List) ReadFromCSV(r io.Reader) *List {
	return &List{}
}

func (l *List) Check() *Results {
	r := NewResults()
	for _, e := range l.s {
		if e.Check() {
			e.AddTo(r)
		}
	}
	return r
}
