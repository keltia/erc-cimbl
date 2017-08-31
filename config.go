package main

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"github.com/BurntSushi/toml"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// Config is the main configuration object
type Config struct {
	From    string
	To      string
	Server  string
	Subject string
	Cc      string
}

func loadConfig() (c *Config, err error) {
	file := filepath.Join(baseDir, configName)

	// Check if there is any config file
	if _, err = os.Stat(file); err != nil {
		c = &Config{}
		return
	}

	if fVerbose {
		log.Printf("file=%s, found it", file)
	}

	// Read it
	buf, err := ioutil.ReadFile(file)
	if err != nil {
		return c, fmt.Errorf("Can not read %s", file)
	}

	cnf := Config{}
	err = toml.Unmarshal(buf, &cnf)
	if err != nil {
		return c, fmt.Errorf("Error parsing toml %s: %v", file, err)
	}
	c = &cnf
	return
}

func setupProxyAuth(ctx *Context, filename string) (err error) {
	err = loadDbrc(filename)
	if err != nil {
		log.Printf("No dbrc file: %v", err)
	}
	if fVerbose {
		log.Printf("Proxy user %s found.", user)
	}

	// Do we have a proxy user/password?
	if user != "" && password != "" {
		auth := fmt.Sprintf("%s:%s", user, password)
		ctx.proxyauth = "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
	}

	return
}

func loadDbrc(file string) (err error) {
	fh, err := os.Open(file)
	if err != nil {
		return fmt.Errorf("Error: can not find %s: %v", file, err)
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
