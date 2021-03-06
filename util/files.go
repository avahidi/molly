package util

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Extensions returns all extensions of a file, e.g. a.tar.gz -> ["gz","tar"]
func Extensions(filename string) []string {
	var ret []string
	for {
		ext := filepath.Ext(filename)
		if ext == "" || len(ext) > 8 { /* assume short extensions */
			return ret
		}
		filename = filename[:len(filename)-len(ext)]
		ret = append(ret, ext[1:])
	}
}

// SanitizeFilename performs file name sanitization
func SanitizeFilename(filename string) string {
	const badchars = "()\\;<>?* \000"
	filename = strings.Replace(filename, "../", "", -1)
	filename = strings.Replace(filename, "..", "", -1)

	var buf bytes.Buffer
	for _, r := range filename {
		if r == 0 {
			break
		}
		if strings.IndexRune(badchars, r) != -1 || !strconv.IsPrint(r) {
			buf.WriteRune('_')
		} else {
			buf.WriteRune(r)
		}
	}
	return buf.String()
}

// NewEmptyDir accepts a new or empty dir and if new creates it
func NewEmptyDir(dirname string) error {
	typ := GetPathType(dirname)
	switch typ {
	case File:
		return fmt.Errorf("'%s' is a file", dirname)
	case NonEmptyDir:
		return fmt.Errorf("'%s' exists and is not empty", dirname)
	case Error:
		return fmt.Errorf("'%s' could not be checked", dirname)
	default:
		return Mkdir(dirname)
	}
}

// Mkdir creates directories in a path
func Mkdir(path string) error {
	return os.MkdirAll(path, 0755)
}

// CreateFile creates a file, adds missing directories
func CreateFile(filename string) (*os.File, error) {
	// make sure its path also exists
	dir, _ := filepath.Split(filename)
	if err := Mkdir(dir); err != nil {
		return nil, err
	}

	return os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_EXCL|os.O_TRUNC, 0644)
}

// PathType is type of a path such as /tmp, i.e. wether its a file or a dir etc
type PathType int

const (
	Error PathType = iota
	NoFile
	File
	EmptyDir
	NonEmptyDir
)

// GetPathType returns type of a path
func GetPathType(path string) PathType {
	info, err := os.Stat(path)
	if err != nil {
		return NoFile
	}
	if !info.IsDir() {
		return File
	}

	if file, err := os.Open(path); err == nil {
		defer file.Close()
		files, _ := file.Readdir(1)
		if len(files) == 0 {
			return EmptyDir
		}
		return NonEmptyDir
	}
	return Error
}
