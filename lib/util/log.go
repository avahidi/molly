package util

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

var warnings []string

// RegisterFatalf regiters a fatal error
func RegisterFatalf(format string, v ...interface{}) {
	log.Fatalf(format, v)
}

// RegisterFatal regiters a fatal error
func RegisterFatal(v ...interface{}) {
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

var logdir = "logs"

func name(suggestedname string, register bool) string {
	// filename := util.SanitizeFilename(suggestedname, nil)
	filename := suggestedname

	return filepath.Join(logdir, filename)
}

// SetLogBase sets the base directory for all logs
func SetLogBase(path string) {
	logdir = path
}

// CreateLog creates a new log or repoert file
func CreateLog(suggestedname string) (*os.File, error) {
	filename := name(suggestedname, false)

	// make sure the path leading to it exist
	dir, _ := filepath.Split(filename)
	os.MkdirAll(dir, 0700)

	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_EXCL|os.O_TRUNC, 0600)
	return file, err
}
