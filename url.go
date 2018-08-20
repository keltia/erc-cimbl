package main

import (
	"fmt"
	"github.com/keltia/proxy"
	"github.com/pkg/errors"
	"net"
	"net/http"
	"net/url"
	"time"
)

const (
	ActionAuth    = "AUTH"
	ActionBlock   = "**BLOCK**"
	ActionBlocked = "BLOCKED-EEC"
)

var (
	proxyURL *url.URL
)

func doCheck(ctx *Context, req *http.Request) string {
	//req.RequestURI = ""

	req.Header.Set("User-Agent", fmt.Sprintf("%s/%s", MyName, MyVersion))

	resp, err := ctx.Client.Do(req)
	if err != nil {
		verbose("err: %s", err)
		return ""
	}

	switch resp.StatusCode {
	case 403:
		return ActionBlocked
	case 407:
		return ActionAuth
	default:
		return ActionBlock
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
	myurl, _ := url.Parse(str)
	if myurl == nil {
		return str, ErrParseError
	}

	// We do not test https
	if myurl.Scheme == "https" {
		return str, ErrHttpsSkip
	}

	// If a special scheme, transform it into http for convenience
	if myurl.Scheme != "http" {
		myurl.Scheme = "http"
	}

	// If host is empty, we might have IP or [IP]
	if myurl.Host == "" {
		//
		if ip := checkForIP(myurl.Path); ip != nil {
			myurl.Host = ip.String()
			myurl.Path = ""
			return myurl.String(), err
		}
		return str, ErrParseError
	}

	// check for [IP]
	l := len(myurl.Host)
	if myurl.Host[0] == '[' && myurl.Host[l-1] == ']' {
		myurl.Host = myurl.Host[1 : l-1]
	}
	return myurl.String(), err
}

func handleURL(ctx *Context, str string) {

	// https URLs will not be blocked, no MITM
	myurl, err := sanitize(str)
	if err == ErrHttpsSkip {
		skipped = append(skipped, str)
		return
	}

	/*
	   Setup connection including proxy stuff
	*/
	req, transport := proxy.SetupTransport(myurl)
	if req == nil || transport == nil {
		return
	}

	// It is better to re-use than creating a new one each time
	if ctx.Client == nil {
		ctx.Client = &http.Client{Transport: transport, Timeout: 10 * time.Second}
	}

	/*
	   Do the thing, manage redirects, auth requests and stuff
	*/
	result := doCheck(ctx, req)
	if result != "" {
		if result == ActionBlock {
			ctx.URLs[myurl] = result
		}
		verbose("Checking %s: %s", myurl, result)
	}
}
