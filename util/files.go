package util

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
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

// Mkdir creates directories in a path
func Mkdir(path string) error {
	return os.MkdirAll(path, 0755)
}

// SafeMkdir creates directories in a path, fails if CreateFile permission is missing
func SafeMkdir(path string) error {
	if !PermissionGet(CreateFile) {
		return fmt.Errorf("Not allowed to create files (mkdir)")
	}
	return Mkdir(path)
}

// SafeMkdirWithTime is similar to SafeMkdir but also sets dirtime
func SafeMkdirWithTime(filename string, tim *time.Time) error {
	err := SafeMkdir(filename)
	if tim != nil && err != nil {
		err = os.Chtimes(filename, *tim, *tim)
	}
	return err
}

// SafeCreateFile creates a file, fails if CreateFile permission is missing
func SafeCreateFile(filename string) (*os.File, error) {
	if !PermissionGet(CreateFile) {
		return nil, fmt.Errorf("Not allowed to create file '%s'", filename)
	}

	// make sure its path also exists
	dir, _ := filepath.Split(filename)
	if err := Mkdir(dir); err != nil {
		return nil, err
	}

	return os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_EXCL|os.O_TRUNC, 0644)
}

//  creates a file, fails if CreateFile permission is missing
// SafeCreateFileWithTime is similar to SafeMkdir but also sets dirtime
func SafeCreateFileWithTime(filename string, tim *time.Time) (*os.File, error) {
	fil, err := SafeCreateFile(filename)
	if tim != nil && err != nil {
		err = os.Chtimes(filename, *tim, *tim)
	}
	return fil, err
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
