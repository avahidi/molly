package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	"github.com/avahidi/molly"
	"github.com/avahidi/molly/operators"
	"github.com/avahidi/molly/types"
	"github.com/avahidi/molly/util"
)

// MultiFlag is used allow multiple values with flag:
type MultiFlag []string

// String implements flag.Value.String() for Multiflag
func (mf MultiFlag) String() string {
	return strings.Join(mf, " ")
}

// Set implements flag.Value.Set() for Multiflag
func (mf *MultiFlag) Set(val string) error {
	*mf = append(*mf, val)
	return nil
}

// flags and usage
var showVersion = flag.Bool("version", false, "show version number")
var showhelp = flag.Bool("h", false, "help information")
var showhelpExt = flag.Bool("H", false, "extended help information")
var outdir = flag.String("o", "output", "output directory")
var rfiles, rtexts, tagops, matchops MultiFlag

var params MultiFlag

func init() {
	flag.Var(&rfiles, "R", "rule files")
	flag.Var(&rtexts, "r", "inline rule")
	flag.Var(&tagops, "on-tag", "tag match operations")
	flag.Var(&matchops, "on-rule", "rule match operations")
	flag.Var(&params, "p", "set parameter")
}

func help(extended bool, errmsg string, exitcode int) {
	if errmsg != "" {
		fmt.Printf("%s\n", errmsg)
	}

	maj, min, mnt := molly.Version()
	fmt.Printf("This is Molly version %d.%d.%d\n", maj, min, mnt)

	flag.Usage()

	fmt.Printf("  files\n\tinput files to be scanned\n")

	if extended {
		operators.Help()
	}
	os.Exit(exitcode)
}

// getStandardRules attempts to find the standard rules included in molly.
// We don't know how molly is installed on this system so we will try a few
// different paths to see if we can find any rules
func getStandardRules() string {
	gopath := os.Getenv("GOPATH")
	home := os.Getenv("HOME")
	dirs := []string{
		path.Join(gopath, "src/github.com/avahidi/molly/data/rules"),
		path.Join(home, "go/src/github.com/avahidi/molly/data/rules"),
		"/usr/share/molly/rules",
		"/usr/lib/molly/rules",
		"data/rules",
	}

	for _, dir := range dirs {
		if fi, err := os.Stat(dir); err == nil && fi.IsDir() {
			return dir
		}
	}
	return ""
}

func main() {
	// 	parse arguments
	flag.Parse()

	if *showhelpExt || *showhelp {
		help(*showhelpExt, "", 0)
	}

	if *showVersion {
		maj, min, mnt := molly.Version()
		fmt.Printf("%d.%d.%d\n", maj, min, mnt)
		return
	}

	// create context
	m := molly.New()
	m.Config.OutDir = *outdir

	// parameters
	for _, param := range params {
		err := setParameters(m.Config, param)
		if err != nil {
			msg := fmt.Sprintf("Error when processing parameter '%s': %v", param, err)
			help(false, msg, 20)
		}
	}

	// include standrad rules if not exdcluded and we can find them
	if loadStandardRules {
		stdrules := getStandardRules()
		if stdrules == "" {
			help(false, "Could not find standard rules. You may need to add -p config.standardrules=false", 20)
		}
		rfiles = append(rfiles, stdrules)
	}

	// input sanity check
	ifiles := flag.Args()
	if len(rfiles) == 0 && len(rtexts) == 0 {
		help(false, "No rules were given", 20)
	}

	if len(ifiles) == 0 {
		help(false, "No input was given", 20)
	}

	for _, filename := range ifiles {
		if strings.HasPrefix(filename, "-") {
			help(false, "Options must come first", 20)
		}
	}

	if err := util.NewEmptyDir(m.Config.OutDir); err != nil {
		util.RegisterFatalf("Failed to create output directory: %v", err)
	}

	// create callbacks
	listmatch, err := opListParse(matchops)
	if err != nil {
		log.Fatalf("match op error: %s", err)
	}
	m.Config.OnMatchRule = func(i *types.FileData, match *types.Match) {
		id := match.Rule.ID
		if cmd, found := listmatch[id]; found {
			output, err := opExecute(m, cmd, i)
			fmt.Printf("RULE %s on %s: %s\n", id, i.Filename, output)
			if err != nil {
				err = fmt.Errorf("on match %s: %v", id, err)
				i.Errors = append(i.Errors, err)
			}
		}
	}

	listtag, err := opListParse(tagops)
	if err != nil {
		log.Fatalf("tag op error: %s", err)
	}
	m.Config.OnMatchTag = func(i *types.FileData, tag string) {
		if cmd, found := listtag[tag]; found {
			output, err := opExecute(m, cmd, i)
			fmt.Printf("TAG %s on %s: %s\n", tag, i.Filename, output)
			if err != nil {
				err = fmt.Errorf("on tag %s: %v", tag, err)
				i.Errors = append(i.Errors, err)
			}
		}
	}

	for k, v := range listmatch {
		fmt.Println("MATCH", k, v)
	}
	for k, v := range listtag {
		fmt.Println("TAG", k, v)
	}

	//  scan rules
	if err := molly.LoadRules(m, rfiles...); err != nil {
		log.Fatalf("ERROR while parsing rule file: %s", err)
	}

	// add inline rules
	for _, ruletext := range rtexts {
		if err := molly.LoadRulesFromText(m, ruletext); err != nil {
			log.Fatalf("ERROR while parsing inline rule: %s", err)
		}
	}

	if len(m.Rules.Top) == 0 {
		help(false, "No rules were loaded", 20)
	}

	// scan input files
	if len(ifiles) == 0 {
		help(false, "No input files", 20)
	}

	err = molly.ScanFiles(m, ifiles...)
	if err != nil {
		fmt.Println("SCAN while parsing file: ", err)
	}

	// show results
	report := molly.ExtractReport(m)
	dumpResult(m, report, m.Config.Verbose)

	var errors []error

	// generate scan file
	if err := writeScanFiles(m, report); err != nil {
		errors = append(errors, err)
	}

	// generate summary file
	if err := writeSummaryFile(m, report, m.Config.OutDir); err != nil {
		errors = append(errors, err)
	}

	// generate match file
	if err := writeMatchFile(m, report, m.Config.OutDir); err != nil {
		errors = append(errors, err)
	}

	// generate rule file
	if err := writeRuleFile(m, m.Config.OutDir); err != nil {
		errors = append(errors, err)
	}

	// calculate some stats
	totalMatches, totalFiles, totalErrors, totalWarns := 0, 0, len(errors), len(util.Warnings())
	for _, f := range report.Files {
		if len(f.Matches) > 0 {
			totalMatches += len(f.Matches)
			totalFiles++
		}
		totalErrors += len(f.Errors)
		totalWarns += len(f.Warnings)
	}

	fmt.Printf("Scanned %d files, %d of which matched %d rules...\n",
		len(m.Files), len(report.Files), totalMatches)
	fmt.Printf("%d errors, %d warnings\n", totalErrors, totalWarns)
	if totalErrors > 0 {
		os.Exit(1)
	}
}
