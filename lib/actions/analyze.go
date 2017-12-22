package actions

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"

	"bitbucket.org/vahidi/molly/lib/types"
	"bitbucket.org/vahidi/molly/lib/util/logging"
)

// Analyzer is the type of functions that will be called in
// an analyze() operation
type Analyzer func(io.ReadSeeker, io.Writer, ...interface{}) error

// RegisterAnalyzer registers a user analyzer
func RegisterAnalyzer(typ string, analyzerfunc Analyzer) {
	analyzers[typ] = analyzerfunc
}

var analyzers = map[string]Analyzer{
	"strings":   stringAnalyzer,
	"histogram": histogramFunction,
}

// analyzeFunction performs some type of analysis on the current binary
func analyzeFunction(e *types.Env, typ string, prefix string, data ...interface{}) (string, error) {
	f, found := analyzers[typ]
	if !found {
		return "", fmt.Errorf("Unknown analyzer: '%s'", typ)
	}

	filename := types.FileName(e)
	path := fmt.Sprintf("%s_%s_%s", typ, prefix, filename)

	w, err := logging.Create(path)
	if err != nil {
		return "", err
	}
	defer w.Close()
	r, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer r.Close()

	return typ, f(r, w, data)
}

func histogramFunction(r io.ReadSeeker, w io.Writer, data ...interface{}) error {
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

func stringAnalyzer(r io.ReadSeeker, w io.Writer, data ...interface{}) error {
	minsize := 5
	bs := bytes.Buffer{}
	br := bufio.NewReader(r)
	for {
		c, err := br.ReadByte()
		if err != nil && err != io.EOF {
			return err
		}
		if err == nil && c >= 0x20 && c < 0x7f { // BSD isprint(c)
			bs.WriteByte(c)
		} else {
			if bs.Len() >= minsize {
				fmt.Fprintf(w, "%s\n", string(bs.Bytes()))
			}
			bs.Reset()
		}
		if err == io.EOF {
			return nil
		}
	}
}

func init() {
	types.FunctionRegister("analyze", analyzeFunction)
}
