package lib

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

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
func suggestBaseName(m *types.Molly, input *types.FileData) string {
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
func processMatch(m *types.Molly, i *types.FileData, match *types.Match) {
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
func processTags(m *types.Molly, fr *types.FileData) {
	if m.OnMatchTag != nil {
		tags := report.ExtractTags(fr)
		for _, tag := range tags {
			m.OnMatchTag(fr, tag)
		}
	}
}

func scanInput(m *types.Molly, env *types.Env, r *types.Report,
	reader io.ReadSeeker, data *types.FileData) {
	// update basename to something we can use to create files from
	if data.Parent == nil {
		data.FilenameOut = suggestBaseName(m, data)

		// make sure its path is there and we have a soft link to the real file
		path, _ := filepath.Split(data.FilenameOut)
		util.SafeMkdir(path)

		// make sure we link to the absolute path
		filename_abs, _ := filepath.Abs(data.Filename)
		os.Symlink(filename_abs, data.FilenameOut)
	}

	env.SetInput(reader, data)
	for pass := types.RulePassMin; pass <= types.RulePassMax; pass++ {
		for _, rule := range m.Rules.Top {
			p, _ := rule.Metadata.GetNumber("pass", uint64(types.RulePassMin))
			if p != uint64(pass) {
				continue
			}
			env.StartRule(rule)
			match, errs := scan.AnalyzeFile(rule, env)
			if match != nil {
				data.Matches = append(data.Matches, match)
				processMatch(m, data, match)
			}
			data.Errors = append(data.Errors, errs...)
		}
	}
	processTags(m, data)

	// this file might have generated a report, so log it
	if !data.Empty() {
		r.Add(data)
	}
	// this file may have created new files, scan them too
	for _, offspring := range data.Children {
		scanFile(m, env, r, offspring.Filename, data)
	}
}

// ScanData scans a byte vector for matches.
func ScanData(m *types.Molly, data []byte) (*types.Report, error) {

	// we need a dummy file name that is unique:
	var fd *types.FileData
	for i := 0; fd == nil; i++ {
		dummyname := fmt.Sprintf("nopath/nofile_%04d", i)
		if _, found := m.Files[dummyname]; !found {
			fd = types.NewFileData(dummyname, nil)
			m.Files[dummyname] = fd
		}
	}
	fd.Filesize = int64(len(data))

	env := types.NewEnv(m)
	report := types.NewReport()
	reader := bytes.NewReader(data)
	scanInput(m, env, report, reader, fd)

	return report, nil
}

// scanFile opens and scans a single file
func scanFile(m *types.Molly, env *types.Env, rep *types.Report,
	filename_ string, parent *types.FileData) {

	fl := &util.FileList{}
	fl.Push(filename_)
	for {
		filename, fi, err := fl.Pop()
		if filename == "" {
			return
		}

		fr, found := m.Files[filename]
		if !found {
			fr = types.NewFileData(filename, parent)
			m.Files[filename] = fr
		}

		// if we for some reason have done this one before, just skip it
		if fr.Processed {
			continue
		}
		fr.Processed = true

		// started with an error, no point moving in
		if err != nil {
			fr.Errors = append(fr.Errors, err)
			continue
		}

		// record what we know about it so far
		fr.SetTime(fi.ModTime())
		fr.Filesize = fi.Size()

		if m.MaxDepth != 0 && fr.Depth >= m.MaxDepth {
			fr.Errors = append(fr.Errors, fmt.Errorf("File depth above %d", m.MaxDepth))
			continue
		}

		reader, err := os.Open(filename)
		if err != nil {
			fr.Errors = append(fr.Errors, err)
		}

		scanInput(m, env, rep, reader, fr)

		// manual Close insted of defer Close, or we will have too many files open
		reader.Close()

		// now that the file is closed, attempt to adjust its time
		if t := fr.GetTime(); t != fi.ModTime() {
			os.Chtimes(fr.FilenameOut, t, t)
		}
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
