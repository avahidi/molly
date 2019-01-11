package types

import (
	"fmt"
	"os"
	"path"

	"bitbucket.org/vahidi/molly/util"
)

// Configuration contains all runtime parameters used by molly
type Configuration struct {
	OutDir      string
	MaxDepth    int
	Verbose     bool
	OnMatchRule func(file *FileData, match *Match)
	OnMatchTag  func(file *FileData, tag string)
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
		OutDir:   "output",
		MaxDepth: 12,
	}

	return &Molly{
		Config:      config,
		Rules:       NewRuleSet(),
		Files:       make(map[string]*FileData),
		FilesByHash: make(map[string]*FileData),
	}
}

func (m *Molly) New(parent *FileData, name string, isdir, islog bool) (string, *FileData) {
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
		util.SafeMkdir(newname)
	} else {
		base, _ := path.Split(newname)
		util.SafeMkdir(base)
	}

	// remember it:
	if islog {
		parent.Logs = append(parent.Logs, newname)
	} else {
		newdata = NewFileData(newname, parent)
		m.Files[newname] = newdata
		parent.Children = append(parent.Children, newdata)
	}

	return newname, newdata
}

func (m *Molly) CreateFile(parent *FileData, name string, islog bool) (*os.File, *FileData, error) {
	newfile, newdata := m.New(parent, name, false, islog)
	f, err := util.SafeCreateFile(newfile)
	return f, newdata, err
}

func (m *Molly) CreateDir(parent *FileData, name string) (string, *FileData, error) {
	newdir, newdata := m.New(parent, name, true, false)
	return newdir, newdata, util.SafeMkdir(newdir)
}
