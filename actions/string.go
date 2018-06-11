package actions

import (
	"strconv"
	"strings"

	"bitbucket.org/vahidi/molly/types"
)

func strlenFunction(e *types.Env, str string) (int, error) {
	return len(str), nil
}

func strtoaFunction(e *types.Env, str string) (int64, error) {
	return strconv.ParseInt(str, 10, 64)
}

func strstrFunction(e *types.Env, str1, str2 string) (bool, error) {
	return strings.Index(str1, str2) != -1, nil
}

func strcasestrFunction(e *types.Env, str1, str2 string) (bool, error) {
	return strings.Index(strings.ToUpper(str1), strings.ToUpper(str2)) != -1, nil
}

func stricmpFunction(e *types.Env, str1, str2 string) (bool, error) {
	return strings.ToUpper(str1) == strings.ToUpper(str2), nil
}

func strupperFunction(e *types.Env, str string) (string, error) {
	return strings.ToUpper(str), nil
}

func strlowerFunction(e *types.Env, str string) (string, error) {
	return strings.ToLower(str), nil
}

func strprefixFunction(e *types.Env, str1, str2 string) (bool, error) {
	return strings.HasPrefix(str1, str2), nil
}

func strsuffixFunction(e *types.Env, str1, str2 string) (bool, error) {
	return strings.HasSuffix(str1, str2), nil
}

func init() {
	ActionRegister("strlen", strlenFunction)
	ActionRegister("strtol", strtoaFunction)

	ActionRegister("strstr", strstrFunction)
	ActionRegister("strcasestr", strcasestrFunction)
	ActionRegister("stricmp", stricmpFunction)

	ActionRegister("strupper", strupperFunction)
	ActionRegister("strlower", strlowerFunction)
	ActionRegister("strprefix", strprefixFunction)
	ActionRegister("strsuffix", strsuffixFunction)
}