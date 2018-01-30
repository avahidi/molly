package util

import (
	"bufio"
	"bytes"
	"io"
	"strconv"
	"strings"
)

// file name sanitization
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