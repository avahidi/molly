package analyzers

import (
	"io"
)

// StringAnalyzer extracts and reports strings found in a file
func StringAnalyzer(r io.ReadSeeker, data ...interface{}) (map[string]interface{}, error) {
	strs, err := extractStrings(r, 5)
	if err != nil {
		return nil, err
	}
	report := map[string]interface{}{"strings": strs}
	return report, nil
}
