package analyzers

import (
	"bufio"
	"io"
)

// HistogramAnalyzer creates histograms out of a binary files
func HistogramAnalyzer(filename string, r io.ReadSeeker, data ...interface{}) (interface{}, error) {

	// extract histogram file file
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

	// extract min, max, sum and average
	min, max, sum := count[0], count[0], 0
	for _, v := range count {
		if min > v {
			min = v
		}
		if max < v {
			max = v
		}
		sum += v
	}
	avg := sum / len(count)

	// jost report
	ret := map[string]interface{}{
		"histrogram": count,
		"sum":        sum,
		"min":        min,
		"max":        max,
		"avg":        avg,
	}
	return ret, nil
}
