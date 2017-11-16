package util

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// ReadUntil reads from stream until it sees byte "until" or
// gathers "maxsize" bytes (ignored if zero or less)
func ReadUntil(r io.Reader, until byte, maxsize int) ([]byte, error) {
	// this is probably horrible inefficient
	const READ = 32
	buf := make([]byte, READ)
	data := make([]byte, 0)
	for {
		n, err := r.Read(buf)
		if err != nil {
			return nil, err
		}
		for i := 0; i < n; i++ {
			if buf[i] == until || (maxsize > 0 && len(data) > maxsize) {
				return data, nil
			}
			data = append(data, buf[i])
		}
	}
}

func Copy(r io.Reader, w io.Writer, bytes int) error {
	cap := 32 * 1024
	buf := make([]byte, cap)

	for bytes > 0 {
		var m int
		var err error

		if bytes < cap {
			m, err = r.Read(buf[:bytes])
		} else {
			m, err = r.Read(buf)
		}
		if err != nil {
			return err
		}
		if m == 0 {
			return fmt.Errorf("Copy: file pre-amture end")
		}
		w.Write(buf[:m])
		bytes -= m
	}
	return nil
}

// FileList represents a set of files
type FileList []string

// Walk scans the list of paths and adds all files to the list
func (fl *FileList) Walk(paths ...string) error {
	for _, path := range paths {
		if err := filepath.Walk(path, fl.walker); err != nil {
			return err
		}
	}
	return nil
}

func (fl *FileList) walker(filename string, info os.FileInfo, eprev error) error {
	if info != nil && !info.IsDir() {
		*fl = append(*fl, filename)
	}
	return eprev
}
