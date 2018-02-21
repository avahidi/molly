package util

import (
	"os"
	"path/filepath"
)

// MaxFileDepth is the maximum depth of files we will accept
const MaxFileDepth = 42

// FileEntry represents a single file
type FileEntry struct {
	Filename string
	Depth    int
	Parent   *FileEntry
}

// FileQueue is a queue where files can be added for later extraction
// it discards files already added and traverses directories
type FileQueue struct {
	Current *FileEntry
	Map     map[string]*FileEntry
	Out     []*FileEntry
	In      []*FileEntry
}

// NewFileQueue creates an empty FileQueue given a output directory
func NewFileQueue() *FileQueue {
	return &FileQueue{Map: make(map[string]*FileEntry)}
}

// Push puts a file into the queue
func (i *FileQueue) Push(paths ...string) {
	for _, path := range paths {
		if _, seen := i.Map[path]; !seen {
			fe := &FileEntry{
				Filename: path,
				Parent:   i.Current,
			}
			if fe.Parent != nil {
				fe.Depth = 1 + fe.Parent.Depth
			}
			if fe.Depth < MaxFileDepth {
				i.Map[path] = fe
				i.In = append(i.In, fe)
			} else {
				RegisterWarningf("ignoring file '%s' at max depth %d\n", path, fe.Depth)
			}
		}
	}
}

// Pop takes a file from the queue
func (i *FileQueue) Pop() string {
	for {
		// pop one from the queue
		n := len(i.In)
		if n == 0 {
			return ""
		}
		i.Current = i.In[n-1]
		i.In = i.In[:n-1]

		path := i.Current.Filename
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
			i.Out = append(i.Out, i.Current)
			return path
		} else {
			RegisterWarningf("ignoring unknown file type '%s'\n", path)
		}
	}
}
