package extractors

import (
	"archive/tar"
	"io"

	"github.com/avahidi/molly/types"
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
			w, d, err := e.Create(prefix + h.Name)
			if err != nil {
				return "", err
			}
			defer w.Close()

			d.SetTime(h.ChangeTime)
			if _, err := io.CopyN(w, tr, h.Size); err != nil {
				return "", err
			}
		}
	}
}
