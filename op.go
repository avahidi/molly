package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"bitbucket.org/vahidi/molly/lib/report"
	"bitbucket.org/vahidi/molly/lib/types"
)

func parseOp(opstr string) (string, string) {
	n := strings.Index(opstr, ":")
	if n == -1 || n+2 > len(opstr) {
		return "", ""
	}
	return strings.Trim(opstr[:n], " \t"), strings.Trim(opstr[n+1:], " \t")
}

func opExecute(filename, id, cmdline string) error {
	info, err := os.Stat(filename)
	if err != nil {
		return err
	}

	// create list of command and arguments,
	cmds := strings.Split(cmdline, " ")

	// replace {name} and {size}
	sizestr := fmt.Sprintf("%d", info.Size())
	for i, s := range cmds {
		s = strings.Replace(s, "{name}", filename, -1)
		s = strings.Replace(s, "{size}", sizestr, -1)
		cmds[i] = s
	}

	// execute command
	out, err := exec.Command(cmds[0], cmds[1:]...).CombinedOutput()
	if err == nil {
		fmt.Printf("\tOKAY:%s %s => %s\n", id, filename, string(out))
	} else {
		fmt.Printf("\tFAIL:%s %s => %s %v\n", id, filename, string(out), err)
	}
	return err
}

func executeAllTagOps(mr *types.Report, tagops []string) []error {
	var errors []error
	tagmap := report.ExtractTagHierarchy(mr)

	for _, tagop := range tagops {
		tag, cmd := parseOp(tagop)
		if len(cmd) <= 0 {
			errors = append(errors, fmt.Errorf("invalid tag operation: '%s'", tagop))
			continue
		}

		fmt.Printf("Tag operation '%s' => '%s':\n", tag, cmd)
		files := tagmap[tag]
		for _, file := range files {
			if err := opExecute(file, tag, cmd); err != nil {
				errors = append(errors, err)
			}
		}
	}
	return errors
}

func executeAllMatchOps(mr *types.Report, matchops []string) []error {
	var errors []error

	for _, matchop := range matchops {
		rulename, cmd := parseOp(matchop)
		if len(cmd) <= 0 {
			errors = append(errors, fmt.Errorf("invalid rule operation: '%s'", matchop))
			continue
		}

		fmt.Printf("Match operation '%s' => '%s':\n", rulename, cmd)
		for _, matchfile := range mr.Files {
			matchfile.Walk(func(match *types.Match) {
				if match.Rule.ID == rulename {
					if err := opExecute(matchfile.Filename, rulename, cmd); err != nil {
						errors = append(errors, err)
					}
				}
			})
		}
	}
	return errors
}
