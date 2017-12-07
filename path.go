package main

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	fileEXTS = []string{
		"apk", "app", "bat", "cab",
		"chm", "cmd", "com", "dll",
		"exe", "hlp", "hta", "inf",
		"jar", "jnl", "jnt", "js",
		"jse", "lnk", "mht", "mhtml",
		"msh", "msh1", "msh1xml", "msh2",
		"msh2xml", "msi", "msp", "ocx",
		"pif", "ps1", ".ps1xml", "ps2",
		"s2xml", "psc1", "psc2", "pub",
		"reg", "scf", "scr", "url", "vb",
		"vbe", "vbs", "ws", "wsc",
		"wsf", "wsh", "mst", "msu",
		".ova", ".ovf", ".vhd", ".vhdx",
		".vmcx", ".vmdk", ".vmx", ".xva",
		".ani", ".cpl", ".iso", ".sct",
		".vdi", "ace",
	}

	restr *regexp.Regexp
)

func init() {
	restr = regexp.MustCompile(fmt.Sprintf("\\.(i:%s)$", strings.Join(fileEXTS, "|")))
}

func handlePath(ctx *Context, path string) {
	if !restr.MatchString(path) {
		if ok, _ := ctx.Paths[path]; !ok {
			verbose("Filename %s CHECK", path)
			ctx.Paths[path] = true
		}
	} else {
		verbose("Filename %s: IGNORED ", path)
	}
}

//
// <filename>|<sig>
func entryToPath(entry string) (path string) {
	all := strings.Split(entry, "|")
	path = all[0]
	return
}
