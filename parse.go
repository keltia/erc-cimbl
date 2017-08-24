package main

import (
	"encoding/csv"
	"os"
	"strings"
	"archive/zip"
	"log"
	"path"
	"io/ioutil"
	"path/filepath"
	"io"
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
	var fn string

	_, err = os.Stat(file)
	if err != nil {
		return
	}

	if fVerbose {
		log.Printf("found %s", file)
	}

	if path.Ext(file) == ".zip" ||
		path.Ext(file) == ".ZIP" {

		if fVerbose {
			log.Printf("found zip file %s", file)
			log.Printf("extracting to %s", ctx.tempdir)
		}

		fn = openZipfile(ctx, file)
	}
	fh, err = os.Open(fn)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	return
}

// openZipfile extracts the first csv file out of he given zip.
func openZipfile(ctx *Context, file string) (fname string) {

	dir := ctx.tempdir
	if fVerbose {
		log.Printf("extracting to %s", dir)
	}

	fullFile, err := filepath.Abs(file)
	if err != nil {
		log.Fatalf("unable to get full path of %s", file)
	}

	// Go on
	err = os.Chdir(dir)
	if err != nil {
		log.Fatalf("unable to use tempdir %s: %v", dir, err)
	}

	zfh, err := zip.OpenReader(fullFile)
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
			break
		}
	}
	fname = file
	return
}

// handleSingleFile creates a tempdir and dispatch csv/zip files to handler.
func handleSingleFile(ctx *Context, file string) (err error) {
	// Extract in safe location
	dir, err := ioutil.TempDir("", "erc-cimbl")
	if err != nil {
		log.Fatalf("unable to create sandbox %s: %v", dir, err)
	}
	defer cleanupTemp(dir)

	ctx.tempdir = dir
	fh, err := openFile(ctx, file)
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
