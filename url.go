package main

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/keltia/proxy"
	"github.com/pkg/errors"
)

const (
	ActionAuth    = "AUTH"
	ActionBlock   = "**BLOCK**"
	ActionBlocked = "BLOCKED-EEC"
)

func doCheck(ctx *Context, req *http.Request) (string, error) {

	resp, err := ctx.Client.Do(req)
	if err != nil {
		debug("err: %s", err)
		return "", errors.Wrap(err, "Do")
	}

	switch resp.StatusCode {
	case 403:
		return ActionBlocked, nil
	case 407:
		return ActionAuth, nil
	default:
		return ActionBlock, nil
	}
}

var (
	ErrHttpsSkip  = errors.New("skipping https")
	ErrParseError = errors.New("error parsing URL")
)

func checkForIP(str string) net.IP {
	if str == "" {
		return net.IP{}
	}
	l := len(str)
	if str[0] == '[' {
		if str[l-1] == ']' {
			return net.ParseIP(str[1 : l-1])
		}
	}
	return net.ParseIP(str)
}

func sanitize(str string) (out string, err error) {
	// We do not try to see if there is an error because of some corner cases
	myurl, err := url.Parse(str)
	if err != nil {
		str = "http://" + str
		myurl, err = url.Parse(str)
	}

	// We do not test https
	if myurl.Scheme == "https" {
		return str, ErrHttpsSkip
	}

	// If a special scheme, transform it into http for convenience
	if myurl.Scheme != "http" {
		myurl.Scheme = "http"
	}

	// If host is empty, we might have IP or [IP] or bare hostname
	if myurl.Host == "" {
		//
		if ip := checkForIP(myurl.Path); ip != nil {
			myurl.Host = ip.String()
			myurl.Path = ""
		} else {
			// Path is actually Host/Path
			u, _ := url.Parse(myurl.Path)
			myurl.Host = u.Host
			myurl.Path = u.Path
		}
		return myurl.String(), nil
	}

	// check for [IP]
	l := len(myurl.Host)
	if myurl.Host[0] == '[' && myurl.Host[l-1] == ']' {
		myurl.Host = myurl.Host[1 : l-1]
	}
	return myurl.String(), err
}

func handleURL(ctx *Context, str string) error {

	// https URLs will not be blocked, no MITM
	myurl, err := sanitize(str)
	if err == ErrHttpsSkip {
		skipped = append(skipped, str)
		return nil
	}
	debug("url=%s", myurl)
	/*
	   Setup connection including proxy stuff
	*/
	_, transport := proxy.SetupTransport(myurl)
	if transport == nil {
		return fmt.Errorf("SetupTransport")
	}

	// It is better to re-use than creating a new one each time
	if ctx.Client == nil {
		ctx.Client = &http.Client{Transport: transport, Timeout: 10 * time.Second}
	}

	/*
	   Do the thing, manage redirects, auth requests and stuff
	*/
	req, err := http.NewRequest("HEAD", myurl, nil)
	if err != nil {
		return errors.Wrap(err, "NewRequest")
	}
	req.Header.Set("User-Agent", fmt.Sprintf("%s/%s", MyName, MyVersion))

	result, err := doCheck(ctx, req)
	if err != nil {
		return errors.Wrap(err, "doCheck")
	}
	if result == ActionBlock {
		ctx.URLs[myurl] = result
	}
	verbose("Checking %s: %s", myurl, result)
	return nil
}
