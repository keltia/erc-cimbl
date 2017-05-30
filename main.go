package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"regexp"
)

var (
	// MyName is the application
	MyName    = "erc-cimbl"
	// MyVersion is our version
	MyVersion = "0.0.1"

	fVerbose bool
	fNoURLs  bool
	fNoPaths bool
	fDoMail  bool
)

// Context is the way to share info across functions.
type Context struct {
	config *Config
	Paths  map[string]bool
	URLs   map[string]string
	Client *http.Client
    files  []string
    proxyauth string
}

func init() {
	flag.BoolVar(&fDoMail, "M", false, "Send mail")
	flag.BoolVar(&fNoPaths, "P", false, "Do not check filenames")
	flag.BoolVar(&fNoURLs, "U", false, "Do not check URLs")
	flag.BoolVar(&fVerbose, "v", false, "Verbose mode")
}

func checkFilename(file string) (ok bool) {
	re := regexp.MustCompile(`CIMBL-\d+-CERTS\.csv`)

	return re.MatchString(file)
}

func main() {
	var config *Config

	// Parse CLI
	flag.Parse()

	if fVerbose {
		log.Printf("%s/%s", MyName, MyVersion)
	}

	if (fNoURLs && fNoPaths) || flag.NArg() == 0 {
		log.Println("Nothing to do!")
		os.Exit(0)
	}

	// No config file is not an error but you do not get to send mail
	config, err := loadConfig()
	if err != nil {
		log.Println("no config file, mail is disabled.")
		fDoMail = false
	}

	// No mail server configured but the rest is valid.
	if config.Server == "" {
		log.Println("no mail server, mail is disabled.")
		fDoMail = false
	}

	ctx := &Context{
		config: config,
		Paths:  map[string]bool{},
		URLs:   map[string]string{},
	}

	err = loadDbrc(ctx, dbrcFile)
	if err != nil {
		log.Println("No dbrc file, no proxy auth.")
	}
	// For all csv files on the CLI
	for _, file := range flag.Args() {
		if checkFilename(file) {
			if fVerbose {
				log.Printf("Checking %s…\n", file)
			}
			err := handleCSV(ctx, file)
			if err != nil {
				log.Printf("error reading %s: %v", file, err)
			}
			ctx.files = append(ctx.files, file)
		} else {
			if fVerbose {
				log.Printf("Ignoring %s…", file)
			}
		}
	}

	// Do something with the results
	err = doSendMail(ctx)
	if err != nil {
		log.Fatalf("sending mail: %v", err)
	}
}
