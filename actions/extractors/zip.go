package extractors

import (
	"archive/zip"
	"io"

	"bitbucket.org/vahidi/molly/types"
)

func extractOneFile(e *types.Env, f *zip.File, prefix string) error {
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	// in reality, we should use f.Mode() but we are replacing it
	// with our own default permissions
	if !f.FileInfo().IsDir() {
		w, err := e.Create(prefix+f.Name, nil)
		if err != nil {
			return err
		}
		defer w.Close()
		if _, err = io.Copy(w, rc); err != nil {
			return err
		}
	}
	return nil
}

func Unzip(e *types.Env, prefix string) (string, error) {
	r, err := zip.OpenReader(e.GetFile())
	if err != nil {
		return "", err
	}
	defer r.Close()

	for _, f := range r.File {
		if err := extractOneFile(e, f, prefix); err != nil {
			return "", err
		}
	}
	return "", nil
}
