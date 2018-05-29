package util

import (
	"os"
	"path/filepath"
)

// fileEntry represents a single file
type fileEntry struct {
	Filename string
	Parent   string
}

// FileQueue is a queue where files can be added for later extraction
// it discards files already added and traverses directories
type FileQueue struct {
	current        *fileEntry
	seen           map[string]*fileEntry
	in             []*fileEntry
	followSymlinks bool
}

// NewFileQueue creates an empty FileQueue given a output directory
func NewFileQueue(followSymlinks bool) *FileQueue {
	return &FileQueue{
		seen:           make(map[string]*fileEntry),
		followSymlinks: followSymlinks,
	}

}

func (i *FileQueue) pushOne(path string, parent string) {
	if _, seen := i.seen[path]; seen {
		return
	}

	fe := &fileEntry{Filename: path, Parent: parent}
	i.seen[path] = fe
	i.in = append(i.in, fe)
}

// decides if we should ignore a file
func (i FileQueue) acceptFile(path string, info os.FileInfo) bool {
	if !i.followSymlinks && (info.Mode()&os.ModeSymlink) != 0 {
		RegisterWarningf("Ignoring symlink '%s'", path)
		return false
	}
	return true
}

// Push puts a file in§to the queue
func (i *FileQueue) Push(paths ...string) {
	pname := ""
	if i.current != nil {
		pname = i.current.Filename
	}
	for _, path := range paths {
		i.pushOne(path, pname)
	}
}

// Pop takes a file from the queue
func (i *FileQueue) Pop() (path string, size int64, parent string) {
	for {
		// pop one from the queue
		n := len(i.in)
		if n == 0 {
			return "", 0, ""
		}
		i.current = i.in[n-1]
		i.in = i.in[:n-1]

		path = i.current.Filename
		parent = i.current.Parent
		fi, err := os.Stat(path)
		if err != nil {
			// let someone else take care of the error
			return
		}

		if !i.acceptFile(path, fi) {
			continue
		}

		size = fi.Size()
		mode := fi.Mode()
		if mode.IsDir() {
			dir, err := os.Open(path)
			if err != nil {
				// let someone else take care of the error
				return
			}
			defer dir.Close()

			files, err := dir.Readdir(0)
			if err != nil {
				// let someone else take care of the error
				return i.current.Filename, size, i.current.Parent
			}
			for _, file := range files {
				if i.acceptFile(file.Name(), file) {
					i.pushOne(filepath.Join(path, file.Name()), i.current.Parent)
				}
			}
		} else if mode.IsRegular() {
			return
		} else {
			RegisterWarningf("Ignoring file '%s' (bad mode %s)", path, mode)
		}
	}
}
