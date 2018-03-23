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

// replaceVars find and replaces <token> with f()
func replaceVars(s, token string, f func() (string, error)) (string, error) {
	for {
		n := strings.Index(s, token)
		if n == -1 {
			return s, nil
		}
		before, after := s[:n], s[n+len(token):]
		replaced, err := f()
		if err != nil {
			return s, err
		}
		return fmt.Sprintf("%s%s%s", before, replaced, after), nil
	}
}

// replaceVarsWithParam find and replaces <prefix><data><postfix> with f(<data>)
func replaceVarsWithParam(s, prefix, postfix string, f func(data string) (string, error)) (string, error) {
	for {
		n := strings.Index(s, prefix)
		if n == -1 {
			return s, nil
		}
		m := strings.Index(s[n:], postfix)
		if m == -1 {
			return s, fmt.Errorf("'%s' without '%s'", postfix, prefix)
		}

		before, data, after := s[:n], s[n+len(prefix):n+m], s[n+m+len(postfix):]
		replaced, err := f(data)
		if err != nil {
			return s, err
		}
		return fmt.Sprintf("%s%s%s", before, replaced, after), nil
	}
}

// opReplaceVariables replaces variables such as {filename} with values
func opReplaceVariables(m *types.Molly, s string, i *types.Input) (string, error) {
	var err error

	// start with the simple ones
	s = strings.Replace(s, "{filename}", i.Filename, -1)
	s = strings.Replace(s, "{filesize}", fmt.Sprintf("%d", i.Filesize), -1)

	// complex variables widthout parameters
	s, err = replaceVars(s, "{newfile}", func() (string, error) {
		return m.CreateName(i, "", false), nil
	})
	if err != nil {
		return s, err
	}

	s, err = replaceVars(s, "{newlog}", func() (string, error) {
		return m.CreateName(i, "", true), nil
	})
	if err != nil {
		return s, err
	}

	// complex variables that have parameters
	s, err = replaceVarsWithParam(s, "{newfile:", "}", func(name string) (string, error) {
		return m.CreateName(i, name, false), nil
	})
	if err != nil {
		return s, err
	}

	s, err = replaceVarsWithParam(s, "{newlog:", "}", func(name string) (string, error) {
		return m.CreateName(i, name, true), nil
	})
	if err != nil {
		return s, err
	}

	return s, nil
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
