package actions

import (
	"bitbucket.org/vahidi/molly/lib/actions/analyzers"
	"bitbucket.org/vahidi/molly/lib/types"
	"bitbucket.org/vahidi/molly/lib/util"
	"encoding/json"
	"fmt"
	"io"
)

// Analyzer is the type of functions that will be called in
// an analyze() operation
// type Analyzer func(io.ReadSeeker, io.Writer, ...interface{}) error
type Analyzer func(io.ReadSeeker, ...interface{}) (map[string]interface{}, error)

// AnalyzerRegister registers a user analyzer
func AnalyzerRegister(typ string, analyzerfunc Analyzer) {
	analyzersList[typ] = analyzerfunc
}

// AnalyzerHelp outputs list of available analyzers
func AnalyzerHelp() {
	fmt.Printf("Available analyzers are:\n")
	for k := range analyzersList {
		fmt.Printf("\t%s\n", k)
	}
}

var analyzersList = map[string]Analyzer{
	"strings":   analyzers.StringAnalyzer,
	"version":   analyzers.VersionAnalyzer,
	"histogram": analyzers.HistogramAnalyzer,
	"elf":       analyzers.ElfAnalyzer,
	"dex":       analyzers.DexAnalyzer,
}

// writeReport will create a log file with the report data
func writeReport(e *types.Env, name string, report map[string]interface{}) (string, error) {
	w, err := e.CreateLog(name)
	if err != nil {
		return "", err
	}
	defer w.Close()

	bs, err := json.MarshalIndent(report, "", "\t")
	if err != nil {
		return "", err
	}
	w.Write(bs)
	return name, nil
}

// analyzeFunction performs some type of analysis on the current binary
func analyzeFunction(e *types.Env, typ string, prefix string, data ...interface{}) (string, error) {
	f, found := analyzersList[typ]
	if !found {
		AnalyzerHelp()
		util.RegisterFatalf("Unknown analyzer: '%s'", typ)
		return "", fmt.Errorf("Unknown analyzer: '%s'", typ)
	}

	report, err := f(e.Reader, data...)
	if err != nil {
		return "", err
	}
	if report != nil {
		logfile := fmt.Sprintf("%s_%s.json", typ, prefix)
		return writeReport(e, logfile, report)
	}
	return "", nil

}

func init() {
	types.FunctionRegister("analyze", analyzeFunction)
}
