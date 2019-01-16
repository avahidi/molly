package analyzers

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"
)

// Analysis represents results of an analysis performed on this file
type Analysis struct {
	Name   string
	Result interface{}

	CreateFile bool
	Error      error
}

// NewAnalysis create an Analysis with the required format
func NewAnalysis(name string, result interface{}, params ...interface{}) *Analysis {
	for _, d := range params {
		name = fmt.Sprintf("%s__%v", name, d)
	}
	return &Analysis{Name: name, Result: result}
}

// Analyzer is a function for analyzing a stream and generating one or more reports
type Analyzer func(filename string, r io.ReadSeeker, res *Analysis, data ...interface{})

// extractStrings extract strings from reader, similar to UNIX strings utility
func extractStrings(r io.Reader, minsize int) ([]string, error) {
	bs := bytes.Buffer{}
	br := bufio.NewReader(r)
	set := make(map[string]bool)

	for {
		c, err := br.ReadByte()
		if err != nil && err != io.EOF {
			return nil, err
		}
		if err == nil && c >= 0x20 && c < 0x7f { // BSD isprint(c)
			bs.WriteByte(c)
		} else {
			if bs.Len() >= minsize {
				set[string(bs.Bytes())] = true
			}
			bs.Reset()
		}

		if err == io.EOF {
			// convert set to array :(
			var ret []string
			for k, _ := range set {
				ret = append(ret, k)
			}
			return ret, nil
		}
	}
}

// containsOnly returns true of text contains only letters in chars
func containsOnly(text, chars string) bool {
	for _, r := range text {
		if !strings.ContainsRune(chars, r) {
			return false
		}
	}
	return true
}
