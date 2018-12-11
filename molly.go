package lib

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"

	_ "bitbucket.org/vahidi/molly/actions" // import default actions
	"bitbucket.org/vahidi/molly/types"
	"bitbucket.org/vahidi/molly/util"
)

// Molly represents the context of a molly program
type Molly struct {
	OutDir string
	Rules  *types.RuleSet

	OnMatchRule func(file *types.FileData, match *types.Match)
	OnMatchTag  func(file *types.FileData, tag string)

	MaxDepth int
	Files    map[string]*types.FileData

	// FilesByHash is mainly need to ignore duplicate files
	FilesByHash map[string]*types.FileData
}

// New creates a new molly context
func New(outdir string, maxDepth int) *types.Molly {
	if outdir == "" {
		outdir, _ = ioutil.TempDir("", "molly-out")
	}
	if err := newEmptyDir(outdir); err != nil {
		util.RegisterFatalf("Failed to create output directory: %v", err)
	}
	return types.NewMolly(outdir, maxDepth)
}

func (m *Molly) New(parent *types.FileData, name string, isdir, islog bool) (string, *types.FileData) {
	name = util.SanitizeFilename(name)
	var newname string
	var newdata *types.FileData

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
		newdata = types.NewFileData(newname, parent)
		m.Files[newname] = newdata
		parent.Children = append(parent.Children, newdata)
	}

	return newname, newdata
}

func (m *Molly) CreateFile(parent *types.FileData, name string, islog bool) (*os.File, *types.FileData, error) {
	newfile, newdata := m.New(parent, name, false, islog)
	f, err := util.SafeCreateFile(newfile)
	return f, newdata, err
}

func (m *Molly) CreateDir(parent *types.FileData, name string) (string, *types.FileData, error) {
	newdir, newdata := m.New(parent, name, true, false)
	return newdir, newdata, util.SafeMkdir(newdir)
}
