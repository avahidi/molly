package lib

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "bitbucket.org/vahidi/molly/actions" // import default actions
	"bitbucket.org/vahidi/molly/report"
	"bitbucket.org/vahidi/molly/scan"
	"bitbucket.org/vahidi/molly/types"
	"bitbucket.org/vahidi/molly/util"
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
		util.RegisterFatalf("Failed to create output directory: %v", err)
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

func scanInput(m *types.Molly, env *types.Env, r *types.Report, i *types.Input) {
	// update basename to something we can use to create files from
	if i.Parent == nil {
		i.FilenameOut = suggestBaseName(m, i)

		// make sure its path is there and we have a soft link to the real file
		path, _ := filepath.Split(i.FilenameOut)
		util.SafeMkdir(path)

		// make sure we link to the absolute path
		filename_abs, _ := filepath.Abs(i.Filename)
		os.Symlink(filename_abs, i.FilenameOut)
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

	// this file might have generated a report, so log it
	if !i.Empty() {
		r.Add(i)
	}
	// this file may have created new files, scan them too
	for _, offspring := range i.Children {
		scanFile(m, env, r, offspring, i)
	}
}

// ScanData scans a byte vector for matches.
func ScanData(m *types.Molly, data []byte) (*types.Report, error) {
	fr := types.NewInput(nil, "nopath/nofile", int64(len(data)), time.Now())
	fr.Reader = bytes.NewReader(data)
	env := types.NewEnv(m)
	report := types.NewReport()
	scanInput(m, env, report, fr)

	return report, nil
}

// scanFile opens and scans a single file
func scanFile(m *types.Molly, env *types.Env, rep *types.Report,
	filename_ string, parent *types.Input) {

	fl := &util.FileList{}
	fl.Push(filename_)
	for {
		filename, fi, err := fl.Pop()
		if filename == "" {
			return
		}

		if _, found := m.Processed[filename]; found {
			continue // how can this even happen?
		}

		fr := types.NewInput(parent, filename, fi.Size(), fi.ModTime())
		m.Processed[filename] = fr

		// started with an error, no point moving in
		if err != nil {
			fr.Errors = append(fr.Errors, err)
			continue
		}

		if m.MaxDepth != 0 && fr.Depth >= m.MaxDepth {
			fr.Errors = append(fr.Errors, fmt.Errorf("File depth above %d", m.MaxDepth))
			continue
		}

		r, err := os.Open(filename)
		if err != nil {
			fr.Errors = append(fr.Errors, err)
		}

		fr.Reader = r
		scanInput(m, env, rep, fr)

		// manual Close insted of defer Close, or we will have too many files open
		r.Close()
	}
}

// ScanFiles scans a set of files for matches.
func ScanFiles(m *types.Molly, files []string) (*types.Report, error) {
	env := types.NewEnv(m)
	report := types.NewReport()

	for _, filename := range files {
		scanFile(m, env, report, filename, nil)
	}

	return report, nil
}
