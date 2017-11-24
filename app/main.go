package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"bitbucket.org/vahidi/molly"
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
var outbase = flag.String("O", ".", "output directory")
var allowSystem = flag.Bool("system", false, "allow system calls from rules")
var rfiles MultiFlag

func init() {
	flag.Var(&rfiles, "R", "rule files")
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
	ifiles := flag.Args()

	// create database and scan rules
	db := molly.New(*outbase)
	if err := db.ScanRules(rfiles); err != nil {
		fmt.Println("ERROR while parsing rule file: ", err)
	}
	if len(db.Rules) == 0 {
		help("No rules were loaded", 20)
	}

	// scan input files
	if len(ifiles) == 0 {
		help("No input files", 20)
	}

	ss, err := db.ScanFiles(ifiles, nil)
	if err != nil {
		fmt.Println("SCAN while parsing file: ", err)
	}

	// show results
	fmt.Println("SCAN RESULTS:")
	for _, match := range ss.Results {
		fmt.Printf("\t%s %s\n", match.Rule, match.Filename)
		for k, v := range match.Vars {
			fmt.Printf("\t\t%s : %T = %v\n", k, v, v)
		}
	}

	if len(ss.Errors) > 0 {
		fmt.Println("ERRORS:")
		for _, e := range ss.Errors {
			fmt.Printf("\t%s\n", e)
		}
	}
}
