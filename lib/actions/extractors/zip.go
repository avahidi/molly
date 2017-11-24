package extractors

import (
	"archive/zip"
	"io"

	"bitbucket.org/vahidi/molly/lib/types"
)

func extractOneFile(fs types.FileSystem, f *zip.File, prefix string) error {
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	// in reality, we should use f.Mode() but we are replacing it
	// with our own default permissions
	if !f.FileInfo().IsDir() {
		w, err := fs.Create(prefix + f.Name)
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

func Unzip(fs types.FileSystem, filename, prefix string) error {
	r, err := zip.OpenReader(filename)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		if err := extractOneFile(fs, f, prefix); err != nil {
			return err
		}
	}
	return nil
}
