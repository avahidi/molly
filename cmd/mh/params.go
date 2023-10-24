package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/avahidi/molly/types"
)

var loadBuiltinRules = false

var parameters = map[string]any{
	"config.builtin":  false,
	"config.maxdepth": 12,
	"config.verbose":  false,
	"perm.create":     true,
	"perm.execute":    false,
}

func parametersHelp() {
	fmt.Printf("Available parameters are:\n")
	for k, v := range parameters {
		fmt.Printf("\t%-25s (%v)\n", k, v)
	}
}

func setParameterInt(c *types.Configuration, name string, i int) error {
	switch name {
	case "config.maxdepth":
		c.MaxDepth = i
	}
	return nil
}

func setParameterBool(c *types.Configuration, name string, b bool) error {
	switch name {
	case "config.verbose":
		c.Verbose = b
	case "config.builtin":
		loadBuiltinRules = b
	case "perm.create":
		c.SetPermission(types.Create, b)
	case "perm.execute":
		c.SetPermission(types.Execute, b)
	default:
	}
	return nil
}

func setParameters(c *types.Configuration, p string) error {
	kv := strings.SplitN(p, "=", 2)

	if len(kv) != 2 {
		return fmt.Errorf("Bad parameter, expected key=value got '%s'", p)
	}

	key, value := kv[0], kv[1]
	oldval, ok := parameters[key]
	if !ok {
		return fmt.Errorf("Unknown parameter '%s'", kv[0])
	}

	switch v := oldval.(type) {
	case int:
		newvalue, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		parameters[key] = int(newvalue)
		return setParameterInt(c, key, int(newvalue))
	case bool:
		newvalue, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		parameters[key] = newvalue
		return setParameterBool(c, key, newvalue)
	default:
		return fmt.Errorf("Internal error: paramater %s has the type %T\n", key, v)
	}
}
