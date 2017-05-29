package main

import (
	"encoding/csv"
	"os"
	"strings"
)

var (
	URLs    = map[string]string{}
	cntURLs int

	Paths    = map[string]bool{}
	cntPaths int
)

/*
Fields in the CSV file:

observable_uuid,
kill_chain,
type,
time_start,
time_end,
value,
to_ids,
blacklist,
malware_research,
vuln_mgt,
indicator_uuid,
indicator_detect_time,
indicator_threat_type,
indicator_threat_level,
indicator_targeted_domain,
indicator_start_time,
indicator_end_time,
indicator_title

We filter on "type", looking for "url" & "filename".

*/

func openFile(file string) (fh *os.File, err error) {
	_, err = os.Stat(file)
	if err != nil {
		return
	}

	fh, err = os.Open(file)
	return
}

func handleCSV(ctx *Context, file string) (err error) {
	fh, err := openFile(file)
	if err != nil {
		return
	}
	defer fh.Close()

	all := csv.NewReader(fh)
	allLines, err := all.ReadAll()

	for _, line := range allLines {
		// type at index 2
		// value at index 5
		vtype := line[2]
		etype := strings.Split(vtype, "|")

		switch etype[0] {
		case "filename":
			if!fNoPaths {
				handlePath(ctx, entryToPath(line[5]))
			}
		case "url":
			if !fNoURLs {
				handleURL(ctx, line[5])
			}
		}
	}
	return nil
}
