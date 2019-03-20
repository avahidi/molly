package main

import (
	"fmt"
	"strconv"
	"strings"

	"bitbucket.org/vahidi/molly/types"
)

var loadStandardRules = true

// setBooleanCheck is a helper function for setting a boolean from a string
func setBooleanCheck(b *bool, val string) error {
	bb, err := strconv.ParseBool(val)
	if err != nil {
		return err
	}
	*b = bb
	return nil
}

// setIntCheck is a helper function for setting an integer from a string
func setIntCheck(i *int, val string) error {
	ii, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return err
	}
	*i = int(ii)
	return nil
}

func setConfig(c *types.Configuration, key, val string) error {
	switch key {
	case "maxdepth":
		return setIntCheck(&c.MaxDepth, val)
	case "verbose":
		return setBooleanCheck(&c.Verbose, val)
	case "standardrules":
		return setBooleanCheck(&loadStandardRules, val)
	default:
		return fmt.Errorf("Unknown configuration: '%s'", key)
	}
	return nil
}

func setPermission(c *types.Configuration, perm, val string) error {
	b, err := strconv.ParseBool(val)
	if err != nil {
		return err
	}
	switch perm {
	case "create":
		c.SetPermission(types.Create, b)
	case "execute":
		c.SetPermission(types.Execute, b)
	default:
		return fmt.Errorf("Unknown permission: '%s'", perm)
	}
	return nil
}

func setParameters(c *types.Configuration, p string) error {
	strs := strings.SplitN(p, "=", 2)

	if len(strs) != 2 {
		return fmt.Errorf("Bad parameter, expected key=value got '%s'", p)
	}

	keys := strings.SplitN(strs[0], ".", 2)
	if len(keys) != 2 {
		return fmt.Errorf("Bad configuration key, expected type.name=value")
	}

	typ, key, val := keys[0], keys[1], strs[1]

	switch typ {
	case "config":
		return setConfig(c, key, val)
	case "perm":
		return setPermission(c, key, val)
	default:
		return fmt.Errorf("Unknown parameter class '%s' in '%s'", typ, p)

	}
}
