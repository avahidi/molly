package types

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"bitbucket.org/vahidi/molly/lib/util"
)

// Molly represents the context of a molly program
type Molly struct {
	ExtractDir string
	ReportDir  string

	Rules *RuleSet
	Files *util.FileQueue

	OnMatchRule func(file *Input, match *Match)
	OnMatchTag  func(file *Input, tag string)

	pathCnt   int
	pathCache map[string]string
}

// NewMolly creates a new Molly context
func NewMolly(extratDir, reportDir string, maxDepth int) *Molly {
	return &Molly{
		ExtractDir: extratDir,
		ReportDir:  reportDir,
		Rules:      NewRuleSet(),
		Files:      util.NewFileQueue(maxDepth, false),
		pathCache:  make(map[string]string),
	}
}

func (m *Molly) CreateName(parent *Input, name string, islog bool) string {
	var parentName string
	if parent != nil {
		parentName = parent.Filename
	}

	home := m.ExtractDir
	if islog {
		home = m.ReportDir
	}

	base, found := m.pathCache[parentName]
	if !found {
		subdir := parentName
		if strings.HasPrefix(subdir, home) {
			subdir = subdir[len(home):]
		} else {
			_, subdir = filepath.Split(subdir)
		}
		subdir = fmt.Sprintf("%s_%04d", subdir, m.pathCnt)
		m.pathCnt++
		base = filepath.Join(home, subdir)
		m.pathCache[parentName] = base
	}
	name = util.SanitizeFilename(name, nil)
	newname := filepath.Join(base, name)
	for {
		if _, err := os.Stat(newname); err != nil {
			break
		}
		newname = filepath.Join(base, fmt.Sprintf("%04d_%s", m.pathCnt, name))
		m.pathCnt++
	}

	// register it!
	if islog {
		parent.Logs = append(parent.Logs, newname)
	} else {
		m.Files.Push(newname)
	}

	return newname
}

func (m *Molly) CreateFile(parent *Input, name string, islog bool) (*os.File, error) {
	newfile := m.CreateName(parent, name, islog)
	return util.SafeCreateFile(newfile)
}

func (m *Molly) CreateDir(parent *Input, name string) (string, error) {
	newdir := m.CreateName(parent, name, false)
	return newdir, util.SafeMkdir(newdir)
}
