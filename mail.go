package main

import (
	"bytes"
	"fmt"
	"log"
	"strings"
	"text/template"
	"net/smtp"
)

var (
	mailTmpl = `Subject: {{.Subject}}
To: {{.To}}
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
		Subject:   ctx.config.Subject,
		MyName:    MyName,
		MyVersion: MyVersion,
		Files:     strings.Join(ctx.files, ", "),
	}

	vars.Paths = addPaths(ctx)
	vars.URLs = addURLs(ctx)

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
				if v == "**BLOCK**" {
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

		if fDoMail {
			err := sendMail(ctx, mailText)
			if err != nil {
				log.Fatalf("sending mail: %v", err)
			}
		} else {
			fmt.Printf("From: %s\n", ctx.config.From)
			fmt.Printf("To: %s\n", ctx.config.To)
			fmt.Printf("Cc: %s\n", ctx.config.Cc)
			fmt.Printf("Subject: %s\n\n", ctx.config.Subject)
			fmt.Println(mailText)
		}
	} else {
		/* Send dummy mail if verbose */
		if fDoMail && fVerbose {
			txt, _ := createMail(ctx)
			err = sendMail(ctx, txt)
		}
		log.Print("Nothing to do…")
	}
	return
}

func sendMail(ctx *Context, text string) (err error) {
	if fVerbose {
		log.Printf("Connecting to %s…", ctx.config.Server)
	}
	from := ctx.config.From
	to := strings.Split(ctx.config.To, ",")

	err = smtp.SendMail(ctx.config.Server, nil, from, to, []byte(text))
	if err != nil {
		log.Printf("error sending mail: %v", err)
	}

	if fVerbose {
		log.Printf("Mail sent to %v…", ctx.config.To)
	}
	return
}
