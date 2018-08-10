package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"bitbucket.org/vahidi/molly/actions"

	"bitbucket.org/vahidi/molly"
	"bitbucket.org/vahidi/molly/types"
	"bitbucket.org/vahidi/molly/util"
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
var outdir = flag.String("outdir", "output", "output directory")
var verbose = flag.Bool("v", false, "be verbose")
var showVersion = flag.Bool("V", false, "show version number")
var showhelp = flag.Bool("h", false, "help information")
var showhelpExt = flag.Bool("hh", false, "extended help information")
var maxDepth = flag.Int("max-depth", 0, "max scan depth")

var rfiles, rtexts, tagops, matchops MultiFlag
var penable, pdisable MultiFlag

func init() {
	flag.Var(&rfiles, "R", "rule files")
	flag.Var(&rtexts, "r", "inline rule")
	flag.Var(&tagops, "on-tag", "tag match operations")
	flag.Var(&matchops, "on-rule", "rule match operations")
	flag.Var(&penable, "enable", "allow permission")
	flag.Var(&pdisable, "disable", "remove permission")
}

func help(extended bool, errmsg string, exitcode int) {
	if errmsg != "" {
		fmt.Printf("%s\n", errmsg)
	}

	maj, min, mnt := lib.Version()
	fmt.Printf("This is Molly version %d.%d.%d\n", maj, min, mnt)

	flag.Usage()

	fmt.Printf("  files\n\tinput files to be scanned\n")

	if extended {
		actions.ActionHelp()
	}
	os.Exit(exitcode)
}

func main() {
	// 	parse arguments
	flag.Parse()

	if *showhelpExt || *showhelp {
		help(*showhelpExt, "", 0)
	}

	if *showVersion {
		maj, min, mnt := lib.Version()
		fmt.Printf("%d.%d.%d\n", maj, min, mnt)
		return
	}
	// update permissions
	for i, list := range []MultiFlag{penable, pdisable} {
		set := (i == 0)
		for _, name := range list {
			p, found := util.PermissionNames[name]
			if !found {
				fmt.Printf("Unknown permission '%s'. ", name)
				util.PermissionHelp()
				help(false, "", 20)
			}
			util.PermissionSet(p, set)
		}
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

	// create context
	molly := lib.New(*outdir, *maxDepth)

	// create callbacks
	listmatch, err := opListParse(matchops)
	if err != nil {
		log.Fatalf("match op error: %s", err)
	}
	molly.OnMatchRule = func(i *types.Input, match *types.Match) {
		id := match.Rule.ID
		if cmd, found := listmatch[id]; found {
			output, err := opExecute(molly, cmd, i)
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
	molly.OnMatchTag = func(i *types.Input, tag string) {
		if cmd, found := listtag[tag]; found {
			output, err := opExecute(molly, cmd, i)
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
	if err := lib.LoadRules(molly, rfiles...); err != nil {
		log.Fatalf("ERROR while parsing rule file: %s", err)
	}

	// add inline rules
	for _, ruletext := range rtexts {
		if err := lib.LoadRulesFromText(molly, ruletext); err != nil {
			log.Fatalf("ERROR while parsing inline rule: %s", err)
		}
	}

	if len(molly.Rules.Top) == 0 {
		help(false, "No rules were loaded", 20)
	}

	// scan input files
	if len(ifiles) == 0 {
		help(false, "No input files", 20)
	}

	report, err := lib.ScanFiles(molly, ifiles)
	if err != nil {
		fmt.Println("SCAN while parsing file: ", err)
	}

	// show results
	dumpResult(molly, report, *verbose)

	var errors []error

	// generate scan file
	if err := writeScanFiles(molly, report); err != nil {
		errors = append(errors, err)
	}

	// generate summary file
	if err := writeSummaryFile(molly, report, *outdir); err != nil {
		errors = append(errors, err)
	}

	// generate match file
	if err := writeMatchFile(molly, report, *outdir); err != nil {
		errors = append(errors, err)
	}

	// generate rule file
	if err := writeRuleFile(molly, *outdir); err != nil {
		errors = append(errors, err)
	}

	// calculate some stats
	totalMatches, totalFiles, totalErrors := 0, 0, len(errors)
	for _, f := range report.Files {
		if len(f.Matches) > 0 {
			totalMatches += len(f.Matches)
			totalFiles++
		}
		totalErrors += len(f.Errors)
	}

	fmt.Printf("Scanned %d files, %d of which matched %d rules...\n",
		len(molly.Processed), len(report.Files), totalMatches)
	fmt.Printf("%d errors, %d warnings\n", totalErrors, len(util.Warnings()))
	if totalErrors > 0 {
		os.Exit(1)
	}
}
