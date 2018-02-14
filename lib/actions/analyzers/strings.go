package analyzers

import (
	"fmt"
	"io"
)

// StringAnalyzer extracts and reports strings found in a file
func StringAnalyzer(r io.ReadSeeker, w io.Writer, data ...interface{}) error {
	strs, err := extractStrings(r, 5)
	if err != nil {
		return err
	}
	for _, str := range strs {
		fmt.Fprintf(w, "%s\n", str)
	}
	return nil
}
