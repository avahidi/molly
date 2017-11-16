package molly

import (
	"path"

	"bitbucket.org/vahidi/molly/util"
)

type InputSet struct {
	Processed []string
	queue     []string
	outputDir string
}

func newInputSet(dir string) *InputSet {
	return &InputSet{outputDir: dir}
}

func (i *InputSet) CreateNew(suggedtedname string) string {
	filename := path.Join(i.outputDir, util.SanitizeFilename(suggedtedname, nil))
	i.queue = append(i.queue, filename)
	return filename
}

/*
func (i *InputSet) Push(paths ...string) {
	var fl util.FileList
	if err := fl.Walk(paths...); err != nil {
		log.Fatal("Error while adding files: %v\n", err)
	}
	for _, f := range fl {
		i.queue = append(i.queue, f)
	}
}
*/

func (i *InputSet) Push(path string) {
	var fl util.FileList
	fl.Walk(path)

	for _, f := range fl {
		i.queue = append(i.queue, f)
	}
}

func (i *InputSet) Pop() string {
	n := len(i.queue)
	if n == 0 {
		return ""
	}
	filename := i.queue[n-1]
	i.queue = i.queue[:n-1]
	i.Processed = append(i.Processed, filename)
	return filename
}
