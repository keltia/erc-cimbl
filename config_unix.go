// +build !windows

package main

import (
	"os"
	"path/filepath"
)

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
