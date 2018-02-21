package lib

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	_ "bitbucket.org/vahidi/molly/lib/actions" // import default actions
	"bitbucket.org/vahidi/molly/lib/scan"
	"bitbucket.org/vahidi/molly/lib/types"
	"bitbucket.org/vahidi/molly/lib/util"
)

// New creates a new molly context
func New(extratDir, reportDir string) *types.Molly {
	if extratDir == "" {
		extratDir, _ = ioutil.TempDir("", "molly-out")
	}
	if reportDir == "" {
		reportDir, _ = ioutil.TempDir("", "molly-log")
	}
	if err := util.Mkdir(extratDir); err != nil {
		util.RegisterFatalf("Failed to create extraction dir: %v", err)
	}
	if err := util.Mkdir(reportDir); err != nil {
		util.RegisterFatalf("Failed to create report dir: %v", err)
	}

	return types.NewMolly(extratDir, reportDir)
}

// LoadRules reads rules from files
func LoadRules(m *types.Molly, files ...string) error {
	return scan.ParseRuleFiles(m, files...)
}

// LoadRulesFromText reads rules from a string
func LoadRulesFromText(m *types.Molly, text string) error {
	return scan.ParseRuleStream(m, strings.NewReader(text))
}

func scanReader(env *types.Env, rules *types.RuleSet, r io.ReadSeeker,
	filename string, filesize uint64, report *types.FileReport) {
	env.SetFile(r, filename, filesize, report)
	for _, rule := range rules.Top {
		env.StartRule(rule)
		match, errs := scan.AnalyzeFile(rule, env)
		if match != nil {
			report.Matches = append(report.Matches, match)
		}
		report.Errors = append(report.Errors, errs...)
	}
}

// ScanData scans a byte vector for matches.
func ScanData(m *types.Molly, data []byte) (*types.Report, error) {
	dummyname := "nopath/nofile"
	fr := types.NewFileReport(dummyname)
	env := types.NewEnv(m)
	scanReader(env, m.Rules, bytes.NewReader(data), dummyname, uint64(len(data)), fr)

	report := types.NewReport()
	if !fr.Empty() {
		report.Add(fr)
	}
	return report, nil
}

// scanFile opens and scans a single file
func scanFile(m *types.Molly, env *types.Env, fr *types.FileReport, filename string ) {
	info, err := os.Stat(filename)
	if err != nil {
		fr.Errors = append(fr.Errors, err)
		return
	}
	f, err := os.Open(filename)
	if err != nil {
		fr.Errors = append(fr.Errors, err)
		return
	}
	defer f.Close()
	scanReader(env, m.Rules, f, filename, uint64(info.Size()), fr)
}

// ScanFiles scans a set of files for matches.
func ScanFiles(m *types.Molly, files []string) (*types.Report, int, error) {
	env := types.NewEnv(m)

	// add inputs
	for _, file := range files {
		abs, err := filepath.Abs(file)
		if err != nil {
			return nil, 0, err
		}
		m.Files.Push(abs)
	}

	report := types.NewReport()
	for filename := m.Files.Pop(); filename != ""; filename = m.Files.Pop() {
		fr := types.NewFileReport(filename)
		scanFile(m, env, fr, filename)
		if !fr.Empty() {
			report.Add(fr)
		}
	}
	return report, len(m.Files.Out), nil
}
