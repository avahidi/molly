package operators

import (
	"fmt"
	"io"

	"bitbucket.org/vahidi/molly/operators/analyzers"
	"bitbucket.org/vahidi/molly/types"
	"bitbucket.org/vahidi/molly/util"
)

// AnalyzerRegister registers a user analyzer
func AnalyzerRegister(typ string, f func(filename string, r io.ReadSeeker, data ...interface{}) (interface{}, error)) {
	analyzersList[typ] = f
}

// AnalyzerHelp outputs list of available analyzers
func AnalyzerHelp() {
	fmt.Printf("Available analyzers are:\n")
	for k := range analyzersList {
		fmt.Printf("\t%s\n", k)
	}
}

var analyzersList = map[string]func(string, io.ReadSeeker, ...interface{}) (interface{}, error){
	"strings":   analyzers.StringAnalyzer,
	"version":   analyzers.VersionAnalyzer,
	"histogram": analyzers.HistogramAnalyzer,
	"elf":       analyzers.ElfAnalyzer,
	"dex":       analyzers.DexAnalyzer,
}

// analyzeFunction performs some type of analysis on the current binary
func analyzeFunction(e *types.Env, typ string, prefix string, data ...interface{}) (string, error) {
	f, found := analyzersList[typ]
	if !found {
		AnalyzerHelp()
		util.RegisterFatalf("Unknown analyzer: '%s'", typ)
		return "", fmt.Errorf("Unknown analyzer: '%s'", typ)
	}

	// do analysis
	res, err := f(e.GetFile(), e.Reader, data...)

	// generate analysis name based on type and parameters
	name := typ
	for _, d := range data {
		name = fmt.Sprintf("%s__%v", name, d)
	}

	// register result
	e.Current.RegisterAnalysis(name, res, err)
	return "", err
}

func init() {
	Register("analyze", analyzeFunction)
}
