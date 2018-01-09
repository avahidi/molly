package lib

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	_ "bitbucket.org/vahidi/molly/lib/actions" // import default actions
	"bitbucket.org/vahidi/molly/lib/scan"
	"bitbucket.org/vahidi/molly/lib/types"
	"bitbucket.org/vahidi/molly/lib/util"
)

// Config defines the configuration for scanning files
type Config struct {
	NewFile       func(string) (string, error)
	MatchCallback func(m *types.MatchEntry)

	// temporary variables set duing scanning
	queue     *util.FileQueue
	env       *types.Env
	report    *types.MatchReport
	generated []string
}

func (c *Config) recordFile(name string) {
	c.generated = append(c.generated, name)
	c.queue.Push(name)
}
func (c *Config) Name(suggested string, addtopath bool) (string, error) {
	newname, err := c.NewFile(suggested)
	if err != nil {
		return "", err
	}
	if addtopath {
		c.recordFile(newname)
	}
	return newname, nil
}
func (c *Config) Mkdir(path string) error {
	if !util.PermissionGet(util.CreateFile) {
		return fmt.Errorf("Not allowed to create files (mkdir")
	}
	newpath, err := c.Name(path, false)
	if err == nil {
		err = os.MkdirAll(newpath, 0700)
		if err == nil {
			c.recordFile(newpath)
		}
	}
	return err
}
func (c *Config) Create(filename string) (*os.File, error) {
	if !util.PermissionGet(util.CreateFile) {
		return nil, fmt.Errorf("Not allowed to create files (mkdir")
	}
	newname, err := c.Name(filename, false)
	if err != nil {
		return nil, err
	}

	// make sure the path leading to it exist
	dir, _ := filepath.Split(newname)
	os.MkdirAll(dir, 0700)

	// open the file and record this event
	file, err := os.OpenFile(newname, os.O_WRONLY|os.O_CREATE|os.O_EXCL|os.O_TRUNC, 0600)
	if err == nil {
		c.recordFile(newname)
	}
	return file, err
}

// NewConfig create a new config
func NewConfig() *Config {
	return &Config{
		NewFile: func(_ string) (string, error) { return "", fmt.Errorf("NewFile was not set in configuration") },
		queue:   util.NewFileQueue(),
		report:  types.NewMatchReport(),
		env:     types.NewEnv(),
	}
}

// FileSystem implementation
var _ types.FileSystem = (*Config)(nil)

// LoadRules reads rules from files
func LoadRules(db *types.RuleSet, files ...string) (*types.RuleSet, error) {
	if db == nil {
		db = types.NewRuleSet()
	}
	return db, scan.ParseRuleFiles(db, files...)
}

// LoadRuleText reads rules from a string
func LoadRuleText(db *types.RuleSet, text string) (*types.RuleSet, error) {
	if db == nil {
		db = types.NewRuleSet()
	}
	return db, scan.ParseRuleStream(db, strings.NewReader(text))
}

/*
func scanStream(config *Config, rules *types.RuleSet, data []byte) error {
	return nil
}

func ScanData(config *Config, rules *types.RuleSet, data []byte) (*types.MatchReport, error) {

	return nil, nil
}
*/

func scanReader(config *Config, rules *types.RuleSet, r io.ReadSeeker) {
	report := config.report
	env := config.env
	env.Reader = r

	for _, rule := range rules.Top {
		env.StartRule(rule)
		match, errs := scan.AnalyzeFile(rule, env)
		if match != nil {
			report.MatchTree = append(report.MatchTree, match)
		}
		report.Errors = append(report.Errors, errs...)
	}
}

// ScanData scans a a data stream for matches against the given rules
// if any files are extracted they will be created within outputDir
func ScanData(config *Config, rules *types.RuleSet, data []byte) (
	*types.MatchReport, error) {
	if config == nil {
		config = NewConfig()
	}

	env := config.env
	globals := env.Globals
	globals.SetString("$path", "nopath/")
	globals.SetString("$shortfilename", "nofile")
	globals.SetString("$filename", "nopath/nofile")
	globals.SetNumber("$filesize", uint64(len(data)))

	scanReader(config, rules, bytes.NewReader(data))

	// user callback?
	if config.MatchCallback != nil {
		for _, me := range config.report.MatchTree {
			me.Walk(config.MatchCallback)
		}
	}

	return config.report, nil
}

// ScanFiles scans a set of files for matches against the given rules
// if any files are extracted they will be created within outputDir
func ScanFiles(config *Config, rules *types.RuleSet, files []string) (
	*types.MatchReport, int, error) {
	if config == nil {
		config = NewConfig()
	}

	// add inputs
	inputs := config.queue
	for _, file := range files {
		abs, err := filepath.Abs(file)
		if err != nil {
			return nil, 0, err
		}
		inputs.Push(abs)
	}

	/*
		// prepare output directory
		outputDir, err := filepath.Abs(outputDir)
		if err != nil {
			util.RegisterFatal(err)
		}
		if err := os.MkdirAll(outputDir, 0700); err != nil {
			util.RegisterFatalf("Could not create output directory: %v", err)
		}
	*/

	env := config.env
	report := config.report
	env.FileSystem = config
	globals := env.Globals
	for filename := inputs.Pop(); filename != ""; filename = inputs.Pop() {

		info, err := os.Stat(filename)
		if err != nil {
			report.Errors = append(report.Errors, err)
			continue
		}
		dir, name := path.Dir(filename), path.Base(filename)

		globals.SetString("$path", dir)
		globals.SetString("$shortfilename", name)
		globals.SetString("$filename", filename)
		globals.SetNumber("$filesize", uint64(info.Size()))

		/*
			// compute the new output path for anything generated out of this file
			var pathnew = filename
			if filepath.HasPrefix(pathnew, outputDir) {
				pathnew = pathnew[len(outputDir):]
			}
			pathnew = filepath.Join(outputDir, util.SanitizeFilename(pathnew, nil)) + "_"
			globals.Set("$outdir", pathnew)
			fs.outdir = pathnew
		*/
		// open file and scan it
		f, err := os.Open(filename)
		if err != nil {
			report.Errors = append(report.Errors, err)
			continue
		}
		defer f.Close()

		config.generated = nil
		scanReader(config, rules, f)
		report.FileHierarchy[filename] = config.generated
	}

	// populate tagged files
	filematch := extractFilesFromReport(report)
	for filename, matches := range filematch {
		tagset := extractTags(matches, rules)
		report.TaggedFiles[filename] = tagset
	}

	// user callback?
	if config.MatchCallback != nil {
		for _, me := range report.MatchTree {
			me.Walk(config.MatchCallback)
		}
	}

	return report, inputs.Count(), nil
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
