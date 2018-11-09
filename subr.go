package main

import (
	"regexp"
)

func checkFilename(file string) (ok bool) {
	re := regexp.MustCompile(`(?i:CIMBL-\d+-CERTS\.(csv|zip)(\.asc|))`)

	return re.MatchString(file)
}

func checkOpenPGP(file string) bool {
	re := regexp.MustCompile(`(?i:OpenPGP.*\.asc)`)

	return re.MatchString(file)
}
