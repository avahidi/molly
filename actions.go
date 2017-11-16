package molly

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"bitbucket.org/vahidi/molly/at"
	"bitbucket.org/vahidi/molly/util"
)

func printfFunction(e *at.Env, format string, args ...interface{}) (interface{}, error) {
	fmt.Printf(format, args...)
	return false, nil
}

func extractFunction(e *at.Env, filename string, offset int32, size int32, queue bool) (interface{}, error) {
	fmt.Printf("extracting %d bytes into %s...\n", size, filename)

	// input file
	if err := e.Seek(int64(offset)); err != nil {
		return nil, err
	}

	// output file:
	newname := e.NewFile(filename)
	w, err := os.Create(newname)
	if err != nil {
		return nil, err
	}
	defer w.Close()

	util.Copy(e, w, int(size))
	e.AddFile(newname)
	return nil, nil
}

func systemFunction(e *at.Env, command string, args ...interface{}) (interface{}, error) {
	// to build parameter list, convert interfaces to strings
	argv := make([]string, len(args))
	for i, arg := range args {
		switch n := arg.(type) {
		case string:
			argv[i] = n
		case fmt.Stringer:
			argv[i] = n.String()
		case fmt.GoStringer:
			argv[i] = n.GoString()
		default:
			argv[i] = fmt.Sprintf("%v", arg)
			// return false, fmt.Errorf("Could not convert %v to string", arg)
		}
	}

	fmt.Println(command, strings.Join(argv, " "))

	// now execute it:
	out, err := exec.Command(command, argv...).Output()
	if err == nil {
		fmt.Printf("%s\n", string(out))
		return string(out), nil
	}
	fmt.Printf("FAILED: %v\n", command, err)
	return nil, err
}

func newfileFunction(e *at.Env, filename string) (interface{}, error) {
	return e.NewFile(filename), nil
}
func addfileFunction(e *at.Env, filename string) (interface{}, error) {
	e.AddFile(filename)
	return filename, nil
}

func init() {
	at.RegisterActionFunction("printf", printfFunction)
	at.RegisterActionFunction("extract", extractFunction)
	at.RegisterActionFunction("system", systemFunction)
	at.RegisterActionFunction("newfile", newfileFunction)
	at.RegisterActionFunction("addfile", addfileFunction)
}
