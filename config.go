package main

import (
	"fmt"
	"github.com/naoina/toml"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"bufio"
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

var (
	baseDir = filepath.Join(os.Getenv("HOME"),
		".config",
		MyName,
	)

	configName = "config.toml"

	dbrcFile = filepath.Join(os.Getenv("HOME"), ".dbrc")

	user     string
	password string
)

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

func loadDbrc(filename string) (err error) {
	err = setupProxy(filename)
	if err != nil {
		log.Printf("No dbrc file: %v", err)
	}
	if fVerbose {
		log.Printf("Proxy user %s found.", user)
	}
	return
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
