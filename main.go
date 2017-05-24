package main

import (
	"flag"
	"log"
	"os"
	"regexp"
)

var (
	MyName = "erc-cimbl"

	fVerbose bool
	fNoURLs  bool
	fNoPaths bool
)

func init() {
	flag.BoolVar(&fNoPaths, "P", false, "Do not check filenames")
	flag.BoolVar(&fNoURLs, "U", false, "Do not check URLs")
	flag.BoolVar(&fVerbose, "v", false, "Verbose mode")
}

func checkFilename(file string) (ok bool) {
	re := regexp.MustCompile(`CIMBL-\d+-CERTS\.csv`)

	return re.MatchString(file)
}

func main() {

	// Parse CLI
	flag.Parse()

	if (fNoURLs && fNoPaths) || flag.NArg() == 0 {
		log.Println("Nothing to do!")
		os.Exit(1)
	}

	if fVerbose {
		log.Printf("%s\n%s", MyName)
	}

	for _, file := range flag.Args() {
		if checkFilename(file) {
			handleCSV(file)
		}
	}
}
