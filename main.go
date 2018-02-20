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
var outbase = flag.String("outdir", "outs", "output directory")
var logbase = flag.String("logdir", "logs", "log output directory")
var verbose = flag.Bool("v", false, "be verbose")
var showhelp = flag.Bool("h", false, "help information")

var rfiles, tagops MultiFlag
var penable, pdisable MultiFlag

func init() {
	flag.Var(&rfiles, "R", "rule files")
	flag.Var(&tagops, "tagop", "tag operation")
	flag.Var(&penable, "enable", "allowed permissions")
	flag.Var(&pdisable, "disable", "removed permissions")
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
	}
	ifiles := flag.Args()

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
	molly := lib.New(*outbase, *logbase)
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

	// execute tags
	for _, tagop := range tagops {
		if err := tagOperation(report, tagop); err != nil {
			report.Errors = append(report.Errors, fmt.Errorf("Tag operation error: %v", err))
		}
	}

	// show results
	dumpResult(report, *verbose)

	// generate report file
	if err := writeReportFile(*logbase, report); err != nil {
		fmt.Printf("ERROR when creating JSON report: %s\n", err)
	}

	fmt.Printf("Scanned %d files, found %d matches...\n", n, len(report.MatchTree))
	fmt.Printf("%d errors, %d warnings\n", len(report.Errors), len(util.Warnings()))
	if len(report.Errors) > 0 {
		os.Exit(1)
	}
}
