package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"bitbucket.org/vahidi/molly/lib/types"
)

func contains(array []string, val string) bool {
	for _, i := range array {
		if i == val {
			return true
		}
	}
	return false
}

func tagExecute(filename, tag, op string) error {
	info, err := os.Stat(filename)
	if err != nil {
		return err
	}
	sizestr := fmt.Sprintf("%d", info.Size())

	op = strings.Replace(op, "{name}", filename, -1)
	op = strings.Replace(op, "{size}", sizestr, -1)
	cmds := strings.Split(op, " ")

	fmt.Printf("TAG-OP: %s -> %s\n", tag, cmds)
	out, err := exec.Command(cmds[0], cmds[1:]...).CombinedOutput()
	fmt.Printf("\t%s\n", string(out))

	return err
}
func tagOperation(so *types.MatchReport, tagop string) error {
	n := strings.Index(tagop, ":")
	if n == -1 || n+2 > len(tagop) {
		return fmt.Errorf("invalid tag operation: '%s'", tagop)
	}

	tag := strings.Trim(tagop[:n], " \t")
	op := strings.Trim(tagop[n+1:], " \t")

	for filename, tags := range so.TaggedFiles {
		if contains(tags, tag) {
			if err := tagExecute(filename, tag, op); err != nil {
				return err
			}
		}
	}
	return nil
}
