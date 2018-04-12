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

	followSymlinks bool
}

// NewFileQueue creates an empty FileQueue given a output directory
func NewFileQueue(maxDepth int, followSymlinks bool) *FileQueue {
	return &FileQueue{
		Map:            make(map[string]*FileEntry),
		MaxDepth:       maxDepth,
		followSymlinks: followSymlinks,
	}

}

func (i *FileQueue) pushOne(path string, parent *FileEntry) {
	if _, seen := i.Map[path]; seen {
		return
	}
	depth := 0
	if parent != nil {
		depth = parent.Depth + 1
	}
	if i.MaxDepth == 0 || depth < i.MaxDepth {
		fe := &FileEntry{Filename: path, Parent: parent, Depth: depth}
		i.Map[path] = fe
		i.In = append(i.In, fe)
	} else {
		RegisterWarningf("ignoring file '%s' at max depth %d\n", path, depth)
	}
}

// decides if we should ignore a file
func (i FileQueue) acceptFile(path string, info os.FileInfo) bool {
	if !i.followSymlinks && (info.Mode()&os.ModeSymlink) != 0 {
		RegisterWarningf("Ignoring symlink '%s'\n", path)
		return false
	}
	return true
}

// Push puts a file in§to the queue
func (i *FileQueue) Push(paths ...string) {
	for _, path := range paths {
		i.pushOne(path, i.Current)
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

		if !i.acceptFile(path, fi) {
			continue
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
				if i.acceptFile(file.Name(), file) {
					i.pushOne(filepath.Join(path, file.Name()), i.Current.Parent)
				}
			}
		} else if mode.IsRegular() {
			i.Current.Size = fi.Size()
			i.Out = append(i.Out, i.Current)
			return i.Current
		} else {
			RegisterWarningf("Ignoring file '%s' (bad mode %s)\n", path, mode)
		}
	}
}
