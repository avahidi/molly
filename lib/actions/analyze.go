package actions

import (
	"encoding/json"
	"fmt"

	"bitbucket.org/vahidi/molly/lib/actions/analyzers"
	"bitbucket.org/vahidi/molly/lib/types"
	"bitbucket.org/vahidi/molly/lib/util"
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

type logContext struct {
	env      *types.Env
	basename string
	err      error
}

func (l *logContext) error(err error) {
	if l.err == nil {
		l.err = err
	}
}

func (l *logContext) newLog(name string, typ string, data interface{}) {
	var filename string
	if name != "" {
		filename = fmt.Sprintf("%s_%s.%s", l.basename, name, typ)
	} else {
		filename = fmt.Sprintf("%s.%s", l.basename, typ)
	}

	w, err := l.env.CreateLog(filename)
	if err != nil {
		l.error(err)
		return
	}
	defer w.Close()

	switch typ {
	case "json":
		bs, err := json.MarshalIndent(data, "", "\t")
		if err != nil {
			l.error(err)
			return
		}
		w.Write(bs)
	default:
		bs, isbytes := data.([]byte) // just some binary data?
		if isbytes {
			w.Write(bs)
			return
		}
		l.error(fmt.Errorf("Unknown log format: %s", typ))
	}
}

// analyzeFunction performs some type of analysis on the current binary
func analyzeFunction(e *types.Env, typ string, prefix string, data ...interface{}) (string, error) {
	f, found := analyzersList[typ]
	if !found {
		AnalyzerHelp()
		util.RegisterFatalf("Unknown analyzer: '%s'", typ)
		return "", fmt.Errorf("Unknown analyzer: '%s'", typ)
	}

	ctx := logContext{env: e, basename: prefix}
	if ctx.basename == "" {
		ctx.basename = typ
	}
	err := f(e.Input, ctx.newLog, data...)
	if err != nil {
		return "", err
	}
	return "", ctx.err
}

func init() {
	ActionRegister("analyze", analyzeFunction)
}
