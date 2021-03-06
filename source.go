package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/go-resty/resty/v2"
	"github.com/maxim2266/csvplus"
	"github.com/pkg/errors"
)

// -----

type Sourcer interface {
	Check(req *resty.Client) bool
	AddTo(r *Results)
}

type URL struct {
	H string
}

func NewURL(u string) *URL {
	return &URL{H: u}
}

// XXX
func (u *URL) Check(c *resty.Client) bool {
	r, _ := handleURL(c, u.H)
	if r == u.H {
		return true
	}
	return false
}

func (u *URL) AddTo(r *Results) {
	verbose("U")
	r.Add("url", u.H)
}

type Filename struct {
	Name string
}

func NewFilename(s string) *Filename {
	return &Filename{Name: s}
}

func (f *Filename) Check(c *resty.Client) bool {
	return true
}

func (f *Filename) AddTo(r *Results) {
	verbose("F")
	r.Add("filename", f.Name)
}

// -----

type List struct {
	s     []Sourcer
	files []string
}

// NewList create a new list from sources, either URL or a CIMBL filename
// NewList() with another filename is not supported although an URL is.
func NewList(files []string) *List {
	if files == nil || len(files) == 0 {
		return &List{}
	}

	l := &List{}

	for _, e := range files {
		if strings.HasPrefix(e, "http:") {
			l.Add(NewURL(e))
		} else if strings.HasSuffix(e, ".txt") {
			if l, err := l.AddFromIP(e); err != nil {
				log.Printf("Invalid IP list: %s", e)
				return l
			}
		} else if REFile.MatchString(e) {
			var err error

			l, err = l.AddFromFile(e)
			if err != nil {
				log.Printf("%v: reading error: %v", e, errors.Wrap(err, "AddFromFile"))
			}
		} else {
			log.Printf("invalid filename")
		}
	}
	return l
}

func (l *List) Add(s Sourcer) *List {
	debug("adding %#v", s)
	l.s = append(l.s, s)
	return l
}

func (l *List) Length() int {
	return len(l.s)
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

	l.files = append(l.files, filepath.Base(base))
	return l.ReadFromCSV(buf)
}

func (l *List) AddFromIP(fn string) (*List, error) {
	if _, err := os.Stat(fn); err != nil {
		return l, errors.Wrapf(err, "unknown fn %s", fn)
	}

	buf, err := readIPlist(fn)
	if err != nil {
		return nil, errors.Wrap(err, "addfromip")
	}

	s := bufio.NewScanner(buf)
	for s.Scan() {
		l.Add(NewURL(fmt.Sprintf("http://%s/", s.Text())))
	}
	return l, nil
}

func (l *List) ReadFromCSV(r io.Reader) (*List, error) {
	allLines := csvplus.FromReader(r).SelectColumns("type", "value", "to_ids")
	rows, err := csvplus.Take(allLines).
		Filter(csvplus.Any(csvplus.Like(csvplus.Row{"type": "url"}),
			csvplus.Like(csvplus.Row{"type": "filename"}),
			csvplus.Like(csvplus.Row{"type": "filename|sha1"}))).
		ToRows()
	if err != nil {
		return l, errors.Wrapf(err, "reading csv")
	}

	verbose("csv/%d entries found.", len(rows))

	for _, row := range rows {
		debug("row=%v", row)
		rt := strings.Split(row["type"], "|")[0]
		debug("rt=%s", rt)
		switch rt {
		case "filename":
			fn := strings.Split(row["value"], "|")[0]
			l.Add(NewFilename(fn))
		case "url":
			// if to_ids is set to 0, do not auto block.
			if row["to_ids"] == "1" {
				l.Add((NewURL(row["value"])))
			}
		}
	}

	return l, nil
}

func (l *List) Files() []string {
	return l.files
}

func (l *List) Merge(l1 *List) *List {
	for _, e := range l1.s {
		l.Add(e)
	}
	return l
}

// Check1 (now reversed), is the initial implementation with a mutex
func (l *List) Check1(ctx *Context) *Results {
	var mut sync.Mutex

	r := NewResults()

	wg := &sync.WaitGroup{}

	queue := make(chan Sourcer, len(l.s))

	debug("setup %d workers\n", ctx.jobs)

	// Setup workers
	for i := 0; i < ctx.jobs; i++ {
		wg.Add(1)

		go func(n int, wg *sync.WaitGroup) {
			defer wg.Done()

			debug("%d is fine\n", n)
			for e := range queue {
				verbose("w%d - %d left", n, len(queue))
				if e.Check(ctx.Client) {
					verbose("adding %#v\n", e)
					mut.Lock()
					e.AddTo(r)
					mut.Unlock()
				}
			}
		}(i, wg)
	}

	debug("scan queue:\n")
	for _, q := range l.s {
		queue <- q
	}

	close(queue)
	wg.Wait()
	r.files = l.Files()
	debug("r/check=%#v\n", r)
	return r
}

// gather results
func res(ins <-chan Sourcer) *Results {
	debug("processing results")
	r := NewResults()

	for e := range ins {
		fmt.Print(".")
		e.AddTo(r)
	}
	return r
}

// Check is an alternate version of Check where we still parallelize checks but serialize updating results
// instead of using a mutex.
func (l *List) Check(ctx *Context) *Results {

	wg := &sync.WaitGroup{}

	// Length for the 2nd one will be tuned
	queue := make(chan Sourcer, 50)

	ins := make(chan Sourcer, l.Length())

	done := make(chan struct{})
	defer close(done)

	debug("setup done")

	// Setup the end of the fan-out
	debug("setup %d workers\n", ctx.jobs)

	// Setup workers (fan-in)
	for i := 0; i < ctx.jobs; i++ {
		wg.Add(1)

		go func(n int, queue <-chan Sourcer, wg *sync.WaitGroup) {
			defer wg.Done()

			debug("%d is fine\n", n)
			for {
				e, ok := <-queue
				if !ok {
					return
				}

				debug("w%d - checking %v", n, e)
				if e.Check(ctx.Client) {
					debug("adding %#v\n", e)
					ins <- e
				}
			}
		}(i, queue, wg)
	}

	var result *Results

	// Feed the queue
	debug("scan queue:\n")
	for _, q := range l.s {
		queue <- q
	}

	go func() {
		wg.Wait()
		debug("after wait")
		close(ins)
	}()

	debug("closing")
	close(queue)

	//done <- struct{}{}
	debug("after close")

	result = res(ins)
	result.files = l.Files()
	debug("r/check=%#v\n", result)
	return result
}
