package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"bitbucket.org/vahidi/molly/lib/report"
	"bitbucket.org/vahidi/molly/lib/types"
)

func tagExecute(filename, tag, op string) error {
	info, err := os.Stat(filename)
	if err != nil {
		return err
	}

	// create list of command and arguments
	cmds := strings.Split(op, " ")

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
		fmt.Printf("\tOKAY %s => %s\n", filename, string(out))
	} else {
		fmt.Printf("\tFAIL %s => %s %v\n", filename, string(out), err)
	}
	return err
}

func executeAllTagOps(mr *types.Report, tagop []string) []error {
	var errors []error
	tagmap := report.ExtractTagHierarchy(mr)

	for _, tagop := range tagops {

		// extract the tag
		n := strings.Index(tagop, ":")
		if n == -1 || n+2 > len(tagop) {
			errors = append(errors, fmt.Errorf("invalid tag operation: '%s'", tagop))
			continue
		}
		tag := strings.Trim(tagop[:n], " \t")
		op := strings.Trim(tagop[n+1:], " \t")
		fmt.Printf("Tag operation '%s' => '%s':\n", tag, op)

		//
		files := tagmap[tag]
		for _, file := range files {
			if err := tagExecute(file, tag, op); err != nil {
				errors = append(errors, err)
			}
		}
	}
	return errors
}
