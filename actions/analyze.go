package actions

import (
	"fmt"
	"log"

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

func (l *logContext) newLog(name string, data interface{}) {

	// XXX: assuming two logs wont have the same name:
	var logname string
	if name != "" {
		logname = fmt.Sprintf("%s_%s", l.basename, name)
	} else {
		logname = l.basename
	}

	switch v := data.(type) {
	case []byte:
		w, err := l.env.CreateLog(logname)
		if err != nil {
			l.error(err)
			return
		}
		defer w.Close()
		w.Write(v)
		l.env.Current.Information[logname] = fmt.Sprintf("see external file '%s'", w.Name())

	case map[string]interface{}:
		l.env.Current.Information[logname] = v
		log.Println("DATA interface", logname)
	default:
		l.error(fmt.Errorf("Unknown analyzer result, in format %t", v))
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
	err := f(e.GetFile(), e.Reader, ctx.newLog, data...)
	if err != nil {
		return "", err
	}
	return "", ctx.err
}

func init() {
	ActionRegister("analyze", analyzeFunction)
}
