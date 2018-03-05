package util

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

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
	return os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_EXCL|os.O_TRUNC, 0644)
}

// FileSystem is a structure for tracking and creating new files
// in a controlled manner and with some historic which is later
// used to create scan reports and such
type FileSystem struct {
	cnt   int
	bases map[string]string
	base  string
	queue *FileQueue
}

// NewFileSystem creates a new filesystem given a base directory
// and an optional queue to store new files in
func NewFileSystem(base string, queue *FileQueue) *FileSystem {
	return &FileSystem{
		cnt:   0,
		base:  base,
		queue: queue,
		bases: make(map[string]string),
	}
}

// record records creation of a file/dir
func (fs *FileSystem) record(realname, parent string) {
	if fs.queue != nil {
		fs.queue.Push(realname)
	}
}

// newFile suggest a new filename
func (fs *FileSystem) newFile(name, parent string) (string, error) {
	base, found := fs.bases[parent]
	if !found {
		subdir := parent
		if strings.HasPrefix(subdir, fs.base) {
			subdir = subdir[len(fs.base):] + "_"
		} else {
			_, subdir = filepath.Split(subdir)
			subdir = fmt.Sprintf("%04d_%s_", fs.cnt, subdir)
			fs.cnt++
		}
		base = filepath.Join(fs.base, subdir)
		fs.bases[parent] = base
	}
	name = SanitizeFilename(name, nil)
	newname := filepath.Join(base, name)
	if _, err := os.Stat(newname); err == nil {
		newname = filepath.Join(base, fmt.Sprintf("%04d_%s", fs.cnt, name))
		fs.cnt++
	}
	return newname, nil
}

// Name suggest a new name for a file based
// on a name suggestion + the currently scanned file
func (fs *FileSystem) Name(name, parent string, addtopath bool) (string, error) {
	newname, err := fs.newFile(name, parent)
	if err != nil {
		return "", err
	}
	if addtopath {
		fs.record(newname, parent)
	}
	return newname, nil
}

// Mkdir creates a new directory based on a path suggestion
func (fs *FileSystem) Mkdir(path, parent string) (string, error) {
	newpath, err := fs.Name(path, parent, false)
	if err != nil {
		return "", err
	}
	err = SafeMkdir(newpath)
	if err != nil {
		return "", err
	}
	fs.record(newpath, parent)
	return newpath, nil
}

// Create a new file based on Name()
func (fs *FileSystem) Create(name, parent string) (*os.File, error) {
	if !PermissionGet(CreateFile) {
		return nil, fmt.Errorf("Not allowed to create files (mkdir")
	}
	newname, err := fs.Name(name, parent, false)
	if err != nil {
		return nil, err
	}

	// make sure the path leading to it exist
	dir, _ := filepath.Split(newname)
	if err := SafeMkdir(dir); err != nil {
		return nil, err
	}

	// open the file and record this event
	file, err := SafeCreateFile(newname)
	if err == nil {
		fs.record(newname, parent)
	}
	return file, err
}
