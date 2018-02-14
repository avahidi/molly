package actions

import (
	"fmt"
	"io"
	"os"

	"bitbucket.org/vahidi/molly/lib/actions/analyzers"
	"bitbucket.org/vahidi/molly/lib/types"
	"bitbucket.org/vahidi/molly/lib/util"
)

// Analyzer is the type of functions that will be called in
// an analyze() operation
type Analyzer func(io.ReadSeeker, io.Writer, ...interface{}) error

// RegisterAnalyzer registers a user analyzer
func RegisterAnalyzer(typ string, analyzerfunc Analyzer) {
	analyzersList[typ] = analyzerfunc
}

var analyzersList = map[string]Analyzer{
	"strings":   analyzers.StringAnalyzer,
	"version":   analyzers.VersionAnalyzer,
	"histogram": analyzers.HistogramAnalyzer,
	"elf":       analyzers.ElfAnalyzer,
}

// analyzeFunction performs some type of analysis on the current binary
func analyzeFunction(e *types.Env, typ string, prefix string, data ...interface{}) (string, error) {
	f, found := analyzersList[typ]
	if !found {
		fmt.Printf("Available analyzers are:\n")
		for k := range analyzersList {
			fmt.Printf("\t%s\n", k)
		}
		util.RegisterFatal("Unknown analyzer: '%s'", typ)
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

func init() {
	types.FunctionRegister("analyze", analyzeFunction)
}
