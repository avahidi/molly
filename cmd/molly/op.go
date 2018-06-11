package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"bitbucket.org/vahidi/molly/types"
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

func opLookupVariable(m *types.Molly, i *types.Input, cmd, data string) (interface{}, error) {
	// simple variables coming from the input file?
	if o, found := i.Get(cmd); found {
		return o, nil
	}

	// actions or complex variables
	switch cmd {
	case "newfile":
		return m.CreateName(i, data, false, false), nil
	case "newlog":
		return m.CreateName(i, data, false, true), nil
	}

	return nil, fmt.Errorf("unknown variable: '%s'", cmd)
}

// opReplaceVariables replaces variables such as {filename} with values
func opReplaceVariables(m *types.Molly, s string, i *types.Input) (string, error) {
	str, cmd, data := bytes.Buffer{}, bytes.Buffer{}, bytes.Buffer{}

	// simple state machine to extract cmd & data from "...{cmd:data}...""
	state := 0
	for _, r := range s {
		switch state {
		case 0:
			if r == '{' {
				cmd.Reset()
				data.Reset()
				state = 1
			} else {
				str.WriteRune(r)
			}
		case 1:
			if r == ':' {
				state = 2
			} else if r == '}' {
				state = 4
			} else {
				cmd.WriteRune(r)
			}
		case 2:
			if r == '}' {
				state = 4
			} else {
				data.WriteRune(r)
			}
		}

		// found a variable, look it up
		if state == 4 {
			o, err := opLookupVariable(m, i, cmd.String(), data.String())
			if err != nil {
				return "", err
			}
			str.WriteString(fmt.Sprintf("%v", o))
			state = 0
		}
	}

	if state != 0 {
		return "", fmt.Errorf("'{' not terminated")
	}
	return str.String(), nil
}

func opExecute(m *types.Molly, cmdline string, i *types.Input) (string, error) {
	var err error
	// create list of command and arguments,
	cmds := strings.Split(cmdline, " ")
	for k, s := range cmds {
		cmds[k], err = opReplaceVariables(m, s, i)
		if err != nil {
			return "", err
		}
	}

	// execute command
	out, err := exec.Command(cmds[0], cmds[1:]...).CombinedOutput()
	return string(out), err
}
