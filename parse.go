package main

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/maxim2266/csvplus"
	"github.com/pkg/errors"
)

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

// These functions assume they are in the sandbox

// openFile looks at the file and give it to openZipfile() if needed
func openFile(ctx *Context, file string) (r io.ReadCloser, err error) {
	myfile := file

	debug("file is %s", file)
	_, err = os.Stat(file)
	if err != nil {
		return nil, errors.Wrap(err, "stat")
	}

	// Decrypt if needed
	if path.Ext(file) == ".asc" ||
		path.Ext(file) == ".ASC" {
		verbose("found encrypted file %s", file)
		myfile, err = decryptFile(ctx, file)
		if err != nil {
			return nil, errors.Wrapf(err, "decryptFile(%s)", file)
		}
	} else {
		verbose("found plain file %s", file)
	}

	// Next pass, check for zip file
	if path.Ext(myfile) == ".zip" ||
		path.Ext(myfile) == ".ZIP" {

		verbose("found zip file %s", myfile)

		myfile, err = openZipfile(ctx, myfile)
		if err != nil {
			return nil, errors.Wrap(err, "ENOZIP")
		}
	}
	return os.Open(myfile)
}

// readCSV reads the first csv in the zip file and copy into a temp file
func readCSV(ctx *Context, fn *zip.File) (file string) {
	sandbox := ctx.tempdir

	if fn == nil {
		return ""
	}

	verbose("found %s", fn.Name)
	// Open the CSV stream
	fh, err := fn.Open()
	if err != nil {
		log.Fatalf("unable to extract %s", fn.Name)
	}

	myfile := filepath.Join(sandbox.Cwd(), fn.Name)
	// Create our temp file
	ours, err := os.Create(myfile)
	if err != nil {
		log.Fatalf("unable to create %s in %s: %v", fn.Name, sandbox, err)
	}
	defer ours.Close()

	debug("created our tempfile %s", myfile)

	// copy all the bits over
	_, err = io.Copy(ours, fh)
	if err != nil {
		log.Fatalf("unable to write %s in %s: %v", fn.Name, sandbox, err)
	}
	file = filepath.Join(sandbox.Cwd(), fn.Name)
	return
}

// openZipfile extracts the first csv file out of he given zip.
func openZipfile(ctx *Context, file string) (string, error) {

	zfh, err := zip.OpenReader(file)
	if err != nil {
		return "", errors.Wrapf(err, "error opening %s", file)
	}
	defer zfh.Close()

	verbose("exploring %s", file)

	for _, fn := range zfh.File {
		verbose("looking at %s", fn.Name)

		if path.Ext(fn.Name) == ".csv" ||
			path.Ext(fn.Name) == ".CSV" {

			file = readCSV(ctx, fn)
			break
		}
	}
	if file == "" {
		return "", fmt.Errorf("no csv or unreadable")
	}
	return file, nil
}

// handleSingleFile creates a tempdir and dispatch csv/zip files to handler.
func handleSingleFile(ctx *Context, file string) (err error) {
	// Look at the file and whatever might be inside (and decrypt/unzip/…)
	r, err := openFile(ctx, file)
	if err != nil {
		return errors.Wrap(err, "openFile")
	}

	allLines := csvplus.FromReader(r).SelectColumns("type", "value")
	rows, err := csvplus.Take(allLines).
		Filter(csvplus.Any(csvplus.Like(csvplus.Row{"type": "url"}),
			csvplus.Like(csvplus.Row{"type": "filename"}),
			csvplus.Like(csvplus.Row{"type": "filename|sha1"}))).
		ToRows()
	if err != nil {
		return errors.Wrapf(err, "reading from %s", file)
	}

	for _, row := range rows {
		debug("row=%v", row)
		rt := strings.Split(row["type"], "|")[0]
		debug("rt=%s", rt)
		switch rt {
		case "filename":
			if !fNoPaths {
				handlePath(ctx, entryToPath(row["value"]))
			}
		case "url":
			if !fNoURLs {
				err = handleURL(ctx, row["value"])
				if err != nil {
					log.Printf("error(%s): %s", row["value"], err.Error())
				}
			}
		}
	}
	return nil
}

// handleAllFiles processes a list of files
func handleAllFiles(ctx *Context, files []string) error {
	// For all files on the CLI
	for _, file := range files {
		if checkFilename(file) {
			verbose("Checking %s…\n", file)

			nfile, _ := filepath.Abs(file)
			err := ctx.tempdir.Run(func() error {
				var err error

				if err = handleSingleFile(ctx, nfile); err != nil {
					log.Printf("error reading %s: %v", nfile, err)
					return err
				}
				ctx.files = append(ctx.files, filepath.Base(nfile))
				return nil
			})
			if err != nil {
				log.Printf("got error %v for %s", err, file)
			}
		} else {
			if strings.HasPrefix(file, "http:") {
				if !fNoURLs {
					err := handleURL(ctx, file)
					if err != nil {
						log.Printf("error checking %s: %v", file, err)
						continue
					}
				}
			} else {
				verbose("Ignoring %s…", file)
			}
		}
	}
	return nil
}
