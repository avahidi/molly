package lib

import (
	"bytes"
	"fmt"
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
func New(extratDir, reportDir string, maxDepth int) *types.Molly {
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

	return types.NewMolly(extratDir, reportDir, maxDepth)
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

func scanInput(m *types.Molly, env *types.Env, i *types.Input) {
	env.SetInput(i)
	for pass := types.RulePassMin; pass <= types.RulePassMax; pass++ {
		for _, rule := range m.Rules.Top {
			p, _ := rule.Metadata.GetNumber("pass", uint64(types.RulePassMin))
			if p != uint64(pass) {
				continue
			}
			env.StartRule(rule)
			match, errs := scan.AnalyzeFile(rule, env)
			if match != nil {
				i.Matches = append(i.Matches, match)
				processMatch(m, i, match)
			}
			i.Errors = append(i.Errors, errs...)
		}
	}
	processTags(m, i)
}

// ScanData scans a byte vector for matches.
func ScanData(m *types.Molly, data []byte) (*types.Report, error) {
	fr := types.NewInput("nopath/nofile", int64(len(data)))
	fr.Reader = bytes.NewReader(data)
	env := types.NewEnv(m)
	scanInput(m, env, fr)

	report := types.NewReport()
	if !fr.Empty() {
		report.Add(fr)
	}
	return report, nil
}

// scanFile opens and scans a single file
func scanFile(m *types.Molly, env *types.Env, filename string,
	filesize int64, parent string) *types.Input {
	fr := types.NewInput(filename, filesize)
	fr.Parent = m.Processed[parent]
	if fr.Parent != nil {
		fr.Depth = fr.Parent.Depth + 1
	}
	m.Processed[filename] = fr

	if m.MaxDepth != 0 && fr.Depth >= m.MaxDepth {
		fr.Errors = append(fr.Errors, fmt.Errorf("File depth above %d", m.MaxDepth))
		return fr
	}

	r, err := os.Open(filename)
	if err != nil {
		fr.Errors = append(fr.Errors, err)
		return fr
	}
	defer r.Close()
	fr.Reader = r

	scanInput(m, env, fr)
	return fr
}

// ScanFiles scans a set of files for matches.
func ScanFiles(m *types.Molly, files []string) (*types.Report, error) {
	env := types.NewEnv(m)

	// add inputs
	for _, file := range files {
		abs, err := filepath.Abs(file)
		if err != nil {
			return nil, err
		}
		m.Files.Push(abs)
	}

	report := types.NewReport()
	for {
		filename, size, parent := m.Files.Pop()
		if filename == "" {
			break
		}

		fr := scanFile(m, env, filename, size, parent)
		if !fr.Empty() {
			report.Add(fr)
		}
	}
	return report, nil
}
