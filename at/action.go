package at

import "fmt"

type ActionFunction func(e *Env, args ...string) error

var ActionRegister = map[string]ActionFunction{
	"echo": echoFunction,
}

func echoFunction(e *Env, args ...string) error {
	fmt.Println(e, args)
	return nil
}
