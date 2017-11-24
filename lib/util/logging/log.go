package logging

import (
	"fmt"
	"log"
)

var warnings []string

// Fatalf regiters a fatal error
func Fatalf(format string, v ...interface{}) {
	log.Fatalf(format, v)
}

// Fatal regiters a fatal error
func Fatal(v ...interface{}) {
	log.Fatal(v)
}

// Warningf regiters a warning that will be printed and recorded
func Warningf(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	warnings = append(warnings, msg)
	log.Println(msg)
}

// Warnings returns all warnings seen by the library
func Warnings() []string { return warnings }
