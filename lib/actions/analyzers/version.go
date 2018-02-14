package analyzers

import (
	"encoding/json"
	"io"
	"strings"
)

// VersionAnalyzer is a first attempt to extract version information from binaries
func VersionAnalyzer(r io.ReadSeeker, w io.Writer, data ...interface{}) error {
	strs, err := extractStrings(r, 5)
	if err != nil {
		return err
	}

	hashes := make([]string, 0)
	versions := make([]string, 0)
	for _, str := range strs {
		t := strings.Trim(str, " \t\n\r")
		tl := strings.ToLower(t)

		if len(tl) == 40 && containsOnly(tl, "0123456789abcdef") {
			hashes = append(hashes, t)
		}

		if strings.HasPrefix(tl, "version") {
			versions = append(versions, t)
		}
	}

	// build report
	report := map[string]interface{}{
		"possible-gitref":  hashes,
		"possible-version": versions,
	}

	// write report
	bs, err := json.MarshalIndent(report, "", "\t")
	if err != nil {
		return err
	}
	w.Write(bs)
	return nil
}
