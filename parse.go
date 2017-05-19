package main

import (
	"os"
	"encoding/csv"
	"regexp"
	"strings"
	"fmt"
)

var (
	EXTS = []string{
		".apk", ".app", ".bat", ".cab",
		".chm", ".cmd", ".com", ".dll",
		".exe", ".hlp", ".hta", ".inf",
		".jar", ".jnl", ".jnt", ".js",
		".jse", ".lnk", ".mht", ".mhtml",
		".msh", ".msh1", ".msh1xml", ".msh2",
		".msh2xml", ".msi", ".msp", ".ocx",
		".pif", ".ps1", ".ps1xml", ".ps2",
		".ps2xml", ".psc1", ".psc2", ".pub",
		".reg", ".scf", ".scr", ".url", ".vb",
		".vbe", ".vbs", ".ws", ".wsc",
		".wsf", ".wsh",
	}

	REstr string
)

func init() {
	REstr = fmt.Sprintf("(%s)$", strings.Join(EXTS, "|"))
}

func handlePath(path string) {
	if regexp.MatchString(REstr, path) {

	}
}

func handleURL(url string) {

}

func handleCSV(file string) {
	var fh *os.File

	_, err := os.Stat(file)
	if err != nil {
		return
	}

	if fh, err = os.Open(file); err != nil {
		return
	}
	defer fh.Close()

	all := csv.NewReader(fh)
	allLines, err := all.ReadAll()

	entries := make(map[string]string)
	for _, line := range allLines {

	}
	return
}
