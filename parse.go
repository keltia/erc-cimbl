package main

import (
	"bytes"
	"os"
	"path/filepath"

	"github.com/keltia/archive"
	"github.com/pkg/errors"
)

// These functions assume they are in the sandbox

// Given an asc/gpg file, create a temp file with uncrypted content
// Assumes it is inside a sandbox
func extractZipFrom(file string) (string, error) {
	debug("reading %s", file)

	// Process the file (gpg encrypted zip file)
	a, err := archive.New(file)
	if err != nil {
		return "", errors.Wrap(err, "archive/new(asc)")
	}

	unc, err := a.Extract(".zip")
	if err != nil {
		return "", errors.Wrap(err, "extract")
	}

	base := filepath.Base(file)

	debug("creating %s.zip", base+".zip")

	// Create a temp file
	zipfh, err := os.Create(base + ".zip")
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
	return base + ".zip", err
}

func readFile(base string) (*bytes.Buffer, error) {
	var buf bytes.Buffer

	debug("openzip %s", base)

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
	list.ctx = ctx

	r := list.Check(ctx)
	debug("r=%#v\n", r)

	return r, nil
}
