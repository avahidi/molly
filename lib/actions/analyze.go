package actions

import (
	"bitbucket.org/vahidi/molly/lib/types"
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

// extractStrings extract strings from reader, similar to UNIX strings utility
func extractStrings(r io.Reader, minsize int) ([]string, error) {
	bs := bytes.Buffer{}
	br := bufio.NewReader(r)
	var ret []string

	for {
		c, err := br.ReadByte()
		if err != nil && err != io.EOF {
			return ret, err
		}
		if err == nil && c >= 0x20 && c < 0x7f { // BSD isprint(c)
			bs.WriteByte(c)
		} else {
			if bs.Len() >= minsize {
				ret = append(ret, string(bs.Bytes()))
			}
			bs.Reset()
		}
		if err == io.EOF {
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

// Analyzer is the type of functions that will be called in
// an analyze() operation
type Analyzer func(io.ReadSeeker, io.Writer, ...interface{}) error

// RegisterAnalyzer registers a user analyzer
func RegisterAnalyzer(typ string, analyzerfunc Analyzer) {
	analyzers[typ] = analyzerfunc
}

var analyzers = map[string]Analyzer{
	"strings":   stringAnalyzer,
	"version":   versionAnalyzer,
	"histogram": histogramFunction,
}

// analyzeFunction performs some type of analysis on the current binary
func analyzeFunction(e *types.Env, typ string, prefix string, data ...interface{}) (string, error) {
	f, found := analyzers[typ]
	if !found {
		return "", fmt.Errorf("Unknown analyzer: '%s'", typ)
	}

	logfile := fmt.Sprintf("%s_%s", typ, prefix)
	w, err := e.CreateLog(logfile)
	if err != nil {
		return "", err
	}
	defer w.Close()

	filename := e.GetFile()
	r, err := os.Open(filename)
	if err != nil {
		fmt.Fprintf(w, "MOLLY ANALYZER FAILED: %v\n", err) // explain why file is empty
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
	strs, err := extractStrings(r, 5)
	if err != nil {
		return err
	}
	for _, str := range strs {
		fmt.Fprintf(w, "%s\n", str)
	}
	return nil
}

// versionAnalyzer is a first attempt to extract version information from binaries
func versionAnalyzer(r io.ReadSeeker, w io.Writer, data ...interface{}) error {
	strs, err := extractStrings(r, 5)
	if err != nil {
		return err
	}

	var hashes, versions []string
	for _, str := range strs {
		t := strings.Trim(str, " \t\n\r")
		tl := strings.ToLower(t)

		if len(tl) == 40 && containsOnly(tl, "0123456789abcdef") {
			hashes = append(hashes, t)
		}

		if strings.HasPrefix(tl, "version") {
			versions = append(versions, t)
		}
	}

	// build report
	report := make(map[string]interface{})
	if hashes != nil {
		report["possible-gitref"] = hashes
	}

	if versions != nil {
		report["possible-version"] = versions
	}

	// write report
	bs, err := json.Marshal(report)
	if err != nil {
		return err
	}
	w.Write(bs)
	return nil
}

func init() {
	types.FunctionRegister("analyze", analyzeFunction)
}
