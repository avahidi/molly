package operators

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/avahidi/molly/types"
)

func epoch2time(e *types.Env, epoch int64) (string, error) {
	t := time.Unix(epoch, 0).Local()
	return t.String(), nil
}
func sprintfFunction(e *types.Env, format string, args ...interface{}) (string, error) {
	return fmt.Sprintf(format, args...), nil
}

func printfFunction(e *types.Env, format string, args ...interface{}) (string, error) {
	fmt.Printf(format, args...)
	return "", nil
}

func systemFunction(e *types.Env, format string, args ...interface{}) (string, error) {
	if !e.HasPermission(types.Execute) {
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
	return string(out), fmt.Errorf("system(%s ... ) failed: %v (%s)", cmd[0], err, string(out))
}

func hasFunction(e *types.Env, typ string, val string) (bool, error) {
	switch typ {
	case "match":
		for _, m := range e.Current.Matches {
			if !m.Walk(func(m *types.Match) bool {
				return m.Rule.ID != val
			}) {
				return true, nil
			}
		}
	default:
		return false, fmt.Errorf("Unknown has-property: %s", typ)
	}

	return false, nil
}

func init() {
	Register("printf", printfFunction)
	Register("sprintf", sprintfFunction)
	Register("system", systemFunction)
	Register("epoch2time", epoch2time)
	Register("has", hasFunction)
}
