package main

import (
	"bytes"
	"fmt"
	"log"
	"strings"
	"text/template"
)

var (
	MailTmpl = `
Dear Service Desk,

After reading the following files received from CERT-EU:
  {{.Files}}

{{.Paths}}
{{.URLs}}
Best regards,
Your friendly script — {{.MyName}}/{{.MyVersion}}
    `

	PathsTmpl = "Please add the following to the list of blocked filenames:\n"
	URLsTmpl  = "Please add the following to the list of blocked URLs on BlueCoat:\n"
)

type MailVars struct {
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

	vars := MailVars{
		From:      ctx.config.From,
		To:        ctx.config.To,
		Subject:   ctx.config.Subject,
		MyName:    MyName,
		MyVersion: MyVersion,
		Files:     strings.Join(ctx.files, ", "),
	}

	vars.Paths = addPaths(ctx)
	vars.URLs = addURLs(ctx)

	t := template.Must(template.New("mail").Parse(MailTmpl))
	err = t.Execute(&txt, vars)
	return txt.String(), err
}

func addPaths(ctx *Context) string {
	var txt string

	if !fNoPaths {
		if len(ctx.Paths) != 0 {
			txt = fmt.Sprintf("%s", PathsTmpl)
			for k, _ := range ctx.Paths {
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
			txt = fmt.Sprintf("%s", URLsTmpl)
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

	mailText, err := createMail(ctx)

	if len(ctx.Paths) != 0 || len(ctx.URLs) != 0 {
		if fDoMail {
			err := sendMail(ctx, mailText)
			if err != nil {
				log.Fatalf("sending mail: %v", err)
			}
		} else {
			fmt.Println(mailText)
		}
	}
	return
}

func sendMail(ctx *Context, text string) (err error) {
	return
}
