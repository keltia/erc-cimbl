package main

import (
	"regexp"
)

func checkFilename(ctx *Context, file string) (ok bool) {
	re := regexp.MustCompile(ctx.config.REFile)

	return re.MatchString(file)
}

func checkOpenPGP(file string) bool {
	re := regexp.MustCompile(`(?i:OpenPGP.*\.asc)`)

	return re.MatchString(file)
}

func checkMultipart(file string) bool {
	re := regexp.MustCompile(`(?i:OpenPGP.*file$)`)

	return re.MatchString(file)
}
