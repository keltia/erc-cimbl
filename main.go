package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/keltia/proxy"
	"github.com/keltia/sandbox"
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
	tempdir   *sandbox.Dir
	Paths     map[string]bool
	URLs      map[string]bool
	files     []string
	Client    *http.Client
	proxyauth string
	mail      MailSender
	gpg       Decrypter
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

func setup() *Context {
	if fDebug {
		fVerbose = true
	}

	// No config file is not an error but you do not get to send mail
	config, err := loadConfig()
	if err != nil {
		verbose("no config file, mail is disabled.")
		fDoMail = false
	}

	// No mail server configured but the rest is valid.
	if config.Server == "" {
		verbose("no mail server, mail is disabled.")
		fDoMail = false
	} else {
		verbose("Got mail server %s…", config.Server)
	}

	ctx := &Context{
		config: config,
		Paths:  map[string]bool{},
		URLs:   map[string]bool{},
		mail:   SMTPMailSender{},
		gpg:    Gpgme{},
	}

	proxyauth, err := proxy.SetupProxyAuth()
	if err != nil {
		verbose("No proxy auth.: %v", err)
	} else {
		verbose("Using %s as proxy…", os.Getenv("http_proxy"))
		debug("Got %s as proxyauth", proxyauth)
		ctx.proxyauth = proxyauth
	}
	return ctx
}

func main() {
	var err error

	// Parse CLI
	flag.Parse()

	ctx := setup()

	verbose("%s/%s", MyName, MyVersion)

	if (fNoURLs && fNoPaths) || flag.NArg() == 0 {
		log.Println("Nothing to do!")
		os.Exit(0)
	}

	ctx.tempdir, err = sandbox.New(MyName)
	if err != nil {
		log.Fatalf("unable to create sandbox: %v", err)
	}
	defer ctx.tempdir.Cleanup()

	err = handleAllFiles(ctx, flag.Args())
	if err != nil {
		log.Fatalf("error processing files: %v", err)
	}

	// Do something with the results
	if err := doSendMail(ctx); err != nil {
		log.Fatalf("sending mail: %v", err)
	}

	if len(skipped) != 0 {
		log.Printf("\nSkipped URLs:\n%s", strings.Join(skipped, "\n"))
	}
}
