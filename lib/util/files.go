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
func SanitizeFilename(filename string, filter func(rune) bool) string {
	var buf bytes.Buffer
	if filter == nil {
		filter = func(r rune) bool {
			const badchars = "()\\;<>?* \000"
			return strings.IndexRune(badchars, r) != -1 || !strconv.IsPrint(r)
		}
	}

	filename = strings.Replace(filename, "..", "_", -1)
	for _, r := range filename {
		if r == 0 {
			break
		}
		if filter(r) {
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
