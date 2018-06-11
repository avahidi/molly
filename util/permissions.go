package util

import "fmt"

// Permission defines a molly permission such as the ability to create new files
type Permission int

const (
	CreateFile Permission = iota
	Execute
)

// permissions is the global variables that holds current permission set
var permissions uint64

// PermissionSet sets or clears a Permission
func PermissionSet(p Permission, val bool) {
	if val {
		permissions |= 1 << (uint64)(p)
	} else {
		permissions &= ^(1 << (uint64)(p))
	}
}

// PermissionGet checks if a permission is set
func PermissionGet(p Permission) bool {
	return (permissions & (1 << (uint64)(p))) != 0
}

// PermissionNames contains name mappings for all available permissions
var PermissionNames = map[string]Permission{
	"create-file": CreateFile,
	"execute":     Execute,
}

// PermissionHelp prints help text for permissions
func PermissionHelp() {
	fmt.Println("Valid permissions are")
	for valid := range PermissionNames {
		fmt.Printf("\t%s\n", valid)
	}
}

func init() {
	// default permissions
	PermissionSet(CreateFile, true)
}
