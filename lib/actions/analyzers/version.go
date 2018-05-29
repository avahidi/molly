package analyzers

import (
	"io"
	"regexp"
	"strings"
)

var gitrefRegex = regexp.MustCompile("[^0-9,a-f]*[0-9,a-f]{40}[^0-9,a-f]*")
var versionRegex = regexp.MustCompile("(v[ ]*\\d\\.\\d+)|(version[ ]*\\d\\.)|((\\d+)\\.(\\d+\\.))")
var copyrightRegex = regexp.MustCompile("(copyright|\\(c\\))(.?)*[12][0-9]{3}")
var ipnumberRegex = regexp.MustCompile("([\\d]{1,3}\\.){3}[\\d]{1,3}[^.]?")

func stringIsGitref(str string) bool {
	return gitrefRegex.MatchString(str)
}

func stringIsVersion(str string) bool {
	return versionRegex.MatchString(str) && !ipnumberRegex.MatchString(str)
}

func stringIsCopyright(str string) bool {
	str = strings.ToLower(str)
	return copyrightRegex.MatchString(str)
}

// VersionAnalyzer is a first attempt to extract version information from binaries
func VersionAnalyzer(r io.ReadSeeker, rep Reporter, data ...interface{}) error {
	strs, err := extractStrings(r, 5)
	if err != nil {
		return err
	}

	hashes := make([]string, 0)
	versions := make([]string, 0)
	copyrights := make([]string, 0)

	for _, str := range strs {
		t := strings.Trim(str, " \t\n\r")
		tl := strings.ToLower(t)

		if stringIsGitref(tl) {
			hashes = append(hashes, t)
		}
		if stringIsVersion(tl) {
			versions = append(versions, t)
		}
		if stringIsCopyright(tl) {
			copyrights = append(copyrights, t)
		}
	}

	// build report
	report := map[string]interface{}{
		"possible-gitref":    hashes,
		"possible-version":   versions,
		"possible-copyright": copyrights,
	}
	rep("", "json", report)
	return nil
}
