package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"net/mail"
	"time"

	"github.com/keltia/archive"
	"github.com/pkg/errors"
)

func decryptMultipart(ctx *Context, file string) ([]byte, error) {
	debug("decrypt/multipart")
	// Process the file (gpg encrypted mail file)
	a, err := archive.New(file)
	if err != nil {
		return []byte{}, errors.Wrap(err, "decrypt/archive/new")
	}

	// Decryption
	unc, err := a.Extract("")
	if err != nil {
		return []byte{}, errors.Wrap(err, "decrypt/extract")
	}

	return unc, nil
}

/*
Current format is:

multipart/mixed
- multipart/mxied
  - text/csv
  - text/xml
*/

const body = `From nobody@cert.europa.eu %s
From: nobody@cert.europa.eu
Subject: Fake mail
Content-Type: multipart/mixed; boundary="fake-mail-just-for-cert-eu"

--fake-mail-just-for-cert-eu

%s
--fake-mail-just-for-cert-eu--

`

func handleMultipart(ctx *Context, file string) (*Results, error) {
	debug("handle/multipart")
	content, err := decryptMultipart(ctx, file)
	if err != nil {
		return &Results{}, errors.Wrap(err, "multipart/decrypt")
	}

	buf := bytes.NewReader(content)

	debug("decoded %d bytes\n", len(content))

	msg, err := mail.ReadMessage(buf)
	if err != nil {
		return &Results{}, errors.Wrap(err, "multipart/message")
	}

	var bnd string

	ct, p, err := mime.ParseMediaType(msg.Header.Get("content-type"))
	if ct == "multipart/mixed" {
		bnd = p["boundary"]
		debug("bnd=%s", bnd)
		r := multipart.NewReader(msg.Body, bnd)

		rp, err := r.NextPart()
		debug("rp=%#v fn=%s", rp, rp.FileName())

		var rpbody []byte

		_, err = rp.Read(rpbody)

		csv, err := handleMixed(ctx, rpbody)
		if err != nil {
			return &Results{}, errors.Wrap(err, "handlemixed")
		}
		debug("csv=%s\n", string(csv))

		rcsv := bytes.NewReader(csv)
		return handleCSV(ctx, rcsv)
	}
	return &Results{}, fmt.Errorf("not mutipart/mixed")
}

// handleMixed decodes the second level multipart/mixed
func handleMixed(ctx *Context, pp []byte) ([]byte, error) {
	debug("handle/mixed")

	today := time.Now().Format(time.RFC822Z)

	// Fake headers to make it look like a mail
	mybody := fmt.Sprintf(body, today, string(pp))

	debug("mybody=%s\n", mybody)
	buf := bytes.NewReader([]byte(mybody))

	debug("mixed/readmessage")
	msg, err := mail.ReadMessage(buf)
	if err != nil {
		return []byte{}, errors.Wrap(err, "mixed/message")
	}

	debug("msg:%#v\n", msg)

	var bnd string

	ct, p, err := mime.ParseMediaType(msg.Header.Get("content-type"))
	debug("ct=%s\n", ct)
	if ct == "multipart/mixed" {
		bnd = p["boundary"]
		debug("bnd=%s", bnd)
	} else {
		return nil, fmt.Errorf("bad content-type %s", ct)
	}

	rp := multipart.NewReader(msg.Body, bnd)

	debug("got new multipart rp=%#v", rp)
	for {
		p, err := rp.NextPart()
		if err == io.EOF {
			return nil, fmt.Errorf("no relevant part")
		}

		if err != nil {
			return nil, errors.Wrap(err, "mixed/readpart")
		}
		if p.Header.Get("content-type") == "text/csv" &&
			checkFilename(p.FileName()) {
			part, err := ioutil.ReadAll(p)
			if err != nil {
				return nil, errors.Wrap(err, "mixed/bodypart")
			}
			return part, nil
		}
	}

}
