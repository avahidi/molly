package types

import (
	"fmt"
	"os"
	"path"

	"github.com/avahidi/molly/util"
)

// Permission defines a molly permission such as the ability to create new files
type Permission uint32

const (
	Create Permission = 1 << iota
	Execute
)

// Configuration contains all runtime parameters used by molly
type Configuration struct {
	OutDir      string
	MaxDepth    int
	Verbose     bool
	Permissions Permission
	OnMatchRule func(file *FileData, match *Match)
	OnMatchTag  func(file *FileData, tag string)
}

// HasPermission checks if a permission is set
func (c Configuration) HasPermission(p Permission) bool {
	// return (c.Permissions & (1 << p)) != 0
	return (c.Permissions & p) != 0
}

// SetPermission sets or clears a Permission
func (c *Configuration) SetPermission(p Permission, val bool) {
	if val {
		c.Permissions |= p
	} else {
		c.Permissions &= ^(p)
	}
}

// Molly represents the context of a molly program
type Molly struct {
	Config *Configuration
	Rules  *RuleSet

	Files map[string]*FileData
	// FilesByHash is mainly need to ignore duplicate files
	FilesByHash map[string]*FileData
}

// NewMolly creates a new Molly context
func NewMolly() *Molly {
	config := &Configuration{
		OutDir:      "output",
		MaxDepth:    12,
		Permissions: Create,
	}

	return &Molly{
		Config:      config,
		Rules:       NewRuleSet(),
		Files:       make(map[string]*FileData),
		FilesByHash: make(map[string]*FileData),
	}
}

func (m *Molly) New(parent *FileData, name string, isdir, islog bool) (*FileData, error) {
	if !m.Config.HasPermission(Create) {
		return nil, fmt.Errorf("Not allowed to create files/dirs")
	}

	name = util.SanitizeFilename(name)
	var newname string
	var newdata *FileData

	// get a unique name we can use
	for i := 0; ; i++ {
		if islog {
			if i == 0 {
				newname = fmt.Sprintf("%s_molly_%s", parent.FilenameOut, name)
			} else {
				newname = fmt.Sprintf("%s_molly_%04d_%s", parent.FilenameOut, i, name)
			}
		} else {
			if i == 0 {
				newname = fmt.Sprintf("%s_/%s", parent.FilenameOut, name)
			} else {
				newname = fmt.Sprintf("%s_/%04d_%s", parent.FilenameOut, i, name)
			}
		}
		if _, found := m.Files[newname]; !found && util.GetPathType(newname) == util.NoFile {
			break
		}
	}

	// make sure parent folders exist
	if isdir {
		util.Mkdir(newname)
	} else {
		base, _ := path.Split(newname)
		util.Mkdir(base)
	}

	// remember it:
	newdata = NewFileData(newname, parent)
	if islog {
		parent.Logs = append(parent.Logs, newname)
	} else {
		m.Files[newname] = newdata
		parent.Children = append(parent.Children, newdata)
	}

	return newdata, nil
}

func (m *Molly) CreateFile(parent *FileData, name string, islog bool) (file *os.File, data *FileData, err error) {
	data, err = m.New(parent, name, false, islog)
	if err == nil {
		file, err = util.CreateFile(data.Filename)
	}
	return
}

func (m *Molly) CreateDir(parent *FileData, name string) (data *FileData, err error) {
	data, err = m.New(parent, name, true, false)
	if err == nil {
		err = util.Mkdir(data.Filename)
	}
	return
}
