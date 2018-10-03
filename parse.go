package main

import (
	"archive/zip"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/maxim2266/csvplus"
	"github.com/pkg/errors"
	"github.com/proglottis/gpgme"
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

		myfile = openZipfile(ctx, myfile)
	}
	return os.Open(myfile)
}

// decryptFiles returns the path name of the decrypted file
func decryptFile(ctx *Context, file string) (string, error) {
	// Carefully open the box
	fh, err := os.Open(file)
	if err != nil {
		return "", errors.Wrap(err, "decryptFile/Open")
	}
	defer fh.Close()

	// Do the decryption thing
	plain, err := gpgme.Decrypt(fh)
	if err != nil {
		return "", errors.Wrap(err, "Decrypt")
	}
	defer plain.Close()

	// Save "plain" text
	base := filepath.Base(file)
	ext := filepath.Ext(base)
	zipname := strings.Replace(base, ext, "", 1)

	plainfile := filepath.Join(ctx.tempdir.Cwd(), zipname)

	verbose("Decrypting %s as %s", file, plainfile)

	dfh, err := os.Create(plainfile)
	if err != nil {
		return "", errors.Wrap(err, "decryptFile/Create")
	}
	defer dfh.Close()

	_, err = io.Copy(dfh, plain)
	if err != nil {
		return "", errors.Wrap(err, "decryptFile/Copy")
	}

	return plainfile, nil
}

// readCSV reads the first csv in the zip file and copy into a temp file
func readCSV(ctx *Context, fn *zip.File) (file string) {
	sandbox := ctx.tempdir
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
func openZipfile(ctx *Context, file string) (fname string) {

	zfh, err := zip.OpenReader(file)
	if err != nil {
		log.Fatalf("error opening %s: %v", file, err)
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
	fname = file
	return
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
		verbose("row=%v", row)
		rt := strings.Split(row["type"], "|")[0]
		verbose("rt=%s", rt)
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
