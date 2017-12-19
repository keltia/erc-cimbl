package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"io/ioutil"
)

var (
	// MyName is the application
	MyName = "erc-cimbl"
	// MyVersion is our version
	MyVersion = "0.4.2"

	fVerbose bool
	fNoURLs  bool
	fNoPaths bool
	fDoMail  bool
)

// Context is the way to share info across functions.
type Context struct {
	config    *Config
	tempdir   string
	Paths     map[string]bool
	URLs      map[string]string
	Client    *http.Client
	files     []string
	proxyauth string
}

func init() {
	flag.BoolVar(&fDoMail, "M", false, "Send mail")
	flag.BoolVar(&fNoPaths, "P", false, "Do not check filenames")
	flag.BoolVar(&fNoURLs, "U", false, "Do not check URLs")
	flag.BoolVar(&fVerbose, "v", false, "Verbose mode")
}

func checkFilename(file string) (ok bool) {
	re := regexp.MustCompile(`(?i:CIMBL-\d+-CERTS\.(csv|zip)(\.asc|))`)

	return re.MatchString(file)
}

// cleanupTemp removes the temporary directory
func cleanupTemp(dir string) {
	err := os.RemoveAll(dir)
	if err != nil {
		log.Printf("cleanup failed for %s: %v", dir, err)
	}
}

// createSandbox creates our our directory with TEMPDIR (wherever it is)
func createSandbox(tag string) (path string) {

	// Extract in safe location
	dir, err := ioutil.TempDir("", tag)
	if err != nil {
		log.Fatalf("unable to create sandbox %s: %v", dir, err)
	}
	return dir
}

func main() {
	var config *Config

	// Parse CLI
	flag.Parse()

	verbose("%s/%s", MyName, MyVersion)

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
	} else {
		verbose("Got mail server %s…", config.Server)
	}

	ctx := &Context{
		config: config,
		Paths:  map[string]bool{},
		URLs:   map[string]string{},
	}

	err = setupProxyAuth(ctx, dbrcFile)
	if err != nil {
		log.Println("No dbrc file, no proxy auth.")
	} else {
		verbose("Using %s as proxy…", os.Getenv("http_proxy"))
	}

	ctx.tempdir = createSandbox(MyName)
	defer cleanupTemp(ctx.tempdir)

	// For all files on the CLI
	for _, file := range flag.Args() {
		if checkFilename(file) {
			verbose("Checking %s…\n", file)
			err = handleSingleFile(ctx, file)
			if err != nil {
				log.Printf("error reading %s: %v", file, err)
			}
			ctx.files = append(ctx.files, file)
		} else {
			if strings.HasPrefix(file, "http:") {
				if !fNoURLs {
					handleURL(ctx, file)
				}
			} else {
				verbose("Ignoring %s…", file)
			}
		}
	}

	// Do something with the results
	err = doSendMail(ctx)
	if err != nil {
		log.Fatalf("sending mail: %v", err)
	}

	if len(skipped) != 0 {
		log.Printf("\nSkipped URLs:\n%s", strings.Join(skipped, "\n"))
	}
}
