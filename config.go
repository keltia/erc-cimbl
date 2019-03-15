package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Config is the main configuration object
type Config struct {
	From    string
	To      string
	Server  string
	Subject string
	Cc      string
	KeyID   string

	// RE to check filenames
	REFile string
}

func loadConfig() (*Config, error) {
	file := filepath.Join(baseDir, configName)

	// Check if there is any config file
	if _, err := os.Stat(file); err != nil {
		return &Config{}, err
	}

	verbose("file=%s, found it", file)

	// Read it
	buf, err := ioutil.ReadFile(file)
	if err != nil {
		return &Config{}, fmt.Errorf("Can not read %s", file)
	}

	var cnf Config

	if err := toml.Unmarshal(buf, &cnf); err != nil {
		return &Config{}, fmt.Errorf("Error parsing toml %s: %v", file, err)
	}

	// Ensure we got sensible defaults
	if cnf.REFile == "" {
		cnf.REFile = `(?i:CIMBL-\d+-(CERTS|EU)\.(csv|zip)(\.asc|))`
	}
	return &cnf, nil
}
