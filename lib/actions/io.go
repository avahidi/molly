package actions

import (
	"fmt"
	"io"
	"os"

	"bitbucket.org/vahidi/molly/lib/actions/extractors"
	"bitbucket.org/vahidi/molly/lib/types"
)

func extractFunction(e *types.Env, filename string, offset int64, size int64, queue bool) (string, error) {
	fmt.Printf("extracting %d bytes into %s...\n", size, filename)

	// input file
	if _, err := e.Reader.Seek(offset, os.SEEK_SET); err != nil {
		return "", err
	}

	// output file:
	w, err := e.Create(filename)
	if err != nil {
		return "", err
	}
	defer w.Close()
	_, err = io.CopyN(w, e.Reader, size)
	return w.Name(), err
}

var extractorlist = map[string]func(*types.Env, string, string) error{
	"zip":  extractors.Unzip,
	"tar":  extractors.Untar,
	"cpio": extractors.Uncpio,
}

// RegisterExtractor provides a method to register user extractor function
func RegisterExtractor(typ string, extoractor func(*types.Env, string, string) error) {
	extractorlist[typ] = extoractor
}

func decompressFunction(e *types.Env, typ string, prefix string) (string, error) {
	filename := e.GetFile()
	f, found := extractorlist[typ]
	if !found {
		return "", fmt.Errorf("Unknown compression format: '%s'", typ)
	}

	fmt.Printf("Extracting '%s' file '%s'...\n", typ, filename)
	return filename, f(e, filename, prefix)
}

func fileFunction(e *types.Env, prefix string) (string, error) {
	return e.Name(prefix, true)
}

func dirFunction(e *types.Env, prefix string) (string, error) {
	return e.Mkdir(prefix)
}

// slice the current file
func sliceFunction(e *types.Env, prefix string, positions ...uint64) (int64, error) {
	if len(positions) == 0 || len(positions)%2 != 0 {
		return 0, fmt.Errorf("Wrong number of parameters in slice()")
	}

	total := e.GetSize()
	file, err := e.Create(prefix)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	var count int64
	for i := 0; i < len(positions); i += 2 {
		start, end := positions[i], positions[i+1]
		if start < 0 || start >= end || end > total {
			return 0, fmt.Errorf("invalid boundaries in slice(): %d-%v", start, end)
		}
		if _, err := e.Reader.Seek(int64(start), os.SEEK_SET); err != nil {
			return 0, err
		}
		if _, err := io.CopyN(file, e.Reader, int64(end-start)); err != nil {
			return 0, err
		}
	}
	return count, nil
}

func init() {
	types.FunctionRegister("extract", extractFunction)
	types.FunctionRegister("file", fileFunction)
	types.FunctionRegister("dir", dirFunction)
	types.FunctionRegister("decompress", decompressFunction)
	types.FunctionRegister("slice", sliceFunction)
}
