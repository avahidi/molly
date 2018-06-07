package types

import (
	"fmt"
	"os"
	"path"

	"bitbucket.org/vahidi/molly/lib/util"
)

// Molly represents the context of a molly program
type Molly struct {
	OutDir string
	Rules  *RuleSet

	OnMatchRule func(file *Input, match *Match)
	OnMatchTag  func(file *Input, tag string)

	MaxDepth  int
	Processed map[string]*Input
}

// NewMolly creates a new Molly context
func NewMolly(outdir string, maxDepth int) *Molly {
	return &Molly{
		OutDir:    outdir,
		Rules:     NewRuleSet(),
		MaxDepth:  maxDepth,
		Processed: make(map[string]*Input),
	}
}

func (m *Molly) CreateName(parent *Input, name string, isdir, islog bool) string {
	name = util.SanitizeFilename(name)
	var newname string

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
		if util.GetPathType(newname) == util.NoFile {
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
		parent.Children = append(parent.Children, newname)
	}
	return newname
}

func (m *Molly) CreateFile(parent *Input, name string, islog bool) (*os.File, error) {
	newfile := m.CreateName(parent, name, false, islog)
	return util.SafeCreateFile(newfile)
}

func (m *Molly) CreateDir(parent *Input, name string) (string, error) {
	newdir := m.CreateName(parent, name, true, false)
	return newdir, util.SafeMkdir(newdir)
}
