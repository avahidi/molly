package extractors

import (
	"compress/gzip"
	"fmt"
	"io"

	"bitbucket.org/vahidi/molly/types"
	"bitbucket.org/vahidi/molly/util"
)

// Ungzip extracts a gzip file
func Ungzip(e *types.Env, prefix string) (string, error) {

	gr, err := gzip.NewReader(e.Input)
	if err != nil {
		return "", err
	}

	// figure out what extension to use and then compute the name
	exts := util.Extensions(e.GetFile())
	var name string
	if len(exts) > 0 && exts[0] != "gz" && exts[0] != "GZ" {
		name = fmt.Sprintf("%s.%s", prefix, exts[0])
	} else if len(exts) > 1 {
		name = fmt.Sprintf("%s.%s", prefix, exts[1])
	} else {
		name = prefix
	}

	// open output file and unpack gzip to it
	w, err := e.Create(name)
	if err != nil {
		return "", err
	}
	defer w.Close()

	io.Copy(w, gr)
	return w.Name(), nil
}
