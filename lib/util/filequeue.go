package util

import (
	"os"
	"path/filepath"
)

// FileQueue is a queue where files can be added for later extraction
// it discards files already added and traverses directories
type FileQueue struct {
	processed []string
	toprocess []string
	recent    []string
	seen      map[string]bool
}

// NewFileQueue creates an empty FileQueue given a output directory
func NewFileQueue() *FileQueue {
	return &FileQueue{seen: make(map[string]bool)}
}

// Count retruns number of files processed (POPed) in this queue
func (i FileQueue) Count() int {
	return len(i.processed)
}

// Push puts a file into the queue
func (i *FileQueue) Push(paths ...string) {
	for _, path := range paths {
		if _, seen := i.seen[path]; !seen {
			i.seen[path] = true
			i.toprocess = append(i.toprocess, path)
			i.recent = append(i.recent, path)
		}
	}
}

// Pop takes a file from the queue
func (i *FileQueue) Pop() string {
	for {
		// pop one from the queue
		n := len(i.toprocess)
		if n == 0 {
			return ""
		}
		path := i.toprocess[n-1]
		i.toprocess = i.toprocess[:n-1]

		fi, err := os.Stat(path)
		if err != nil {
			return path // let someone else take care of the error
		}

		mode := fi.Mode()
		if mode.IsDir() {
			dir, err := os.Open(path)
			if err != nil {
				return path // let someone else take care of the error
			}
			defer dir.Close()

			files, err := dir.Readdir(0)
			if err != nil {
				return path // let someone else take care of the error
			}
			for _, file := range files {
				i.Push(filepath.Join(path, file.Name()))
			}
		} else if mode.IsRegular() {
			i.processed = append(i.processed, path)
			return path
		} else {
			RegisterWarningf("ignoring unknown file type '%s'\n", path)
		}
	}
}

// RecentlyAdded returns list of files that were added
// since the last call to this function
func (i *FileQueue) RecentlyAdded() []string {
	tmp := i.recent
	i.recent = nil
	return tmp
}
