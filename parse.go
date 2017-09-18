package main

import (
	"archive/zip"
	"encoding/csv"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
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

// cleanupTemp removes the temporary directory
func cleanupTemp(dir string) {
	err := os.RemoveAll(dir)
	if err != nil {
		log.Printf("cleanup failed for %s: %v", dir, err)
	}
}

// openFile looks at the file and give it to openZipfile() if needed
func openFile(ctx *Context, file string) (fh *os.File, err error) {
	var myfile string

	_, err = os.Stat(file)
	if err != nil {
		return
	}

	if fVerbose {
		log.Printf("found %s", file)
	}

	myfile = file
	// Decrypt if needed
	if path.Ext(file) == ".asc" ||
		path.Ext(file) == ".ASC" {
		if fVerbose {
			log.Printf("found encrypted file %s", file)
		}
		myfile, err = decryptFile(ctx, file)
		if err != nil {
			log.Fatalf("error decrypting %s: %v", file, err)
		}
	}

	// Next pass, check for zip file
	if path.Ext(myfile) == ".zip" ||
		path.Ext(myfile) == ".ZIP" {

		if fVerbose {
			log.Printf("found zip file %s", myfile)
		}

		myfile = openZipfile(ctx, myfile)
	}

	fh, err = os.Open(myfile)
	if err != nil {
		log.Fatalf("error opening %s: %v", myfile, err)
	}
	return
}

// decryptFiles returns the path name of the decrypted file
func decryptFile(ctx *Context, file string) (string, error) {
	dir := ctx.tempdir
	if fVerbose {
		log.Printf("Sandbox is %s", dir)
	}

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

	if fVerbose {
		log.Printf("Decrypting %s as %s", file, plainfile)
	}

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
	if fVerbose {
		log.Printf("found %s", fn.Name)
	}

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

	if fVerbose {
		log.Printf("created our tempfile %s", filepath.Join(dir, fn.Name))
	}

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
	if fVerbose {
		log.Printf("extracting to %s", dir)
	}

	// Go on
	if err := os.Chdir(dir); err != nil {
		log.Fatalf("unable to use tempdir %s: %v", dir, err)
	}

	zfh, err := zip.OpenReader(file)
	if err != nil {
		log.Fatalf("error opening %s: %v", file, err)
	}
	defer zfh.Close()

	if fVerbose {
		log.Printf("exploring %s", file)
	}

	for _, fn := range zfh.File {
		if fVerbose {
			log.Printf("looking at %s", fn.Name)
		}

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

	// Extract in safe location
	dir, err := ioutil.TempDir("", "erc-cimbl")
	if err != nil {
		log.Fatalf("unable to create sandbox %s: %v", dir, err)
	}
	defer cleanupTemp(dir)

	// We want the full path
	if myfile, err = filepath.Abs(file); err != nil {
		log.Fatalf("error checking %s in %s", myfile)
	}

	ctx.tempdir = dir
	fh, err := openFile(ctx, myfile)
	if err != nil {
		return
	}
	defer fh.Close()

	all := csv.NewReader(fh)
	allLines, err := all.ReadAll()

	for _, line := range allLines {
		// type at index 2
		// value at index 5
		vtype := line[2]
		etype := strings.Split(vtype, "|")

		switch etype[0] {
		case "filename":
			if !fNoPaths {
				handlePath(ctx, entryToPath(line[5]))
			}
		case "url":
			if !fNoURLs {
				handleURL(ctx, line[5])
			}
		}
	}
	return nil
}
