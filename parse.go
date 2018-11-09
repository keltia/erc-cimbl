package main

import (
	"bytes"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/keltia/archive"
	"github.com/maxim2266/csvplus"
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

// handleSingleFile creates a tempdir and dispatch csv/zip files to handler.
func handleSingleFile(ctx *Context, file string) (*Results, error) {
	var (
		base string
		err  error
	)

	if _, err := os.Stat(file); err != nil {
		return &Results{}, errors.Wrapf(err, "unknown file %s", file)
	}

	ext := strings.ToLower(filepath.Ext(file))
	base = file

	// Special case for .zip.asc
	if ext == ".zip.asc" {
		rbase, err := extractZipFrom(file)
		if err != nil {
			return &Results{}, errors.Wrap(err, "extractzip")
		}
		base, err = filepath.Abs(rbase)
		if err != nil {
			return &Results{}, errors.Wrap(err, "basename")
		}
	}

	debug("opening %s", base)

	buf, err := readFile(base)
	if err != nil {
		return &Results{}, errors.Wrap(err, "single/readfile")
	}

	return handleCSV(ctx, buf)
}

// handleCSV decodes the CSV file
func handleCSV(ctx *Context, r io.Reader) (*Results, error) {
	res := NewResults()

	allLines := csvplus.FromReader(r).SelectColumns("type", "value")
	rows, err := csvplus.Take(allLines).
		Filter(csvplus.Any(csvplus.Like(csvplus.Row{"type": "url"}),
			csvplus.Like(csvplus.Row{"type": "filename"}),
			csvplus.Like(csvplus.Row{"type": "filename|sha1"}))).
		ToRows()
	if err != nil {
		return res, errors.Wrapf(err, "reading csv")
	}

	for _, row := range rows {
		debug("row=%v", row)
		rt := strings.Split(row["type"], "|")[0]
		debug("rt=%s", rt)
		switch rt {
		case "filename":
			if !fNoPaths {
				p := handlePath(ctx, entryToPath(row["value"]))
				if p != "" {
					res.Add("path", p)
				}
			}
		case "url":
			if !fNoURLs {
				u, err := handleURL(ctx, row["value"])
				if err != nil {
					log.Printf("error(%s): %s", row["value"], err.Error())
					continue
				}
				res.Add("url", u)

			}
		}
	}
	return res, err
}

// handleAllFiles processes a list of files
func handleAllFiles(ctx *Context, files []string) (*Results, error) {
	// For all files on the CLI
	res := NewResults()
	for _, file := range files {
		if checkFilename(file) {
			verbose("Checking %s…\n", file)

			nfile, _ := filepath.Abs(file)
			err := ctx.tempdir.Run(func() error {
				var err error

				r, err := handleSingleFile(ctx, nfile)
				if err != nil {
					log.Printf("error reading %s: %v", nfile, err)
					return err
				}
				res.Merge(r)
				ctx.files = append(ctx.files, filepath.Base(nfile))
				return nil
			})
			if err != nil {
				log.Printf("got error %v for %s", err, file)
				continue
			}
		} else {
			// Do we have an incomplete mail attachement disguised into a OpenPGP message?
			if checkOpenPGP(file) {
				verbose("New encrypted multipart format %s…\n", file)

				nfile, _ := filepath.Abs(file)
				err := ctx.tempdir.Run(func() error {
					var err error

					r, err := handleMultipart(ctx, nfile)
					if err != nil {
						log.Printf("error reading %s: %v", nfile, err)
						return err
					}
					res.Merge(r)
					ctx.files = append(ctx.files, filepath.Base(nfile))
					return nil
				})
				if err != nil {
					log.Printf("got error %v for %s", err, file)
					continue
				}
			} else {
				if strings.HasPrefix(file, "http:") {
					if !fNoURLs {
						u, err := handleURL(ctx, file)
						if err != nil {
							log.Printf("error checking %s: %v", file, err)
							continue
						}
						res.Add("url", u)
					}
				} else {
					verbose("Ignoring %s…", file)
				}
			}
		}
	}
	return res, nil
}
