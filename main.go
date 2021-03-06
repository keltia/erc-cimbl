package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"runtime"
	"runtime/pprof"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/keltia/archive"
	"github.com/keltia/sandbox"
	"github.com/pkg/errors"
)

const (
	REfn = `(?i:CIMBL-\d+-(CERTS|EU)\.(csv|zip)(\.asc|))`
)

var (
	// MyName is the application
	MyName = "erc-cimbl"
	// MyVersion is our version, add our features
	MyVersion = "0.11.0,parallel,resty"

	fNoCleanup bool
	fDebug     bool
	fDoMail    bool
	fVerbose   bool
	fNoURLs    bool
	fNoPaths   bool
	fProfile   bool
	fSkipped   bool
	fJobs      int

	// RE to check filenames — sensible default
	REFile *regexp.Regexp = regexp.MustCompile(REfn)
)

// Context is the way to share info across functions.
type Context struct {
	Client *resty.Client

	config  *Config
	tempdir *sandbox.Dir
	mail    MailSender
	jobs    int
}

// Usage string override.
var Usage = func() {
	fmt.Fprintf(os.Stderr, "%s/%s (Archive/%s Sandbox/%s)\n\n",
		MyName, MyVersion, archive.Version(), sandbox.Version())

	flag.PrintDefaults()
}

func init() {
	flag.Usage = Usage

	flag.BoolVar(&fNoCleanup, "C", false, "No cleanup for temp files.")
	flag.BoolVar(&fDebug, "D", false, "Debug mode")
	flag.BoolVar(&fDoMail, "M", false, "Send mail")
	flag.BoolVar(&fNoPaths, "P", false, "Do not check filenames")
	flag.BoolVar(&fSkipped, "S", false, "Display skipped URLs")
	flag.BoolVar(&fNoURLs, "U", false, "Do not check URLs")
	flag.IntVar(&fJobs, "j", runtime.NumCPU(), "parallel jobs")
	flag.BoolVar(&fVerbose, "v", false, "Verbose mode")
	flag.BoolVar(&fProfile, "prof", false, "Profiling")
}

func setup() (*Context, error) {
	if fDebug {
		fVerbose = true
	}

	verbose("%s/%s Archive/%s Sandbox/%s",
		MyName, MyVersion, archive.Version(), sandbox.Version())

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

	if config.REFile != "" {
		REFile = regexp.MustCompile(config.REFile)
	}

	ctx := &Context{
		config: config,
		mail:   SMTPMailSender{},
		jobs:   fJobs,
	}

	if fProfile {
		f, _ := os.Create("cpu.prof")
		if err = pprof.StartCPUProfile(f); err != nil {
			log.Fatalf("cant profile")
		}
	}
	proxy := os.Getenv("http_proxy")
	c := resty.New().SetProxy(proxy).SetTimeout(10 * time.Second)
	ctx.Client = c

	if proxy == "" {
		verbose("No proxy auth.: %v", err)
	} else {
		verbose("Found http_proxy variable")
		debug("Using %s as proxy…", os.Getenv("http_proxy"))
	}

	// Create our sandbox
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
	if fProfile {
		defer pprof.StopCPUProfile()
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

	if fSkipped {
		if len(skipped) != 0 {
			log.Printf("\nSkipped URLs:\n%s", strings.Join(skipped, "\n"))
		}
	}

	if !fNoCleanup {
		for _, fn := range res.files {
			if err := os.Remove(fn); err != nil {
				log.Printf("Can not delete %s: %v", fn, err)
			}
		}
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
