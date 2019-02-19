package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/keltia/archive"
	"github.com/keltia/proxy"
	"github.com/keltia/sandbox"
	"github.com/pkg/errors"
)

var (
	// MyName is the application
	MyName = "erc-cimbl"
	// MyVersion is our version
	MyVersion = "0.8.0"

	fDebug   bool
	fDoMail  bool
	fVerbose bool
	fNoURLs  bool
	fNoPaths bool
)

// Context is the way to share info across functions.
type Context struct {
	Client *http.Client

	config    *Config
	tempdir   *sandbox.Dir
	files     []string
	proxyauth string
	mail      MailSender
}

// Usage string override.
var Usage = func() {
	fmt.Fprintf(os.Stderr, "%s/%s (Archive/%s Proxy/%s Sandbox/%s)\n\n",
		MyName, MyVersion, archive.Version(), proxy.Version(), sandbox.Version())

	flag.PrintDefaults()
}

func init() {
	flag.Usage = Usage

	flag.BoolVar(&fDebug, "D", false, "Debug mode")
	flag.BoolVar(&fDoMail, "M", false, "Send mail")
	flag.BoolVar(&fNoPaths, "P", false, "Do not check filenames")
	flag.BoolVar(&fNoURLs, "U", false, "Do not check URLs")
	flag.BoolVar(&fVerbose, "v", false, "Verbose mode")
}

func setup() (*Context, error) {
	if fDebug {
		fVerbose = true
	}

	verbose("%s/%s Archive/%s Proxy/%s Sandbox/%s",
		MyName, MyVersion, archive.Version(), proxy.Version(), sandbox.Version())

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
		mail:   SMTPMailSender{},
	}

	proxyauth, err := proxy.SetupProxyAuth()
	if err != nil {
		verbose("No proxy auth.: %v", err)
	} else {
		verbose("Found http_proxy variable")
		debug("Using %s as proxy…", os.Getenv("http_proxy"))
		debug("Got %s as proxyauth", proxyauth)
		ctx.proxyauth = proxyauth
	}

	// Create our sendbox
	ctx.tempdir, err = sandbox.New(MyName)
	if err != nil {
		return nil, errors.Wrap(err, "setup")
	}

	return ctx, nil
}

func realmain(args []string) error {
	ctx, err := setup()
	if err != nil {
		return errors.Wrap(err, "realmain")
	}

	defer ctx.tempdir.Cleanup()

	if (fNoURLs && fNoPaths) || len(args) == 0 {
		log.Println("Nothing to do!")
		return nil
	}

	res, err := handleAllFiles(ctx, args)
	if err != nil {
		return errors.Wrap(err, "error processing files")
	}

	verbose("res=%v", res)

	// Do something with the results
	if err := doSendMail(ctx, res); err != nil {
		return errors.Wrap(err, "sending mail")
	}

	if len(skipped) != 0 {
		log.Printf("\nSkipped URLs:\n%s", strings.Join(skipped, "\n"))
	}

	return nil
}

func main() {
	// Parse CLI
	flag.Parse()

	if err := realmain(flag.Args()); err != nil {
		log.Fatalf("Error %v\n", err)
	}
}
