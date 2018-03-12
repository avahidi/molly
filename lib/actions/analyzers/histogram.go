package analyzers

import (
	"bufio"
	"io"
)

// HistogramAnalyzer creates histograms out of a binary files
func HistogramAnalyzer(r io.ReadSeeker,
	gen func(name string, typ string, data interface{}),
	data ...interface{}) error {
	count := make([]int, 256)
	br := bufio.NewReader(r)
	for {
		c, err := br.ReadByte()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		count[c]++
	}

	report := map[string]interface{}{"histrogram": count}
	gen("", "json", report)

	return nil
}
