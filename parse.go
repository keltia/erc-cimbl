package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/keltia/archive"
	"github.com/pkg/errors"
)

// These functions assume they are in the sandbox

// RemoveExt removes the extension of the file, used for .zip.asc
func RemoveExt(fn string) string {
	ext := filepath.Ext(fn)
	all := strings.Split(fn, ".")
	if len(all) <= 1 {
		return fn
	}
	if "."+all[len(all)-1] == ext {
		return strings.Join(all[0:len(all)-1], ".")
	}
	return ""
}

// Given an asc/gpg file, create a temp file with uncrypted content
// Assumes it is inside a sandbox
func extractZipFrom(file string) (string, error) {
	debug("extractZipFile %s", file)

	// Process the file (gpg encrypted zip file)
	a, err := archive.New(file)
	if err != nil {
		return "", errors.Wrap(err, "archive/new(asc)")
	}

	unc, err := a.Extract(".zip")
	if err != nil {
		return "", errors.Wrap(err, "extract")
	}

	base := RemoveExt(filepath.Base(file))

	debug("creating %s", base)

	// Create a temp file
	zipfh, err := os.Create(base)
	if err != nil {
		return "", errors.Wrap(err, "create/temp")
	}

	n, err := zipfh.Write(unc)
	if err != nil {
		return "", errors.Wrap(err, "buffer/write")
	}

	if n != len(unc) {
		return "", errors.Wrap(err, "short read")
	}
	err = zipfh.Close()
	return base, err
}

// Read the actual file out of a possible archive (zip, gpg, etc.)
func readFile(base string) (*bytes.Buffer, error) {
	var buf bytes.Buffer

	debug("readFile %s", base)

	// Here buf is the decrypted arc or plain file
	arc, err := archive.New(base)
	if err != nil {
		return nil, errors.Wrap(err, "archive/new")
	}

	unc, err := arc.Extract(".csv")
	if err != nil {
		return nil, errors.Wrap(err, "extract(csv)")
	}

	n, err := buf.Write(unc)
	if err != nil {
		return nil, errors.Wrap(err, "buffer/write")
	}

	if n != buf.Len() {
		return nil, errors.Wrap(err, "short read")
	}
	return &buf, nil
}

func readIPlist(base string) (*bytes.Buffer, error) {
	var buf bytes.Buffer

	debug("readFile %s", base)

	unc, err := ioutil.ReadFile(base)
	if err != nil {
		return nil, errors.Wrapf(err, "ioutil/read")
	}
	n, err := buf.Write(unc)
	if err != nil {
		return nil, errors.Wrap(err, "buffer/write")
	}

	if n != buf.Len() {
		return nil, errors.Wrap(err, "short read")
	}
	return &buf, nil
}

/*
Fields in the CSV file:

observable_uuid,
kill_chain,
type,
time_start,
time_end,
value,
to_ids,
blacklist,
malware_research,
vuln_mgt,
indicator_uuid,
indicator_detect_time,
indicator_threat_type,
indicator_threat_level,
indicator_targeted_domain,
indicator_start_time,
indicator_end_time,
indicator_title

We filter on "type", looking for "url" & "filename".

*/

// handleAllFiles processes a list of files
func handleAllFiles(ctx *Context, files []string) (*Results, error) {
	// For all files on the CLI
	//res := NewResults()

	list := NewList(files)
	debug("list=%#v\n", list)

	if list.Length() != 0 {
		t1 := time.Now()
		r := list.Check(ctx)
		t2 := time.Since(t1)
		verbose("time=%v", t2)
		debug("r(main)=%#v\n", r)
		return r, nil
	}
	log.Printf("Empty list.")
	return NewResults(), nil
}
