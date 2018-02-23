package main

import (
	"bitbucket.org/vahidi/molly/lib"
	"bitbucket.org/vahidi/molly/lib/util"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
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
var outbase = flag.String("outdir", "build/extracted", "output directory")
var repbase = flag.String("repdir", "build/reports", "report output directory")
var verbose = flag.Bool("v", false, "be verbose")
var showVersion = flag.Bool("V", false, "show version number")
var showVersionOnly = flag.Bool("VV", false, "show only version number")
var showhelp = flag.Bool("h", false, "help information")

var rfiles, tagops MultiFlag
var penable, pdisable MultiFlag

func init() {
	flag.Var(&rfiles, "R", "rule files")
	flag.Var(&tagops, "tagop", "tag operation")
	flag.Var(&penable, "enable", "allow permission")
	flag.Var(&pdisable, "disable", "remove permission")
}

func help(errmsg string, exitcode int) {
	if errmsg != "" {
		fmt.Printf("%s\n", errmsg)
	}
	flag.Usage()
	fmt.Printf("  files\n\tinput files to be scanned\n")
	os.Exit(exitcode)
}

func main() {
	// 	parse arguments
	flag.Parse()
	if *showhelp {
		help("", 0)
		os.Exit(0)
	}
	ifiles := flag.Args()

	if *showVersionOnly {
		maj, min, mnt := lib.Version()
		fmt.Printf("%d.%d.%d\n", maj, min, mnt)
		return
	} else if *showVersion {
		maj, min, mnt := lib.Version()
		fmt.Printf("This is Molly version %d.%d.%d\n", maj, min, mnt)
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
				help("", 20)
			}
			util.PermissionSet(p, set)
		}
	}

	// create context
	molly := lib.New(*outbase, *repbase)
	//  scan rules
	if err := lib.LoadRules(molly, rfiles...); err != nil {
		log.Fatalf("ERROR while parsing rule file: %s", err)
	}
	if len(molly.Rules.Top) == 0 {
		help("No rules were loaded", 20)
	}

	// scan input files
	if len(ifiles) == 0 {
		help("No input files", 20)
	}

	report, n, err := lib.ScanFiles(molly, ifiles)
	if err != nil {
		fmt.Println("SCAN while parsing file: ", err)
	}

	// show results
	dumpResult(molly, report, *verbose)

	// execute tags
	errors := executeAllTagOps(report, tagops)

	// generate report file
	if err := writeReportFile(molly, report, *repbase); err != nil {
		errors = append(errors, err)
	}

	// generate rule file
	if err := writeRuleFile(molly, *repbase); err != nil {
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

	fmt.Printf("Scanned %d files, %d of which matched %d rules...\n", n, totalFiles, totalMatches)
	fmt.Printf("%d errors, %d warnings\n", totalErrors, len(util.Warnings()))
	if totalErrors > 0 {
		os.Exit(1)
	}
}
