package molly

import (
	"bytes"
	"fmt"
	"io"

	_ "bitbucket.org/vahidi/molly/actions" // import default actions
	"bitbucket.org/vahidi/molly/report"
	"bitbucket.org/vahidi/molly/scan"
	"bitbucket.org/vahidi/molly/types"
)

// processMatch will process a match on a rule
func processMatch(c *types.Configuration, i *types.FileData, match *types.Match) {
	if len(match.Children) == 0 {
		if c.OnMatchRule != nil {
			c.OnMatchRule(i, match)
		}
	}
	for _, ch := range match.Children {
		processMatch(c, i, ch)
	}
}

// processMatch will process a tag on a file
func processTags(c *types.Configuration, fr *types.FileData) {
	if c.OnMatchTag != nil {
		tags := report.ExtractTags(fr)
		for _, tag := range tags {
			c.OnMatchTag(fr, tag)
		}
	}
}

func scanInput(m *types.Molly, env *types.Env, reader io.ReadSeeker, data *types.FileData) {

	env.SetInput(reader, data)
	for pass := types.RulePassMin; pass <= types.RulePassMax; pass++ {
		for _, rule := range m.Rules.Top {
			if p, _ := rule.Metadata.Get("pass", int64(types.RulePassMin)); p != int64(pass) {
				continue
			}
			env.StartRule(rule)
			match, errs := scan.AnalyzeFile(rule, env)
			if match != nil {
				data.Matches = append(data.Matches, match)
				processMatch(m.Config, data, match)
			}
			for _, err := range errs {
				data.RegisterError(err)
			}
		}
	}
	processTags(m.Config, data)

	// this file may have created new files, scan them too
	for _, offspring := range data.Children {
		scanFile(m, env, offspring.Filename, data)
	}
}

// ScanData scans a byte vector for matches.
func ScanData(m *types.Molly, data []byte) (*types.Report, error) {

	// we need a dummy file name that is unique:
	var fd *types.FileData
	for i := 0; fd == nil; i++ {
		dummyname := fmt.Sprintf("nopath/nofile_%04d", i)
		if _, found := m.Files[dummyname]; !found {
			fd = types.NewFileData(dummyname, nil)
			m.Files[dummyname] = fd
			m.Report.Add(fd)
		}
	}
	fd.Filesize = int64(len(data))

	env := types.NewEnv(m)
	reader := bytes.NewReader(data)
	scanInput(m, env, reader, fd)
	m.Report = m.Report.RemoveEmpty()

	return m.Report, nil
}
