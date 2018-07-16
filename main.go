package main

import (
	"flag"
	"github.com/keltia/proxy"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	// MyName is the application
	MyName = "erc-cimbl"
	// MyVersion is our version
	MyVersion = "0.5.1"

	fDebug   bool
	fDoMail  bool
	fVerbose bool
	fNoURLs  bool
	fNoPaths bool
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
	flag.BoolVar(&fDebug, "D", false, "Debug mode")
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

func setup() *Context {
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

	proxyauth, err := proxy.SetupProxyAuth()
	if err != nil {
		log.Printf("No dbrc file, no proxy auth.: %v", err)
	} else {
		verbose("Using %s as proxy…", os.Getenv("http_proxy"))
		debug("Got %s as proxyauth", proxyauth)
		ctx.proxyauth = proxyauth
	}
	return ctx
}

func main() {
	// Parse CLI
	flag.Parse()

	if fDebug {
		fVerbose = true
	}

	verbose("%s/%s", MyName, MyVersion)

	if (fNoURLs && fNoPaths) || flag.NArg() == 0 {
		log.Println("Nothing to do!")
		os.Exit(0)
	}

	ctx := setup()

	ctx.tempdir = createSandbox(MyName)
	defer cleanupTemp(ctx.tempdir)

	// For all files on the CLI
	for _, file := range flag.Args() {
		if checkFilename(file) {
			verbose("Checking %s…\n", file)
			if err := handleSingleFile(ctx, file); err != nil {
				log.Printf("error reading %s: %v", file, err)
			}
			ctx.files = append(ctx.files, filepath.Base(file))
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
	if err := doSendMail(ctx); err != nil {
		log.Fatalf("sending mail: %v", err)
	}

	if len(skipped) != 0 {
		log.Printf("\nSkipped URLs:\n%s", strings.Join(skipped, "\n"))
	}
}
