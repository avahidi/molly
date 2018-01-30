package actions

import (
	"fmt"
	"os/exec"
	"strings"

	"bitbucket.org/vahidi/molly/lib/types"
	"bitbucket.org/vahidi/molly/lib/util"
)

func sprintfFunction(e *types.Env, format string, args ...interface{}) (string, error) {
	return fmt.Sprintf(format, args...), nil
}

func printfFunction(e *types.Env, format string, args ...interface{}) (string, error) {
	fmt.Printf(format, args...)
	return "", nil
}

func systemFunction(e *types.Env, format string, args ...interface{}) (interface{}, error) {
	if !util.PermissionGet(util.Execute) {
		return "", fmt.Errorf("system actions are not allowed, action ignored (%s)", e)
	}

	cmd := strings.Split(fmt.Sprintf(format, args...), " ")
	fmt.Println("EXECUTING", cmd, "...") // DEBUG

	// now execute it:
	out, err := exec.Command(cmd[0], cmd[1:]...).CombinedOutput()
	if err == nil {
		return string(out), nil
	}
	fmt.Println(string(out), err)
	return nil, fmt.Errorf("system(%s ... ) failed: %v (%s)", cmd[0], err, string(out))
}

func init() {
	types.FunctionRegister("printf", printfFunction)
	types.FunctionRegister("sprintf", sprintfFunction)
	types.FunctionRegister("system", systemFunction)
}