package analyzers

import (
	"io"
)

// StringAnalyzer extracts and reports strings found in a file
func StringAnalyzer(filename string, r io.ReadSeeker, rep Reporter, data ...interface{}) error {

	strs, err := extractStrings(r, 5)
	if err != nil {
		return err
	}
	report := map[string]interface{}{"strings": strs}
	rep("", "json", report)
	return nil
}
