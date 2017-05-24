package main

import (
	"bytes"
	"fmt"
	"log"
	"text/template"
)

var (
	MailTmpl = `
Dear Service Desk,

{{.Paths}}
{{.URLs}}

Best regards,
Your friendly script — {{.MyName}}/{{.MyVersion}}
    `

	PathsTmpl = "Please add the following to the list of blocked filenames:"
	URLsTmpl  = "Please add the following to the list of blocked URLs on BlueCoat:"
)

type MailVars struct {
	From      string
	To        string
	Subject   string
	MyName    string
	MyVersion string
	URLs      string
	Paths     string
}

func createMail(config *Config) (str string, err error) {
	var txt bytes.Buffer

	vars := MailVars{
		From:      config.From,
		To:        config.To,
		Subject:   config.Subject,
		MyName:    MyName,
		MyVersion: MyVersion,
	}

	fmt.Println("Results:")

    vars.Paths = addPaths()
    vars.URLs  = addURLs()

	t := template.Must(template.New("mail").Parse(MailTmpl))
	err = t.Execute(&txt, vars)
	return txt.String(), err
}

func addPaths() string {
	var txt string

	if !fNoPaths {
		if cntPaths != 0 {
			txt = fmt.Sprintf("%s", PathsTmpl)
			for k, _ := range Paths {
				txt = fmt.Sprintf("%s  %s\n", txt, k)
			}
		}
	}
	return txt
}

func addURLs() string {
	var txt string

	if !fNoURLs {
		if cntURLs != 0 {
			txt = fmt.Sprintf("%s", URLsTmpl)
            for k, v := range URLs {
                if v == "**BLOCK**" {
                    txt = fmt.Sprintf("%s  %s\n", txt, k)
                }
			}
		}
	}
	return txt
}

func doSendMail(config *Config) (err error) {

	mailText, err := createMail(config)

	if fDoMail {
		err := sendMail(config, mailText)
		if err != nil {
			log.Fatalf("sending mail: %v", err)
		}
	} else {
		fmt.Println(mailText)
	}
	return
}

func sendMail(config *Config, text string) (err error) {
	return
}
