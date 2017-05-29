package main

import (
	"bufio"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var (
	dbrcFile = filepath.Join(os.Getenv("HOME"), ".dbrc")

	user     string
	password string

	proxyURL *url.URL
)

func init() {
	err := setupProxy(dbrcFile)
	if err != nil {
		log.Printf("No dbrc file: %v", err)
	}
	if fVerbose {
		log.Printf("Proxy user %s found.", user)
	}
}

func setupProxy(file string) (err error) {
	fh, err := os.Open(file)
	if err != nil {
		return fmt.Errorf("Error: can not find %s: %v", dbrcFile, err)
	}
	defer fh.Close()

	/*
	   Format:
	   <db>     <user>    <pass>   <type>
	*/
	scanner := bufio.NewScanner(fh)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			break
		}

		// Replace all tabs by a single space
		l := strings.Replace(line, "\t", " ", -1)
		flds := strings.Split(l, " ")

		// Check what we need
		if flds[0] == "cimbl" {
			user = flds[1]
			password = flds[2]
			break
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("reading dbrc %s", dbrcFile)
	}

	if user == "" {
		return fmt.Errorf("no user/password for cimbl in %s", dbrcFile)
	}

	return
}

func setupCheck(str string) (*http.Request, *http.Transport) {

	// Fix really invalid URLs
	if !strings.HasPrefix(str, "http://") {
		str = "http://" + str
	}

	/*
	   Proxy code taken from https://github.com/LeoCBS/poc-proxy-https/blob/master/main.go
	*/
	myurl, err := url.Parse(str)
	if err != nil {
		log.Printf("error parsing %s: %v", str, err)
		return nil, nil
	}

	req, err := http.NewRequest("HEAD", str, nil)
	if err != nil {
		log.Printf("error: req is nil: %v", err)
		return nil, nil
	}
	req.Header.Set("Host", myurl.Host)
	req.Header.Add("User-Agent", fmt.Sprintf("%s/%s", MyName, MyVersion))

	proxyURL, err = http.ProxyFromEnvironment(req)
	if err != nil {
		log.Printf("Error: no user/password for cimbl.")
		os.Exit(2)
	} else if proxyURL == nil {
		log.Println("No proxy configured")
	} else {
		auth := fmt.Sprintf("%s:%s", user, password)
		basic := "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
		req.Header.Add("Proxy-Authorization", basic)
	}

	transport := &http.Transport{
		Proxy:           http.ProxyURL(proxyURL),
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	transport.ProxyConnectHeader = req.Header

	return req, transport
}

func doCheck(ctx *Context, req *http.Request) string {
	//req.RequestURI = ""

	resp, err := ctx.Client.Do(req)
	if err != nil {
		if fVerbose {
			log.Printf("err: %s", err)
		}
		return ""
	}

	switch resp.StatusCode {
	case 302:
		return fmt.Sprintf("REDIRECT: %s", resp.Header.Get("Location"))
	case 403:
		return "BLOCKED-EEC"
	case 407:
		return "AUTH"
	default:
		return "**BLOCK**"
	}

}

func handleURL(ctx *Context, str string) {
	/*
	   Setup connection including proxy stuff
	*/
	req, transport := setupCheck(str)
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
	ctx.URLs[str] = result
	if fVerbose {
		log.Printf("Checking %s: %s", str, result)
	}
}
