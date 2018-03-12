package analyzers

import (
	"io"
)

// StringAnalyzer extracts and reports strings found in a file
func StringAnalyzer(r io.ReadSeeker,
	gen func(name string, typ string, data interface{}),
	data ...interface{}) error {

	strs, err := extractStrings(r, 5)
	if err != nil {
		return err
	}
	report := map[string]interface{}{"strings": strs}
	gen("", "json", report)
	return nil
}
