package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"bitbucket.org/vahidi/molly/lib/types"
)

func tagExecute(filename, tag, op string) error {
	info, err := os.Stat(filename)
	if err != nil {
		return err
	}

	// replace {name} and {size}
	sizestr := fmt.Sprintf("%d", info.Size())
	op = strings.Replace(op, "{name}", filename, -1)
	op = strings.Replace(op, "{size}", sizestr, -1)

	// execute command
	cmds := strings.Split(op, " ")
	out, err := exec.Command(cmds[0], cmds[1:]...).CombinedOutput()

	fmt.Printf("\t%s => %s", filename, string(out))
	return err
}
func tagOperation(so *types.MatchReport, tagop string) error {
	n := strings.Index(tagop, ":")
	if n == -1 || n+2 > len(tagop) {
		return fmt.Errorf("invalid tag operation: '%s'", tagop)
	}

	tag := strings.Trim(tagop[:n], " \t")
	op := strings.Trim(tagop[n+1:], " \t")
	fmt.Printf("Tag operation '%s' => '%s'\n", tag, op)

	if files, valid := so.Tagged[tag]; valid {
		for _, filename := range files {
			if err := tagExecute(filename, tag, op); err != nil {
				return err
			}
		}
	}
	return nil
}
