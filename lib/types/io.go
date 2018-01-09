package types

import "os"

// FileSystem handles is a virtual file hierarchy
// where names have to be checked before and modified creation
type FileSystem interface {
	Name(suggestedname string, register bool) (string, error)
	Create(suggestedname string) (*os.File, error)
	Mkdir(suggestedpath string) error
}
