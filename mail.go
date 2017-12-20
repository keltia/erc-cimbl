package main

import (
	"bytes"
	"fmt"
	"log"
	"net/smtp"
	"strings"
	"text/template"
)

var (
	mailTmpl = `Subject: {{.Subject}}
To: {{.To}}
Cc: {{.Cc}}
X-Contact-Info: {{.From}}

Dear Service Desk,

After reading the following files received from CERT-EU:
  {{.Files}}

{{.Paths}}
{{.URLs}}
Best regards,
--
Your friendly script - {{.MyName}}/{{.MyVersion}}
`

	pathsTmpl = "Please add the following to the list of blocked filenames:\n"
	urlsTmpl  = "Please add the following to the list of blocked URLs on BlueCoat:\n"

	skipped = []string{}
)

type mailVars struct {
	From      string
	To        string
	Cc        string
	Subject   string
	MyName    string
	MyVersion string
	URLs      string
	Paths     string
	Files     string
}

func createMail(ctx *Context) (str string, err error) {
	var txt bytes.Buffer

	vars := mailVars{
		From:      ctx.config.From,
		To:        ctx.config.To,
		Cc:        ctx.config.Cc,
		Subject:   ctx.config.Subject,
		MyName:    MyName,
		MyVersion: MyVersion,
		Files:     strings.Join(ctx.files, ", "),
		Paths:	   addPaths(ctx),
		URLs:      addURLs(ctx),
	}

	t := template.Must(template.New("mail").Parse(mailTmpl))
	err = t.Execute(&txt, vars)
	return txt.String(), err
}

func addPaths(ctx *Context) string {
	var txt string

	if !fNoPaths {
		if len(ctx.Paths) != 0 {
			txt = fmt.Sprintf("%s", pathsTmpl)
			for k := range ctx.Paths {
				txt = fmt.Sprintf("%s  %s\n", txt, k)
			}
		}
	}
	return txt
}

func addURLs(ctx *Context) string {
	var txt string

	if !fNoURLs {
		if len(ctx.URLs) != 0 {
			txt = fmt.Sprintf("%s", urlsTmpl)
			for k, v := range ctx.URLs {
				if v == ActionBlock {
					txt = fmt.Sprintf("%s  %s\n", txt, k)
				}
			}
		}
	}
	return txt
}

func doSendMail(ctx *Context) (err error) {
	if len(ctx.Paths) != 0 || len(ctx.URLs) != 0 {
		mailText, err := createMail(ctx)
		if err != nil {
			return err
		}

		// Really sendmail now
		if fDoMail {
			return sendMail(ctx, mailText)
		}

		// Otherwise, display it
		fmt.Printf("From: %s\n", ctx.config.From)
		fmt.Printf("Cc: %s\n", ctx.config.Cc)
		fmt.Println(mailText)

		return nil
	}

	// Send dummy mail if verbose
	if fDoMail && fVerbose {
		debug("A mail would have been sent here.")
		txt, _ := createMail(ctx)
		debug("mail content:\n%s", txt)
	}
	log.Print("Nothing to do…")

	return
}

func sendMail(ctx *Context, text string) (err error) {
	var to []string

	verbose("Connecting to %s…", ctx.config.Server)
	from := ctx.config.From

	// Debug mode only send to me
	if fDebug {
		to = []string{from}
	} else {
		to = strings.Split(ctx.config.To, ",")
		if ctx.config.Cc != "" {
		    cc := strings.Split(ctx.config.Cc, ",")
		    to = append(to, cc...)
        }
	}

	debug("from: %s - To: %v", from, to)

	err = smtp.SendMail(ctx.config.Server, nil, from, to, []byte(text))

	verbose("Mail sent to %v…", to)
	return
}
