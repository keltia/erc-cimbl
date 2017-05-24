package main

import (
    "path/filepath"
    "os"
    "io/ioutil"
    "fmt"
    "log"
    "github.com/naoina/toml"
)

type Config struct {
	From    string
	To      string
	Server  string
	Subject string
}

var (
    baseDir = filepath.Join(os.Getenv("HOME"),
        ".config",
        MyName,
    )

    configName = "config.toml"
)

func loadConfig() (c *Config, err error) {
    file := filepath.Join(baseDir, configName)

    // Check if there is any config file
    if _, err := os.Stat(file); err != nil {
        c = &Config{}
        return
    }

    log.Printf("file=%s, found it", file)
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
