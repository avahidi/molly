package analyzers

import (
	"bufio"
	"bytes"
	"image"
	"image/color"
	"image/png"
	"io"
)

func histogramToImage(hist []int, max int) image.Image {
	if max < 1 {
		max = 1 /* avoid divide by zero */
	}
	palette := []color.Color{color.White, color.Black}
	img := image.NewPaletted(image.Rect(0, 0, 256, 100), palette)
	for x, y := range hist {
		n := 100 - (y*100)/max
		img.SetColorIndex(x, n, 1)

	}
	return img
}

// HistogramAnalyzer creates histograms out of a binary files
func HistogramAnalyzer(r io.ReadSeeker, rep Reporter, data ...interface{}) error {

	// extract histogram file file
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
	report := map[string]interface{}{
		"histrogram": count,
		"sum":        sum,
		"min":        min,
		"max":        max,
		"avg":        avg,
	}
	rep("", "json", report)

	// PNG image
	img := histogramToImage(count, max)
	buff := &bytes.Buffer{}
	if err := png.Encode(buff, img); err != nil {
		return err
	}
	rep("", "png", buff.Bytes())

	return nil
}
