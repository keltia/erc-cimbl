package main

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"io/ioutil"
	"os"
	"path/filepath"
)

const (
	proxyTag = "proxy"
)

// Config is the main configuration object
type Config struct {
	From    string
	To      string
	Server  string
	Subject string
	Cc      string
	KeyID   string
}

func loadConfig() (c *Config, err error) {
	file := filepath.Join(baseDir, configName)

	// Check if there is any config file
	if _, err = os.Stat(file); err != nil {
		c = &Config{}
		return
	}

	verbose("file=%s, found it", file)

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

