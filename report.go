package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"bitbucket.org/vahidi/molly/lib/types"
	"bitbucket.org/vahidi/molly/lib/util"
)

func writeReportFile(so *types.MatchReport) error {
	w, err := util.CreateLog("report.json")
	if err != nil {
		return err
	}
	defer w.Close()

	f1 := make(map[string]interface{})
	f2 := make(map[string]interface{})
	f2["command"] = os.Args
	f2["time"] = time.Now()
	dir, err := os.Getwd()
	if err != nil {
		f1["dir"] = dir
	}

	f1["configuration"] = f2
	f1["results"] = so.MatchTree
	f1["tags"] = so.TaggedFiles

	// create new copy of hierarchy without nils:
	hrc := make(map[string][]string)
	for k, v := range so.FileHierarchy {
		if v != nil {
			hrc[k] = v
		}
	}
	f1["hierarchy"] = hrc

	bs, err := json.Marshal(f1)
	if err != nil {
		return err
	}
	w.Write(bs)

	return nil
}

func padlevel(level int) string {
	var padding = "\t\t\t\t\t\t\t\t\t\t\t\t\t\t"
	return padding[:level]

}
func printmatch(match *types.MatchEntry, verbose bool, level int) {
	pad := padlevel(level)
	fmt.Printf("\t%s * %s %s\n", pad, match.Rule, match.Filename)
	if verbose {
		for k, v := range match.Vars {
			fmt.Printf("\t%s   . %s : %T = %v\n", pad, k, v, v)
		}
	}

	for _, ch := range match.Children {
		printmatch(ch, verbose, level+1)
	}
}

func dumpResult(so *types.MatchReport, verbose bool) {
	fmt.Println("SCAN RESULTS:")
	for _, match := range so.MatchTree {
		printmatch(match, verbose, 0)
	}

	if verbose {
		fmt.Println("\nFile hierarchy:")
		for p, fs := range so.FileHierarchy {
			if fs != nil {
				fmt.Println("\t", p)
				for _, f := range fs {
					fmt.Println("\t\t=>", f)
				}
			}
		}
	}

	es, ws := so.Errors, util.Warnings()
	if len(es) > 0 {
		fmt.Println("ERRORS:")
		for _, e := range es {
			fmt.Printf("\t%s\n", e)
		}
	}
	if len(ws) > 0 {
		fmt.Println("Warnings:")
		for _, w := range ws {
			fmt.Printf("\t%s\n", w)
		}
	}
}
