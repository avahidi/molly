package operators

import (
	"fmt"
	"io"
	"log"

	"bitbucket.org/vahidi/molly/operators/analyzers"
	"bitbucket.org/vahidi/molly/types"
	"bitbucket.org/vahidi/molly/util"
)

func analyzerPrototype(ilename string, r io.ReadSeeker, data ...interface{}) (interface{}, error) {
	return nil, nil
}

var analyzersList *util.FunctionDatabase

// AnalyzerRegister registers a user analyzer
func AnalyzerRegister(typ string, fun interface{}) error {
	err := analyzersList.Register(typ, fun)
	if err != nil {
		util.RegisterFatalf("Failed to register analyzer function: %v", err)
	}
	return err

}

// AnalyzerHelp outputs list of available analyzers
func AnalyzerHelp() {
	fmt.Printf("Available analyzers are:\n")
	for _, v := range analyzersList.Functions {
		ins, outs := v.Signature(true), v.Signature(false)
		fmt.Printf("\t%-12s (%s) -> %s\n", v, ins, outs)
	}
}

// analyzeFunction performs some type of analysis on the current binary
func analyzeFunction(e *types.Env, typ string, prefix string, data ...interface{}) (string, error) {
	f, found := analyzersList.Find(typ)
	if !found {
		AnalyzerHelp()
		util.RegisterFatalf("Unknown analyzer: '%s'", typ)
		return "", fmt.Errorf("Unknown analyzer: '%s'", typ)
	}

	// build parameter list
	ps := make([]interface{}, len(data)+2)
	ps[0] = e.GetFile()
	ps[1] = e.Reader
	for i, d := range data {
		ps[i+2] = d
	}
	// do analysis
	// res, err := f(e.GetFile(), e.Reader, data...)
	res, err := f.Call(ps)

	// generate analysis name based on type and parameters
	name := typ
	for _, d := range data {
		name = fmt.Sprintf("%s__%v", name, d)
	}

	// call didn't return an error but operator did
	if err != nil && res[1] != nil {
		err = res[1].(error)
	}

	// register result
	e.Current.RegisterAnalysis(name, res[0], err)
	return "", err
}

func init() {
	var err error
	analyzersList, err = util.NewFunctionDatabase(analyzerPrototype)
	if err != nil {
		log.Panic(err)
	}

	Register("analyze", analyzeFunction)

	AnalyzerRegister("strings", analyzers.StringAnalyzer)
	AnalyzerRegister("version", analyzers.VersionAnalyzer)
	AnalyzerRegister("histogram", analyzers.HistogramAnalyzer)
	AnalyzerRegister("elf", analyzers.ElfAnalyzer)
	AnalyzerRegister("dex", analyzers.DexAnalyzer)
}
