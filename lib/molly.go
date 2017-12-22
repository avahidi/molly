package lib

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	_ "bitbucket.org/vahidi/molly/lib/actions" // import default actions
	"bitbucket.org/vahidi/molly/lib/scan"
	"bitbucket.org/vahidi/molly/lib/types"
	"bitbucket.org/vahidi/molly/lib/util"
	"bitbucket.org/vahidi/molly/lib/util/logging"
)

type filesystem struct {
	outdir    string
	owner     string
	queue     types.FileQueue
	generated []string
}

func newFilesystem(owner, outdir string, queue types.FileQueue) *filesystem {
	_, owner = filepath.Split(owner)
	return &filesystem{owner: owner, outdir: outdir, queue: queue}
}

func (fs *filesystem) record(path string) {
	fs.generated = append(fs.generated, path)
	fs.queue.Push(path)
}
func (fs *filesystem) Name(suggested string, addtopath bool) string {
	if filepath.HasPrefix(suggested, fs.outdir) {
		suggested = suggested[len(fs.outdir):]
	}

	path := filepath.Join(fs.outdir, util.SanitizeFilename(suggested, nil))
	if addtopath {
		fs.record(path)
	}
	return path
}
func (fs *filesystem) Mkdir(path string) error {
	if !util.PermissionGet(util.CreateFile) {
		return fmt.Errorf("Not allowed to create files (mkdir")
	}
	path = fs.Name(path, false)
	err := os.MkdirAll(path, 0700)
	if err == nil {
		fs.record(path)
	}
	return err
}
func (fs *filesystem) Create(filename string) (*os.File, error) {
	if !util.PermissionGet(util.CreateFile) {
		return nil, fmt.Errorf("Not allowed to create files (mkdir")
	}

	filename = fs.Name(filename, false)

	// make sure the path leading to it exist
	dir, _ := filepath.Split(filename)
	os.MkdirAll(dir, 0700)

	// open the file and record this event
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_EXCL|os.O_TRUNC, 0600)
	if err == nil {
		fs.record(filename)
	}
	return file, err
}

// MatchCallback is the type of function called when a match is found
type MatchCallback func(m *types.MatchEntry)

// ScanRules reads rules from files
func ScanRules(files []string) (*types.RuleSet, error) {
	rs := types.NewRuleSet()
	ins := newfileset()
	ins.Push(files...)
	err := scan.ParseRules(ins, rs)
	return rs, err
}

// ScanFiles scans a set of files for matches against the given rules
// if any files are extracted they will be created within outputDir
func ScanFiles(files []string, rules *types.RuleSet, outputDir string,
	callback MatchCallback) (*types.MatchReport, int, error) {

	// prepare output directory
	outputDir, err := filepath.Abs(outputDir)
	if err != nil {
		logging.Fatal(err)
	}
	if err := os.MkdirAll(outputDir, 0700); err != nil {
		logging.Fatalf("Could not create output directory: %v", err)
	}

	// inputs
	inputs := newfileset()
	for _, file := range files {
		abs, err := filepath.Abs(file)
		if err != nil {
			return nil, 0, err
		}
		inputs.Push(abs)
	}
	// globals := util.NewRegister()
	env := types.NewEnv()
	globals := env.Globals
	report := types.NewMatchReport()

	for filename := inputs.Pop(); filename != ""; filename = inputs.Pop() {
		fs := newFilesystem(filename, outputDir, inputs)

		// prepare env and glonal variables for this file
		env.FileSystem = fs
		info, err := os.Stat(filename)
		if err != nil {
			report.Errors = append(report.Errors, err)
			continue
		}
		dir, name := path.Dir(filename), path.Base(filename)
		globals.SetString("$path", dir)
		globals.SetString("$shortfilename", name)
		globals.SetString("$filename", filename)
		globals.SetNumber("$filesize", info.Size())

		// compute the new output path for anything generated out of this file
		var pathnew = filename
		if filepath.HasPrefix(pathnew, outputDir) {
			pathnew = pathnew[len(outputDir):]
		}
		pathnew = filepath.Join(outputDir, util.SanitizeFilename(pathnew, nil)) + "_"
		globals.Set("$outdir", pathnew)
		fs.outdir = pathnew

		// open file and scan it
		f, err := os.Open(filename)
		if err != nil {
			report.Errors = append(report.Errors, err)
			continue
		}
		defer f.Close()

		env.StartFile(f)
		for _, r := range rules.Top {
			env.StartRule(r)
			match, errs := scan.AnalyzeFile(filename, r, env)
			if match != nil {
				report.MatchTree = append(report.MatchTree, match)
			}
			report.Errors = append(report.Errors, errs...)
		}
		report.FileHierarchy[filename] = fs.generated
	}

	// populate tagged files
	filematch := extractFilesFromReport(report)
	for filename, matches := range filematch {
		tagset := extractTags(matches, rules)
		report.TaggedFiles[filename] = tagset
	}

	// user callback?
	if callback != nil {
		for _, me := range report.MatchTree {
			me.Walk(callback)
		}
	}

	return report, len(inputs.processed), nil
}

// extractFilesFromReport gathers all files that have at least one match
func extractFilesFromReport(report *types.MatchReport) map[string][]*types.MatchEntry {
	files := make(map[string][]*types.MatchEntry)
	for _, me := range report.MatchTree {
		me.Walk(func(mc *types.MatchEntry) {
			hits, _ := files[mc.Filename]
			files[mc.Filename] = append(hits, mc)
		})
	}
	return files
}

// extractTags is a very inefficient way of gathering all tags in a match tree list
func extractTags(matches []*types.MatchEntry, rules *types.RuleSet) []string {
	tagset := make(map[string]bool)
	for _, match := range matches {
		rule := rules.Flat[match.Rule]
		if tagmeta, valid := rule.Metadata.GetString("tag", ""); valid {
			tags := strings.Split(tagmeta, ",")
			for _, tag := range tags {
				if tag2 := strings.Trim(tag, " \t\n\r"); tag2 != "" {
					tagset[tag2] = true
				}
			}
		}
	}

	// convert map to array
	ret := make([]string, 0, len(tagset))
	for k := range tagset {
		ret = append(ret, k)
	}
	return ret
}
