package util

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Extensions returns all extensions of a file, e.g. a.tar.gz -> ["gz","tar"]
func Extensions(filename string) []string {
	var ret []string
	for {
		ext := filepath.Ext(filename)
		if ext == "" || len(ext) > 8 { /* assume short extensions */
			return ret
		}
		filename = filename[:len(filename)-len(ext)]
		ret = append(ret, ext[1:])
	}
}

// SanitizeFilename performs file name sanitization
func SanitizeFilename(filename string, filter func(rune) bool) string {
	var buf bytes.Buffer
	if filter == nil {
		filter = func(r rune) bool {
			const badchars = "\\;<>?* \000"
			return strings.IndexRune(badchars, r) != -1 || !strconv.IsPrint(r)
		}
	}

	filename = strings.Replace(filename, "..", "_", -1)
	for _, r := range filename {
		if r == 0 {
			break
		}
		if filter(r) {
			buf.WriteRune('_')
		} else {
			buf.WriteRune(r)
		}
	}
	return buf.String()
}

// ReadUntil reads from stream until it sees byte "until" or
// gathers "maxsize" bytes (ignored if zero or less)
func ReadUntil(r io.Reader, until byte, maxsize int) ([]byte, bool, error) {
	var bw bytes.Buffer
	br := bufio.NewReaderSize(r, 1)
	for i := 0; i < maxsize; i++ {
		c, err := br.ReadByte()
		if err != nil {
			return nil, false, err
		}
		if c == until {
			return bw.Bytes(), true, nil
		}
		bw.WriteByte(c)
	}
	return bw.Bytes(), false, nil
}

// ReadStructAt is a helper for loading a struct from a specific offset
func ReadStructAt(r io.ReadSeeker, offset int64, endian binary.ByteOrder, data interface{}) error {
	if _, err := r.Seek(offset, os.SEEK_SET); err != nil {
		return err
	}
	return binary.Read(r, endian, data)
}

type readeratwrapper struct {
	r io.ReadSeeker
}

func (rsa readeratwrapper) ReadAt(p []byte, off int64) (int, error) {
	_, err := rsa.r.Seek(off, os.SEEK_SET)
	if err != nil {
		return 0, err
	}
	return rsa.r.Read(p)
}

// NewReaderAt turn a ReadSeeker into a ReaderAt for the rare cases its needed
func NewReaderAt(r io.ReadSeeker) io.ReaderAt {
	return &readeratwrapper{r: r}
}

// BufreaderAt is a helper for creating a buffered reader at a position
func BufreaderAt(r io.ReadSeeker, offset int64) (*bufio.Reader, error) {
	if _, err := r.Seek(offset, os.SEEK_SET); err != nil {
		return nil, err
	}
	return bufio.NewReader(r), nil
}
