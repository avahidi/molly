package main

import (
	"flag"
	"fmt"
	"log"
	"os"
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
		parametersHelp()
	}
	os.Exit(exitcode)
}

// load rules from built-in, text and file
func loadRules(m *types.Molly, loadBuitin bool, fromText []string, fromFile []string) error {
	// 1. load builtin rules if enabled
	if loadBuitin {
		builtinFiles, builtinData := molly.LoadBuiltinRules()
		for i, file := range builtinFiles {
			source := fmt.Sprintf("<builtin>/%s", file)
			if err := molly.LoadRulesFromText(m, source, builtinData[i]); err != nil {
				return fmt.Errorf("ERROR while parsing built-in rule: %s", err)
			}
		}
	}

	// 2. load inline rules from command-line
	for i, ruletext := range fromText {
		filename := fmt.Sprintf("<inline>/%d", i)
		if err := molly.LoadRulesFromText(m, filename, ruletext); err != nil {
			return fmt.Errorf("ERROR while parsing inline rule: %s", err)
		}
	}
	// 3. load rules from file
	if err := molly.LoadRules(m, fromFile...); err != nil {
		return fmt.Errorf("ERROR while parsing rule file: %s", err)
	}
	return nil
}

func installCallbacks(m *types.Molly, onMatch, onTag []string) error {
	listmatch, err := opListParse(matchops)
	if err != nil {
		return err
	}
	listtag, err := opListParse(tagops)
	if err != nil {
		return err

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
	return nil
}

func showResults(m *types.Molly) int {
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
	return totalErrors
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

	if err := util.NewEmptyDir(m.Config.OutDir); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	// set params
	for _, param := range params {
		err := setParameters(m.Config, param)
		if err != nil {
			msg := fmt.Sprintf("Error when processing parameter '%s': %v", param, err)
			help(false, msg, 20)
		}
	}

	// input sanity check. This must come after creating context and setting parameters
	ifiles := flag.Args()
	if len(ifiles) == 0 {
		help(false, "No input files", 20)
	}
	for _, filename := range ifiles {
		if strings.HasPrefix(filename, "-") {
			help(false, "Options must come first", 20)
		}
	}

	if len(rfiles) == 0 && len(rtexts) == 0 && !loadBuiltinRules {
		help(false, "No rules were given", 20)
	}

	// Create and install callbacks when a rule or tag match is found
	err := installCallbacks(m, matchops, tagops)
	if err != nil {
		log.Fatalf("ERROR when creating callbacks: %v", err)
	}

	// Load rules
	err = loadRules(m, loadBuiltinRules, rtexts, rfiles)
	if err != nil {
		log.Fatal(err)
	}
	if len(m.Rules.Top) == 0 {
		help(false, "No rules were loaded", 20)
	}

	// scan input files
	err = molly.ScanFiles(m, ifiles...)
	if err != nil {
		fmt.Println("SCAN while parsing file: ", err)
	}

	// and show results
	totalErrors := showResults(m)
	if totalErrors > 0 {
		os.Exit(1)
	}
}
