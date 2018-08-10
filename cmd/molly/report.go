package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"bitbucket.org/vahidi/molly"
	"bitbucket.org/vahidi/molly/report"
	"bitbucket.org/vahidi/molly/types"
	"bitbucket.org/vahidi/molly/util"
)

// inputToReportEntry converts an input structure to a more readable format
func inputToReportEntry(file *types.Input) map[string]interface{} {
	ret := make(map[string]interface{})

	ret["filename"] = file.Filename
	ret["filesize"] = file.Filesize
	ret["depth"] = file.Depth
	ret["time"] = file.Time.Format(time.RFC3339)

	if len(file.Matches) > 0 {
		ret["matches"] = report.ExtractFlatMatches(file)
	}

	if file.Parent != nil {
		ret["parent"] = file.Parent.Filename
	}

	if tags := report.ExtractTags(file); len(tags) > 0 {
		ret["tags"] = tags
	}

	errstrs := make([]string, len(file.Errors))
	for i, e := range file.Errors {
		errstrs[i] = e.Error()
	}
	ret["errors"] = errstrs

	if len(file.Logs) > 0 {
		ret["logs"] = file.Logs
	}

	return ret
}

func createConfigurationReport(molly *types.Molly) map[string]interface{} {
	f2 := make(map[string]interface{})
	f2["command"] = os.Args
	f2["time"] = time.Now()
	f2["outdir"] = molly.OutDir
	f2["rulecount"] = len(molly.Rules.Flat)
	maj, min, mnt := lib.Version()
	f2["version"] = fmt.Sprintf("%d.%d.%d", maj, min, mnt)
	dir, err := os.Getwd()
	if err != nil {
		f2["dir"] = dir
	}
	return f2
}

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

func writeSummaryFile(molly *types.Molly, r *types.Report, base string) error {
	w, err := os.Create(filepath.Join(base, "summary.json"))
	if err != nil {
		return err
	}
	defer w.Close()

	f1 := make(map[string]interface{})
	f1["configuration"] = createConfigurationReport(molly)
	f1["file-hierarchy"] = report.ExtractFileHierarchy(molly)
	f1["logs"] = report.ExtractLogHierarchy(r)
	f1["tags"] = report.ExtractTagHierarchy(r)

	matches := make(map[string]int)
	for _, file := range molly.Processed {
		matches[file.Filename] = 0 // update below!
	}
	for _, file := range r.Files {
		fm := report.ExtractFlatMatches(file)
		matches[file.Filename] = len(fm)
	}
	f1["matches"] = matches

	// report errors for each file
	errs := make(map[string][]string)
	for _, input := range r.Files {
		if len(input.Errors) != 0 {
			estrs := make([]string, len(input.Errors))
			for i, e := range input.Errors {
				estrs[i] = e.Error()
			}
			errs[input.Filename] = estrs
		}
	}
	f1["errors"] = errs

	bs, err := json.MarshalIndent(f1, "", "\t")
	if err != nil {
		return err
	}
	w.Write(bs)

	return nil
}

func writeScanFiles(molly *types.Molly, r *types.Report) error {
	for _, file := range molly.Processed {
		rep := inputToReportEntry(file)

		bs, err := json.MarshalIndent(rep, "", "\t")
		if err != nil {
			return err
		}

		w, err := os.Create(fmt.Sprintf("%s_molly.json", file.FilenameOut))
		if err != nil {
			return err
		}
		w.Write(bs)
		w.Close() // manual Close() or we will have too many files open
	}
	return nil
}

func writeMatchFile(molly *types.Molly, r *types.Report, base string) error {
	w, err := os.Create(filepath.Join(base, "match.json"))
	if err != nil {
		return err
	}
	defer w.Close()

	f1 := make(map[string]interface{})
	f1["configuration"] = createConfigurationReport(molly)

	results := make(map[string]interface{})
	for _, file := range r.Files {
		results[file.Filename] = report.ExtractMatchNames(file, true)
	}
	f1["matches"] = results

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
		if file.Empty() {
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

	firstError := false
	for _, file := range r.Files {
		if len(file.Errors) == 0 {
			continue
		}
		if firstError {
			firstError = false
			fmt.Println("ERRORS:")
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
