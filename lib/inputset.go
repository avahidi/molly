package lib

import (
	"os"
	"path/filepath"

	"bitbucket.org/vahidi/molly/lib/util/logging"
)

// inputs is a queue for files being parsed
type inputs struct {
	processed []string
	queue     []string
	seen      map[string]bool
}

// newinputs creates an empty inputs given a output directory
func newinputset() *inputs {
	return &inputs{seen: make(map[string]bool)}
}

func (i inputs) Files() []string {
	return i.processed
}

func (i *inputs) Push(paths ...string) {
	for _, path := range paths {
		if _, seen := i.seen[path]; !seen {
			i.seen[path] = true
			i.queue = append(i.queue, path)
		}
	}
}

func (i *inputs) popOne() (string, bool) {
	n := len(i.queue)
	if n == 0 {
		return "", false
	}
	filename := i.queue[n-1]
	i.queue = i.queue[:n-1]
	return filename, true
}

func (i *inputs) Pop() string {

	for {
		path, valid := i.popOne()
		if !valid {
			return ""
		}
		fi, err := os.Stat(path)
		if err != nil {
			return path // let someone else take care of the error
		}

		mode := fi.Mode()
		if mode.IsDir() {
			dir, err := os.Open(path)
			if err != nil {
				return path // let someone else take care of the error
			}
			defer dir.Close()

			files, err := dir.Readdir(0)
			if err != nil {
				return path // let someone else take care of the error
			}
			for _, file := range files {
				i.Push(filepath.Join(path, file.Name()))
			}
		} else if mode.IsRegular() {
			i.processed = append(i.processed, path)
			return path
		} else {
			logging.Warningf("ignoring unknown file type '%s'\n", path)
		}
	}
}
