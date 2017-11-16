package molly

import (
	"fmt"
	"os"

	"bitbucket.org/vahidi/molly/at"
	"bitbucket.org/vahidi/molly/at/parser"
	"bitbucket.org/vahidi/molly/util"
)

// MatchCallback is the type of function called when a match is found
type MatchCallback func(m *MatchEntry)

// Molly is the main structure and contains the rule database and configuration
type Molly struct {
	Rules    []*at.Rule
	inputset *InputSet
}

func (m Molly) matchCallback(me *MatchEntry) {
	fmt.Printf("Found '%s' in file '%s'\n", me.Rule, me.Filename)
}

func New(outputDir string) *Molly {
	m := &Molly{
		inputset: newInputSet(outputDir),
	}
	if err := os.MkdirAll(outputDir, 0700); err != nil {
		panic("Could not create output directory")
	}
	return m
}

func (m *Molly) ScanRules(files []string) error {
	var fl util.FileList
	if err := fl.Walk(files...); err != nil {
		return err
	}

	rs, err := parser.ParseRules(fl)
	if err != nil {
		return err
	}

	for _, c := range rs.Rules {
		m.Rules = append(m.Rules, c)
	}
	return nil
}

func (m *Molly) ScanFiles(files []string, mcs []MatchCallback) (*OutputSet, error) {
	for _, file := range files {
		m.inputset.Push(file)
	}
	outputset := &OutputSet{}

	env := at.NewEnvironment(m.inputset.CreateNew, m.inputset.Push)
	for {
		filename := m.inputset.Pop()
		if filename == "" {
			break
		}
		info, err := os.Stat(filename)
		if err != nil {
			outputset.Errors = append(outputset.Errors, err)
			continue
		}
		// fmt.Printf("Scaning file %s...\n", filename)

		f, err := os.Open(filename)
		if err != nil {
			outputset.Errors = append(outputset.Errors, err)
			continue
		}
		defer f.Close()

		env.SetFile(f, uint64(info.Size()), filename)
		for _, r := range m.Rules {
			env.Reset()
			evalOne(filename, outputset, r, env)
		}
	}

	// callbacks
	for _, me := range outputset.Results {
		m.matchCallback(me)
		for _, mc := range mcs {
			mc(me)
		}
	}

	return outputset, nil
}

// evalOne evaluates one rule against one file,
// if the rule has children they will also be evaluated
func evalOne(filename string, s *OutputSet, r *at.Rule, e *at.Env) {
	// 1. evaluate the rule
	match, err := r.Eval(e)
	if err != nil {
		s.Errors = append(s.Errors, err)
		return
	}
	if !match {
		return
	}

	// 2. call all action functions
	for _, a := range r.Actions {
		_, err := a.Action.Eval(e)
		if err != nil {
			s.Errors = append(s.Errors, err)
		}
	}

	// 3. record the match
	m := &MatchEntry{Filename: filename, Rule: r.Id, Vars: e.Extract()}
	s.Results = append(s.Results, m)

	// 4. call children
	for _, cr := range r.Children {
		e.PushScope()
		evalOne(filename, s, cr, e)
		e.PopScope()
	}
}
