package actions

import (
	"fmt"

	"bitbucket.org/vahidi/molly/actions/analyzers"
	"bitbucket.org/vahidi/molly/types"
	"bitbucket.org/vahidi/molly/util"
)

// AnalyzerRegister registers a user analyzer
func AnalyzerRegister(typ string, analyzerfunc analyzers.Analyzer) {
	analyzersList[typ] = analyzerfunc
}

// AnalyzerHelp outputs list of available analyzers
func AnalyzerHelp() {
	fmt.Printf("Available analyzers are:\n")
	for k := range analyzersList {
		fmt.Printf("\t%s\n", k)
	}
}

var analyzersList = map[string]analyzers.Analyzer{
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

	res := analyzers.NewAnalysis(typ, nil, data)

	f(e.GetFile(), e.Reader, res, data...)
	e.Current.Analyses[res.Name] = res
	return "", res.Error
}

func init() {
	ActionRegister("analyze", analyzeFunction)
}
