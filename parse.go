package main

import (
	"archive/zip"
	"github.com/maxim2266/csvplus"
	"github.com/proglottis/gpgme"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
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

// openFile looks at the file and give it to openZipfile() if needed
func openFile(ctx *Context, file string) (fn string, err error) {
	var myfile string

	_, err = os.Stat(file)
	if err != nil {
		return
	}

	myfile = file
	// Decrypt if needed
	if path.Ext(file) == ".asc" ||
		path.Ext(file) == ".ASC" {
		verbose("found encrypted file %s", file)
		myfile, err = decryptFile(ctx, file)
		if err != nil {
			log.Fatalf("error decrypting %s: %v", file, err)
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
	fn = myfile
	return
}

// decryptFiles returns the path name of the decrypted file
func decryptFile(ctx *Context, file string) (string, error) {
	dir := ctx.tempdir
	verbose("Sandbox is %s", dir)

	// Insure we got the full path
	file, _ = filepath.Abs(file)

	// Go into the sandbox
	err := os.Chdir(dir)
	if err != nil {
		log.Fatalf("unable to use tempdir %s: %v", dir, err)
	}

	// Carefully open the box
	fh, err := os.Open(file)
	if err != nil {
		return "", err
	}
	defer fh.Close()

	// Do the decryption thing
	plain, err := gpgme.Decrypt(fh)
	if err != nil {
		return "", err
	}
	defer plain.Close()

	// Save "plain" text
	base := filepath.Base(file)
	ext := filepath.Ext(base)
	zipname := strings.Replace(base, ext, "", 1)

	plainfile := filepath.Join(dir, zipname)

	verbose("Decrypting %s as %s", file, plainfile)

	dfh, err := os.Create(plainfile)
	if err != nil {
		return "", err
	}

	_, err = io.Copy(dfh, plain)
	if err != nil {
		return "", err
	}
	dfh.Close()

	return plainfile, nil
}

// readCSV reads the first csv in the zip file and copy into a temp file
func readCSV(ctx *Context, fn *zip.File) (file string) {
	dir := ctx.tempdir
	verbose("found %s", fn.Name)

	// Open the CSV stream
	fh, err := fn.Open()
	if err != nil {
		log.Fatalf("unable to extract %s", fn.Name)
	}

	// Create our temp file
	ours, err := os.Create(filepath.Join(dir, fn.Name))
	if err != nil {
		log.Fatalf("unable to create %s in %s: %v", fn.Name, dir, err)
	}
	defer ours.Close()

	verbose("created our tempfile %s", filepath.Join(dir, fn.Name))

	// copy all the bits over
	_, err = io.Copy(ours, fh)
	if err != nil {
		log.Fatalf("unable to write %s in %s: %v", fn.Name, dir, err)
	}
	file = filepath.Join(dir, fn.Name)
	return
}

// openZipfile extracts the first csv file out of he given zip.
func openZipfile(ctx *Context, file string) (fname string) {

	dir := ctx.tempdir

	// Go on
	if err := os.Chdir(dir); err != nil {
		log.Fatalf("unable to use tempdir %s: %v", dir, err)
	}

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
	var myfile string

	// We want the full path
	if myfile, err = filepath.Abs(file); err != nil {
		log.Fatalf("error checking %s in %s", myfile)
	}

	// Look at the file and whatever might be inside (and decrypt/unzip/…)
	myfile, err = openFile(ctx, myfile)
	if err != nil {
		return
	}

	allLines := csvplus.FromFile(myfile).SelectColumns("type", "value")
	rows, err := csvplus.Take(allLines).
		Filter(csvplus.Any(csvplus.Like(csvplus.Row{"type": "url"}),
			csvplus.Like(csvplus.Row{"type": "filename"}))).
		ToRows()

	for _, row := range rows {
		switch row["type"] {
		case "filename":
			if !fNoPaths {
				handlePath(ctx, entryToPath(row["value"]))
			}
		case "url":
			if !fNoURLs {
				handleURL(ctx, row["value"])
			}
		}
	}
	return nil
}
