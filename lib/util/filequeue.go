package util

import (
	"os"
	"path/filepath"
)

// FileEntry represents a single file
type FileEntry struct {
	Filename string
	Size     int64
	Depth    int
	Parent   *FileEntry
}

// FileQueue is a queue where files can be added for later extraction
// it discards files already added and traverses directories
type FileQueue struct {
	MaxDepth int
	Current  *FileEntry
	Map      map[string]*FileEntry
	Out      []*FileEntry
	In       []*FileEntry
}

// NewFileQueue creates an empty FileQueue given a output directory
func NewFileQueue(maxDepth int) *FileQueue {
	return &FileQueue{Map: make(map[string]*FileEntry), MaxDepth: maxDepth}

}

func (i *FileQueue) pushOne(path string, parent *FileEntry, incDepth bool) {
	if _, seen := i.Map[path]; seen {
		return
	}
	depth := 0
	if parent != nil {
		depth = parent.Depth
		if incDepth {
			depth++
		}
	}
	if i.MaxDepth == 0 || depth < i.MaxDepth {
		fe := &FileEntry{Filename: path, Parent: parent, Depth: depth}
		i.Map[path] = fe
		i.In = append(i.In, fe)
	} else {
		RegisterWarningf("ignoring file '%s' at max depth %d\n", path, depth)
	}
}

// Push puts a file in§to the queue
func (i *FileQueue) Push(paths ...string) {
	for _, path := range paths {
		i.pushOne(path, i.Current, true)
	}
}

// Pop takes a file from the queue
func (i *FileQueue) Pop() *FileEntry {
	for {
		// pop one from the queue
		n := len(i.In)
		if n == 0 {
			return nil
		}
		i.Current = i.In[n-1]
		i.In = i.In[:n-1]

		path := i.Current.Filename
		fi, err := os.Stat(path)
		if err != nil {
			return i.Current // let someone else take care of the error
		}
		mode := fi.Mode()
		if mode.IsDir() {
			dir, err := os.Open(path)
			if err != nil {
				return i.Current // let someone else take care of the error
			}
			defer dir.Close()

			files, err := dir.Readdir(0)
			if err != nil {
				return i.Current // let someone else take care of the error
			}
			for _, file := range files {
				i.pushOne(filepath.Join(path, file.Name()), i.Current, false)
			}
		} else if mode.IsRegular() {
			i.Current.Size = fi.Size()
			i.Out = append(i.Out, i.Current)
			return i.Current
		} else {
			RegisterWarningf("ignoring unknown file type '%s'\n", path)
		}
	}
}
