// +build !unix,windows

package main

import (
	"os"
	"path/filepath"
)

var (
	baseDir = filepath.Join(os.Getenv("%LOCALAPPDATA%"),
		"DG-CSS",
		"CIMBL",
	)

	configName = "config.toml"

	dbrcFile = filepath.Join(baseDir, "dbrc")

	user     string
	password string
)
