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
func (fl *FileList) Pop() (string, int64, error) {
	for {
		// pop one from the queue
		n := len(fl.In)
		if n == 0 {
			return "", 0, nil
		}

		filename := fl.In[n-1]
		fl.In = fl.In[:n-1]

		fi, err := os.Stat(filename)
		if err != nil {
			// let someone else take care of the error
			return "", 0, err
		}

		if !fl.FollowSymlinks && (fi.Mode()&os.ModeSymlink) != 0 {
			RegisterWarningf("Ignoring symlink '%s'", filename)
			continue
		}

		mode := fi.Mode()
		if mode.IsDir() {
			dir, err := os.Open(filename)
			if err != nil {
				return filename, 0, err
			}
			defer dir.Close()

			files, err := dir.Readdir(0)
			if err != nil {
				return filename, 0, err
			}
			for _, file := range files {
				fl.Push(filepath.Join(filename, file.Name()))
			}
		} else if mode.IsRegular() {
			return filename, fi.Size(), nil
		} else {
			RegisterWarningf("Ignoring file '%s' (bad mode %s)", filename, mode)
		}
	}
}
