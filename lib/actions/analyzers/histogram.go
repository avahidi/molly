package analyzers

import (
	"bufio"
	"io"
)

// HistogramAnalyzer creates histograms out of a binary files
func HistogramAnalyzer(r io.ReadSeeker, data ...interface{}) (map[string]interface{}, error) {
	count := make([]int, 256)
	br := bufio.NewReader(r)
	for {
		c, err := br.ReadByte()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		count[c]++
	}

	report := map[string]interface{}{"histrogram": count}
	return report, nil
}
