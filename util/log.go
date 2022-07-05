package util

import (
	"fmt"
	"log"
)

var warnings []string

// RegisterFatalf regiters a fatal error
func RegisterFatalf(format string, v ...interface{}) {
	log.Fatalf(format, v...)
}

// RegisterFatal regiters a fatal error
func RegisterFatal(v interface{}) {
	log.Fatal(v)
}

// RegisterWarningf regiters a warning that will be printed and recorded
func RegisterWarningf(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	warnings = append(warnings, msg)
	log.Println(msg)
}

// Warnings returns all warnings seen by the library
func Warnings() []string { return warnings }
