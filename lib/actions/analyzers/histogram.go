package analyzers

import (
	"bufio"
	"fmt"
	"io"
)

// HistogramAnalyzer creates histograms out of a binary files
func HistogramAnalyzer(r io.ReadSeeker, w io.Writer, data ...interface{}) error {
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

	fmt.Fprintf(w, "histogram = {")
	for i, v := range count {
		if (i & 15) == 0 {
			fmt.Fprintf(w, "\n\t")
		}
		fmt.Fprintf(w, "%6d, ", v)
	}
	fmt.Fprintf(w, "\n};\n")
	return nil
}
