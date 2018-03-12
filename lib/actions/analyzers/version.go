package analyzers

import (
	"io"
	"regexp"
	"strings"
)

var versionRegex = regexp.MustCompile("(version[ ]*\\d)|((\\d+)\\.(\\d+\\.))")

// VersionAnalyzer is a first attempt to extract version information from binaries
func VersionAnalyzer(r io.ReadSeeker,
	gen func(name string, typ string, data interface{}),
	data ...interface{}) error {
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
		if versionRegex.MatchString(tl) {
			versions = append(versions, t)
		}
	}

	// build report
	report := map[string]interface{}{
		"possible-gitref":  hashes,
		"possible-version": versions,
	}
	gen("", "json", report)
	return nil
}
