package main

import (
    "regexp"
    "strings"
    "fmt"
)

var (
    EXTS = []string{
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
        "wsf", "wsh",
    }

    REstr *regexp.Regexp

)

func init() {
    REstr = regexp.MustCompile(fmt.Sprintf("\\.(i:%s)$", strings.Join(EXTS, "|")))
}

func handlePath(path string) {
    if !REstr.MatchString(path) {
        Paths = append(Paths, path)
        cntPaths++
    }
}

//
// <filename>|<sig>
func entryToPath(entry string) (path string) {
    all := strings.Split(entry, "|")
    path = all[0]
    return
}

