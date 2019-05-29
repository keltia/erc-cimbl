package main

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/maxim2266/csvplus"
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
	l := &List{s:[]Sourcer{}}
	if files == nil || len(files) == 0{
		return &List{}
	}
	for _, e := range files {
		if checkFilename(ctx, e) {
			var err error

			l, err = l.AddFromFile(e)
			if err != nil {
				log.Printf("%v: reading error", e)
			}
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

	return l.ReadFromCSV(buf)
}

func (l *List) ReadFromCSV(r io.Reader) (*List, error) {
	allLines := csvplus.FromReader(r).SelectColumns("type", "value")
	rows, err := csvplus.Take(allLines).
		Filter(csvplus.Any(csvplus.Like(csvplus.Row{"type": "url"}),
			csvplus.Like(csvplus.Row{"type": "filename"}),
			csvplus.Like(csvplus.Row{"type": "filename|sha1"}))).
		ToRows()
	if err != nil {
		return &List{}, errors.Wrapf(err, "reading csv")
	}

	verbose("%d entries found.", len(rows))

	for _, row := range rows {
		debug("row=%v", row)
		rt := strings.Split(row["type"], "|")[0]
		debug("rt=%s", rt)
		switch rt {
		case "filename":
			l.Add(NewFilename(row["value"]))
		case "url":
			l.Add((NewURL(row["value"])))
		default:
			continue
		}
	}

	return &List{}, nil
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
