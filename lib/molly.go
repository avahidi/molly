package lib

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	_ "bitbucket.org/vahidi/molly/lib/actions" // import default actions
	"bitbucket.org/vahidi/molly/lib/report"
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

// processMatch will process a match on a rule
func processMatch(m *types.Molly, i *types.Input, match *types.Match) {
	if len(match.Children) == 0 {
		if m.OnMatchRule != nil {
			m.OnMatchRule(i, match)
		}
	}
	for _, ch := range match.Children {
		processMatch(m, i, ch)
	}
}

// processMatch will process a tag on a file
func processTags(m *types.Molly, fr *types.Input) {
	if m.OnMatchTag != nil {
		tags := report.ExtractTags(fr)
		for _, tag := range tags {
			m.OnMatchTag(fr, tag)
		}
	}
}

func scanReader(m *types.Molly, env *types.Env, r io.ReadSeeker,
	filename string, filesize uint64) *types.Input {
	input := types.NewInput(r, filename, filesize)
	env.SetInput(input)
	for _, rule := range m.Rules.Top {
		env.StartRule(rule)
		match, errs := scan.AnalyzeFile(rule, env)
		if match != nil {
			input.Matches = append(input.Matches, match)
			processMatch(m, input, match)
		}
		input.Errors = append(input.Errors, errs...)
	}
	processTags(m, input)

	return input
}

// ScanData scans a byte vector for matches.
func ScanData(m *types.Molly, data []byte) (*types.Report, error) {
	dummyname := "nopath/nofile"
	r := bytes.NewReader(data)
	env := types.NewEnv(m)

	fr := scanReader(m, env, r, dummyname, uint64(len(data)))

	report := types.NewReport()
	if !fr.Empty() {
		report.Add(fr)
	}
	return report, nil
}

// scanFile opens and scans a single file
func scanFile(m *types.Molly, env *types.Env, filename string) *types.Input {
	info, err := os.Stat(filename)
	if err != nil {
		fr := types.NewInput(nil, filename, 0)
		fr.Errors = append(fr.Errors, err)
		return fr
	}
	r, err := os.Open(filename)
	if err != nil {
		fr := types.NewInput(nil, filename, 0)
		fr.Errors = append(fr.Errors, err)
		return fr
	}
	defer r.Close()

	return scanReader(m, env, r, filename, uint64(info.Size()))
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
		fr := scanFile(m, env, filename)
		if !fr.Empty() {
			report.Add(fr)
		}
	}
	return report, len(m.Files.Out), nil
}
