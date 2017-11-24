package lib

import (
	"os"
	"path"
	"path/filepath"
	"strings"

	_ "bitbucket.org/vahidi/molly/lib/actions" // import default actions
	"bitbucket.org/vahidi/molly/lib/exp"
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
	path = fs.Name(path, false)
	err := os.MkdirAll(path, 0700)
	if err == nil {
		fs.record(path)
	}
	return err
}
func (fs *filesystem) Create(filename string) (*os.File, error) {
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

// Molly is the main structure and contains the rule database and configuration
type Molly struct {
	Globals *util.Register
	RuleSet *types.RuleSet

	matchCallback MatchCallback
	outputDir     string

	// these are valid only during scanning
	currInputs types.InputSet
	currOutpus *types.MatchReport
}

func New(outputDir string, callback MatchCallback) *Molly {
	outputDir, err := filepath.Abs(outputDir)
	if err != nil {
		logging.Fatal(err)
	}
	if err := os.MkdirAll(outputDir, 0700); err != nil {
		logging.Fatalf("Could not create output directory: %v", err)
	}

	rs := &types.RuleSet{
		Files: make(map[string][]types.Rule),
		Top:   make(map[string]types.Rule),
		Flat:  make(map[string]types.Rule),
	}

	m := &Molly{
		Globals:       util.NewRegister(),
		RuleSet:       rs,
		outputDir:     outputDir,
		matchCallback: callback,
	}
	m.Globals.Set("$outputdir", outputDir)
	return m
}

func (m *Molly) ScanRules(files []string) error {
	ins := newinputset()
	ins.Push(files...)
	return scan.ParseRules(ins, m.RuleSet)
}

func (m *Molly) ScanFiles(files []string) (types.InputSet, *types.MatchReport, error) {
	env := exp.NewEnvironment(m.Globals)
	m.currInputs = newinputset()
	m.currOutpus = types.NewMatchReport()

	for _, file := range files {
		abs, err := filepath.Abs(file)
		if err != nil {
			return nil, nil, err
		}
		m.currInputs.Push(abs)
	}

	for filename := m.currInputs.Pop(); filename != ""; filename = m.currInputs.Pop() {
		fs := newFilesystem(filename, m.outputDir, m.currInputs)

		env.SetFileSystem(fs)
		info, err := os.Stat(filename)
		if err != nil {
			m.currOutpus.Errors = append(m.currOutpus.Errors, err)
			continue
		}

		f, err := os.Open(filename)
		if err != nil {
			m.currOutpus.Errors = append(m.currOutpus.Errors, err)
			continue
		}
		defer f.Close()

		// set the global variables for this file
		dir, name := path.Dir(filename), path.Base(filename)
		m.Globals.SetString("$path", dir)
		m.Globals.SetString("$shortfilename", name)
		m.Globals.SetString("$filename", filename)
		m.Globals.SetNumber("$filesize", info.Size())

		// compute the new output path for anything generated out of this file
		var pathnew = filename
		if filepath.HasPrefix(pathnew, m.outputDir) {
			pathnew = pathnew[len(m.outputDir):]
		}
		pathnew = filepath.Join(m.outputDir, util.SanitizeFilename(pathnew, nil)) + "_"
		m.Globals.Set("$outdir", pathnew)
		fs.outdir = pathnew

		env.StartFile(f)
		for _, r := range m.RuleSet.Top {
			env.StartRule(r)
			match, errs := scan.AnalyzeFile(filename, r, env)
			if match != nil {
				m.currOutpus.MatchTree = append(m.currOutpus.MatchTree, match)
			}
			m.currOutpus.Errors = append(m.currOutpus.Errors, errs...)
		}
		m.currOutpus.FileHierarchy[filename] = fs.generated
	}

	// populate tagged files
	// 1. get a list of all files and their matches
	filematch := make(map[string][]*types.MatchEntry)
	for _, me := range m.currOutpus.MatchTree {
		me.Walk(func(mc *types.MatchEntry) {
			hits, _ := filematch[mc.Filename]
			filematch[mc.Filename] = append(hits, mc)
		})
	}
	// 2. for each matched files, get the tags
	// get a list of all files and their matches
	for filename, matches := range filematch {
		tagset := make(map[string]bool)
		for _, match := range matches {
			rule := m.RuleSet.Flat[match.Rule]
			if tagmeta, valid := rule.GetMetadata().GetString("tag", ""); valid {
				tags := strings.Split(tagmeta, ",")
				for _, tag := range tags {
					if tag2 := strings.Trim(tag, " \t\n\r"); tag2 != "" {
						tagset[tag2] = true
					}
				}
			}
		}
		if len(tagset) != 0 {
			asarray := []string{}
			for tag := range tagset {
				asarray = append(asarray, tag)
			}
			m.currOutpus.TaggedFiles[filename] = asarray
		}
	}

	// user callback?
	if m.matchCallback != nil {
		for _, me := range m.currOutpus.MatchTree {
			me.Walk(m.matchCallback)
		}
	}

	return m.currInputs, m.currOutpus, nil
}
