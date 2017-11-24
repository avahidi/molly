package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"bitbucket.org/vahidi/molly/lib"
	"bitbucket.org/vahidi/molly/lib/types"
	"bitbucket.org/vahidi/molly/lib/util/logging"
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
var allowSystem = flag.Bool("system", false, "allow system calls from rules")
var verbose = flag.Bool("v", false, "be verbose")

var shortHelp = flag.Bool("h", false, "help information")
var longHelp = flag.Bool("H", false, "extended help information")

var rfiles, tagops MultiFlag

func init() {
	flag.Var(&rfiles, "R", "rule files")
	flag.Var(&tagops, "tagop", "tag operation")
}

func help(errmsg string, extended bool, exitcode int) {
	if errmsg != "" {
		fmt.Printf("%s\n", errmsg)
	}
	flag.Usage()
	fmt.Printf("  files\n\tinput files to be scanned\n")

	if extended {
		fmt.Printf(" The following functions are defined:\n")
		strs := types.FunctionHelp()
		for _, str := range strs {
			fmt.Printf("\t%s\n", str)
		}
	}

	os.Exit(exitcode)
}

func main() {
	// 	parse arguments
	flag.Parse()
	ifiles := flag.Args()

	// set logoutput
	logging.SetBase(*logbase)

	if *longHelp {
		help("", true, 0)
	} else if *shortHelp {
		help("", false, 0)
	}

	// create database
	db := lib.New(*outbase, nil)
	db.Globals.Set("AllowSystemAction", *allowSystem)

	//  scan rules
	if err := db.ScanRules(rfiles); err != nil {
		fmt.Println("ERROR while parsing rule file: ", err)
		os.Exit(20)
	}
	if len(db.RuleSet.Top) == 0 {
		help("No rules were loaded", false, 20)
	}

	// scan input files
	if len(ifiles) == 0 {
		help("No input files", false, 20)
	}

	ins, outs, err := db.ScanFiles(ifiles)
	if err != nil {
		fmt.Println("SCAN while parsing file: ", err)
	}

	fmt.Printf("Scanned %d files, found %d matches...\n",
		len(ins.Files()), len(outs.MatchTree))

	// execute tags
	for _, tagop := range tagops {
		if err := tagOperation(outs, tagop); err != nil {
			outs.Errors = append(outs.Errors, fmt.Errorf("Tag operation error: %v", err))
		}
	}

	// show results
	dumpResult(outs, *verbose)

	// generate report file
	if err := writeReportFile(outs); err != nil {
		fmt.Printf("ERROR when creating JSON report: %s\n", err)
	}

	fmt.Printf("%d errors, %d warnings\n", len(outs.Errors), len(logging.Warnings()))
	if len(outs.Errors) > 0 {
		os.Exit(1)
	}
}
