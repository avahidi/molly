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

// newEmptyDir accepts a new or empty dir and if new creates it
func newEmptyDir(dirname string) error {
	typ := util.GetPathType(dirname)
	switch typ {
	case util.File:
		return fmt.Errorf("'%s' is a file", dirname)
	case util.NonEmptyDir:
		return fmt.Errorf("'%s' exists and is not empty", dirname)
	case util.Error:
		return fmt.Errorf("'%s' could not be checked", dirname)
	default:
		return util.Mkdir(dirname)
	}
}

// suggestBaseName picks a new base name for a file
func suggestBaseName(m *types.Molly, input *types.Input) string {
	for i := 0; ; i++ {
		basename := filepath.Join(m.OutDir, util.SanitizeFilename(input.Filename))
		if i != 0 {
			basename = fmt.Sprintf("%s_%04d", basename, i)
		}
		if util.GetPathType(basename) == util.NoFile && util.GetPathType(basename+"_") == util.NoFile {
			return basename
		}
	}
}

// New creates a new molly context
func New(outdir string, maxDepth int) *types.Molly {
	if outdir == "" {
		outdir, _ = ioutil.TempDir("", "molly-out")
	}
	if err := newEmptyDir(outdir); err != nil {
		util.RegisterFatalf("Failed to create output dir: %v", err)
	}
	return types.NewMolly(outdir, maxDepth)
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
	// update basename to something we can use to create files from
	if i.Parent == nil {
		i.Basename = suggestBaseName(m, i)

		// make sure its path is there and we have a soft link to the real file
		path, _ := filepath.Split(i.Basename)
		util.SafeMkdir(path)
		os.Symlink(i.Filename, i.Basename)
	}

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
	m.Files.Push(files...)

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
