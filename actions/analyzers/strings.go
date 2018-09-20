package analyzers

import (
	"io"
)

// StringAnalyzer extracts and reports strings found in a file
func StringAnalyzer(filename string, r io.ReadSeeker, res *Analysis, data ...interface{}) {

	strs, err := extractStrings(r, 5)
	if err != nil {
		res.Error = err
		return
	}
	res.Result = map[string]interface{}{"strings": strs}
}
