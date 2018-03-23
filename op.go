package main

import (
	"bitbucket.org/vahidi/molly/lib/types"
	"fmt"
	"os/exec"
	"strings"
)

func opListParse(opstrs []string) (map[string]string, error) {
	m := make(map[string]string)
	for _, opstr := range opstrs {
		n := strings.Index(opstr, ":")
		if n == -1 || n+2 > len(opstr) {
			return nil, fmt.Errorf("invalid operation format: '%s'", opstr)
		}
		k, v := strings.Trim(opstr[:n], " \t"), strings.Trim(opstr[n+1:], " \t")
		m[k] = v
	}
	return m, nil
}

func opExecute(cmdline string, i *types.Input) (string, error) {
	// create list of command and arguments,
	cmds := strings.Split(cmdline, " ")

	// replace {name} and {size}
	sizestr := fmt.Sprintf("%d", i.Filesize)
	for k, s := range cmds {
		s = strings.Replace(s, "{name}", i.Filename, -1)
		s = strings.Replace(s, "{size}", sizestr, -1)
		cmds[k] = s
	}
	// execute command
	out, err := exec.Command(cmds[0], cmds[1:]...).CombinedOutput()
	return string(out), err
}
