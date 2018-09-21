package lib

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"bitbucket.org/vahidi/molly/actions/analyzers"

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
			for _, err := range errs {
				data.RegisterError(err)
			}
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

// checkDuplicate if a file is duplicate and does all the book keeping
func checkDuplicate(m *types.Molly, file *types.FileData) (bool, error) {
	hash, err := util.HashFile(file.Filename)
	if err != nil {
		return false, err
	}

	hashtxt := hex.EncodeToString(hash)

	// seen it has been done, add the checksum to our analysis
	file.Analyses["checksum"] = analyzers.NewAnalysis("checksum", hashtxt)

	// check if we have already seen this checksum:
	org, alreadyseen := m.FilesByHash[hashtxt]
	if !alreadyseen {
		m.FilesByHash[hashtxt] = file
		return false, nil
	}

	// record it
	file.DuplicateOf = org
	file.RegisterWarning("duplicate of %s", org.Filename)
	return true, nil
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

			// update basename to something we can use to create files from
			if fr.Parent == nil {
				fr.FilenameOut = suggestBaseName(m, fr)

				// make sure its path is there and we have a soft link to the real file
				path, _ := filepath.Split(fr.FilenameOut)
				util.SafeMkdir(path)

				// make sure we link to the absolute path
				filename_abs, _ := filepath.Abs(fr.Filename)
				os.Symlink(filename_abs, fr.FilenameOut)
			}
		}

		// if we for some reason have done this one before, just skip it
		if fr.Processed {
			continue
		}
		fr.Processed = true

		// started with an error, no point moving in
		if err != nil {
			fr.RegisterError(err)
			continue
		}

		// record what we know about it so far
		fr.SetTime(fi.ModTime())
		fr.Filesize = fi.Size()

		if m.MaxDepth != 0 && fr.Depth >= m.MaxDepth {
			fr.RegisterErrorf("File depth above %d", m.MaxDepth)
			continue
		}

		alreadyseen, err := checkDuplicate(m, fr)
		if err != nil {
			fr.RegisterError(err)
		}
		if alreadyseen {
			// if we already have this guy, just delete it (assuming its ours)
			// XXX: this does not follow the extraction hierarchy
			if fr.Parent != nil {
				os.Remove(fr.FilenameOut)
				os.Symlink(fr.DuplicateOf.Filename, fr.FilenameOut)
			}
			continue
		}

		reader, err := os.Open(filename)
		if err != nil {
			fr.RegisterError(err)
			continue
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
