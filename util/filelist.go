package util

import (
	"os"
	"path/filepath"
)

// FileList is a list of files, including files in folders
type FileList struct {
	FollowSymlinks bool
	In             []string
}

// Push puts a file in§to the queue
func (fl *FileList) Push(paths ...string) {
	for _, path := range paths {
		fl.In = append(fl.In, path)
	}
}

// Pop takes a file from the queue
func (fl *FileList) Pop() (string, os.FileInfo, error) {
	var fi os.FileInfo
	for {
		// pop one from the queue
		n := len(fl.In)
		if n == 0 {
			return "", fi, nil
		}

		filename := fl.In[n-1]
		fl.In = fl.In[:n-1]

		fi, err := os.Lstat(filename)
		if err != nil {
			// let someone else take care of the error
			return filename, fi, err
		}

		if !fl.FollowSymlinks && (fi.Mode()&os.ModeSymlink) != 0 {
			RegisterWarningf("Ignoring symlink '%s'", filename)
			continue
		}

		mode := fi.Mode()
		if mode.IsDir() {
			dir, err := os.Open(filename)
			if err != nil {
				return filename, fi, err
			}
			defer dir.Close()

			files, err := dir.Readdir(0)
			if err != nil {
				return filename, fi, err
			}
			for _, file := range files {
				fl.Push(filepath.Join(filename, file.Name()))
			}
		} else if mode.IsRegular() {
			return filename, fi, nil
		} else {
			RegisterWarningf("Ignoring file '%s' (bad mode %s)", filename, mode)
		}
	}
}
