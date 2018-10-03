package main

import (
	"bytes"
	"fmt"
	"log"
	"net/smtp"
	"strings"
	"text/template"

	"github.com/pkg/errors"
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

type MailSender interface {
	SendMail(server, from string, to []string, body []byte) error
}

type SMTPMailSender struct{}

func (SMTPMailSender) SendMail(server, from string, to []string, text []byte) error {
	return smtp.SendMail(server, nil, from, to, text)
}

type NullMailer struct{}

func (NullMailer) SendMail(server, from string, to []string, text []byte) error {
	log.Printf(`There should be a mail to %s:
From %s
To %v
Body
%s `, server, from, to, string(text))
	return nil
}

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

	if ctx == nil {
		return "", fmt.Errorf("null context")
	}
	if ctx.config == nil {
		return "", fmt.Errorf("null config")
	}
	vars := mailVars{
		From:      ctx.config.From,
		To:        ctx.config.To,
		Cc:        ctx.config.Cc,
		Subject:   ctx.config.Subject,
		MyName:    MyName,
		MyVersion: MyVersion,
		Files:     strings.Join(ctx.files, ", "),
		Paths:     addPaths(ctx),
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
			for k, _ := range ctx.URLs {
				txt = fmt.Sprintf("%s  %s\n", txt, k)
			}
		}
	}
	return txt
}

func doSendMail(ctx *Context) (err error) {
	if len(ctx.Paths) != 0 || len(ctx.URLs) != 0 {
		mailText, err := createMail(ctx)
		if err != nil {
			return errors.Wrap(err, "createMail")
		}

		// Really sendmail now
		if fDoMail {
			verbose("Sending the mail")
			return sendMail(ctx, mailText)
		}

		// Otherwise, display it
		fmt.Printf("From: %s\n", ctx.config.From)
		fmt.Printf("Cc: %s\n", ctx.config.Cc)
		fmt.Println(mailText)
	} else {
		log.Print("Nothing to do…")
	}

	return nil
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

	if fDebug {
		verbose("null mailer")
		ctx.mail = NullMailer{}
	}

	verbose("Mail sent to %v…", to)

	return ctx.mail.SendMail(ctx.config.Server, from, to, []byte(text))
}
