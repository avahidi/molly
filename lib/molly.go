package lib

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"

	_ "bitbucket.org/vahidi/molly/lib/actions" // import default actions
	"bitbucket.org/vahidi/molly/lib/scan"
	"bitbucket.org/vahidi/molly/lib/types"
	"bitbucket.org/vahidi/molly/lib/util"
)

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

func scanReader(env *types.Env, report *types.MatchReport,
	rules *types.RuleSet, r io.ReadSeeker) {
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
func ScanData(config *types.Config, rules *types.RuleSet, data []byte) (
	*types.MatchReport, error) {

	report := types.NewMatchReport()
	env := types.NewEnv(config, util.NewFileQueue())
	env.SetFile("nopath/nofile", uint64(len(data)))
	scanReader(env, report, rules, bytes.NewReader(data))
	return report, nil
}

// ScanFiles scans a set of files for matches against the given rules
// if any files are extracted they will be created within outputDir
func ScanFiles(config *types.Config, rules *types.RuleSet, files []string) (
	*types.MatchReport, int, error) {

	inputs := util.NewFileQueue()
	env := types.NewEnv(config, inputs)

	// add inputs
	for _, file := range files {
		abs, err := filepath.Abs(file)
		if err != nil {
			return nil, 0, err
		}
		inputs.Push(abs)
	}

	report := types.NewMatchReport()
	for filename := inputs.Pop(); filename != ""; filename = inputs.Pop() {
		info, err := os.Stat(filename)
		if err != nil {
			report.Errors = append(report.Errors, err)
			continue
		}
		// open file and scan it
		f, err := os.Open(filename)
		if err != nil {
			report.Errors = append(report.Errors, err)
			continue
		}
		defer f.Close()

		report.Files = append(report.Files, filename)
		env.SetFile(filename, uint64(info.Size()))
		scanReader(env, report, rules, f)

		// close it manually to avoid "too many open files"
		f.Close()
	}

	// populate tagged files
	for _, top := range report.MatchTree {
		tagExtract(report.Tagged, rules.Flat, top)
	}
	// record files generated
	for k, v := range env.Output().Trace {
		if k != "" {
			report.OutHierarchy[k] = v
		}
	}

	// record logs generated
	report.LogHierarchy = env.Log().Trace

	return report, inputs.Count(), nil
}

func tagExtract(tagmap map[string][]string, lookup map[string]*types.Rule, match *types.MatchEntry) {
	rule, found := lookup[match.Rule]
	if !found {
		return
	}

	if tagmeta, valid := rule.Metadata.GetString("tag", ""); valid {
		tags := strings.Split(tagmeta, ",")
		for _, tag := range tags {
			if tag2 := strings.Trim(tag, " \t\n\r"); tag2 != "" {
				tagmap[tag2] = append(tagmap[tag2], match.Filename)
			}
		}
	}

	for _, cm := range match.Children {
		tagExtract(tagmap, lookup, cm)
	}
}
