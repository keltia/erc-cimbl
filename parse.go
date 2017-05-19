package main

import (
	"os"
	"encoding/csv"
	"strings"
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
vuln_mgt,indicator_uuid,
indicator_detect_time,
indicator_threat_type,
indicator_threat_level,
indicator_targeted_domain,
indicator_start_time,
indicator_end_time,
indicator_title

We filter on "type", looking for "url" & "filename".

 */
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

	for _, line := range allLines {
		// type at index 2
		// value at index 5
		vtype := line[2]
		etype := strings.Split(vtype, "|")

		switch etype[0] {
		case "filename":
			handlePath(entryToPath(line[5]))
		case "url":
			handleURL(line[5])
		}

	}
	return
}
