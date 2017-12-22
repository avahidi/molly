package types

import "os"

// FileQueue is a queue where files can be added for later extraction
// it discards files already added
type FileQueue interface {
	Push(files ...string)
	Pop() string
}

// FileSystem handles is a virtual file hierarchy
// where names have to be checked before and modified creation
type FileSystem interface {
	Name(suggestedname string, register bool) string
	Create(suggestedname string) (*os.File, error)
	Mkdir(suggestedpath string) error
}
