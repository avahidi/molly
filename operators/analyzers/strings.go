package analyzers

import (
	"io"
)

// StringAnalyzer extracts and reports strings found in a file
func StringAnalyzer(filename string, r io.ReadSeeker, data ...interface{}) (interface{}, error) {

	strs, err := extractStrings(r, 5)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"strings": strs}, err
}
