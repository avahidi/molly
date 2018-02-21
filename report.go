package main

import (
	"bitbucket.org/vahidi/molly/lib/report"
	"bitbucket.org/vahidi/molly/lib/types"
	"bitbucket.org/vahidi/molly/lib/util"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func writeRuleFile(molly *types.Molly, base string) error {
	w, err := os.Create(filepath.Join(base, "rules.json"))
	if err != nil {
		return err
	}
	defer w.Close()

	bs, err := json.MarshalIndent(molly.Rules, "", "\t")
	if err != nil {
		return err
	}
	w.Write(bs)

	return nil
}

func writeReportFile(molly *types.Molly, r *types.Report, base string) error {
	w, err := os.Create(filepath.Join(base, "report.json"))
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

	f1["file-hierarchy"] = report.ExtractFileHierarchy(molly)
	f1["logs"] = report.ExtractLogHierarchy(r)
	f1["tags"] = report.ExtractTagHierarchy(r)
	f1["configuration"] = f2
	f1["results"] = report.ExtractFlatReport(r)

	bs, err := json.MarshalIndent(f1, "", "\t")
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

func printmatchShort(match *types.Match, level int) {
	if level == 0 {
		fmt.Print("\t\t =>")
	}
	fmt.Printf(" %s", match.Rule.ID)
	for _, ch := range match.Children {
		printmatchShort(ch, level+1)
	}
	if level == 0 {
		fmt.Println()
	}
}
func printmatchVerbose(match *types.Match, level int) {
	pad := padlevel(level)
	fmt.Printf("\t\t%s * %s\n", pad, match.Rule.ID)

	for k, v := range match.Vars {
		fmt.Printf("\t%s   . %s : %T = %v\n", pad, k, v, v)
	}
	for _, ch := range match.Children {
		printmatchVerbose(ch, level+1)
	}
}

func dumpResult(m *types.Molly, r *types.Report, verbose bool) {
	fmt.Println("SCAN RESULTS:")
	for _, file := range r.Files {
		if len(file.Matches) == 0 {
			continue
		}
		fmt.Printf("\t* File %s (%d errors):\n", file.Filename, len(file.Errors))
		for _, match := range file.Matches {
			if verbose {
				printmatchVerbose(match, 0)
			} else {
				printmatchShort(match, 0)
			}
		}
	}

	if verbose {
		fmt.Println("\nFile hierarchy:")
		files := report.ExtractFileHierarchy(m)
		for p, fs := range files {
			if fs != nil {
				fmt.Println("\t", p)
				for _, f := range fs {
					fmt.Println("\t\t=>", f)
				}
			}
		}
	}

	fmt.Println("ERRORS:")
	for _, file := range r.Files {
		if len(file.Errors) == 0 {
			continue
		}
		fmt.Printf("\t* File %s:\n", file.Filename)
		for i, err := range file.Errors {
			fmt.Printf("\t\t %d: %v\n", i+1, err)
		}
	}

	ws := util.Warnings()
	if len(ws) > 0 {
		fmt.Println("Warnings:")
		for _, w := range ws {
			fmt.Printf("\t%s\n", w)
		}
	}
}
