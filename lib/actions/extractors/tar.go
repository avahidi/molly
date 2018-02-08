package extractors

import (
	"archive/tar"
	"bitbucket.org/vahidi/molly/lib/types"
	"io"
)

func Untar(e *types.Env, prefix string) (string, error) {
	r := e.Reader

	tr := tar.NewReader(r)
	for {
		h, err := tr.Next()
		if err == io.EOF {
			return "", nil // not really an error...
		}
		if h == nil || err != nil {
			return "", err
		}

		if !h.FileInfo().IsDir() {
			w, err := e.Create(prefix + h.Name)
			if err != nil {
				return "", err
			}
			defer w.Close()
			if _, err := io.CopyN(w, tr, h.Size); err != nil {
				return "", err
			}
		}
	}
}
