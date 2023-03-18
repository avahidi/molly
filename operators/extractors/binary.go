package extractors

import (
	"fmt"
	"io"
	"os"

	"github.com/avahidi/molly/types"
)

// BinarySlice creates a file from a slice
func BinarySlice(e *types.Env, name string, positions ...uint64) (string, error) {
	w, _, err := e.Create(name)
	if err != nil {
		return "", err
	}
	defer w.Close()

	var total = e.GetSize()
	for i := 0; i < len(positions); i += 2 {
		start, size := positions[i], positions[i+1]
		if start < 0 || start+size > total {
			return "", fmt.Errorf("invalid boundaries in slice(): %d-%d", start, start+size)
		}
		if _, err := e.Reader.Seek(int64(start), os.SEEK_SET); err != nil {
			return "", err
		}
		if _, err := io.CopyN(w, e.Reader, int64(size)); err != nil {
			return "", err
		}
	}
	return w.Name(), nil
}
