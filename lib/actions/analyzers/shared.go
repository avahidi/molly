package analyzers

import (
	"bufio"
	"bytes"
	"io"
	"strings"
)

// extractStrings extract strings from reader, similar to UNIX strings utility
func extractStrings(r io.Reader, minsize int) ([]string, error) {
	bs := bytes.Buffer{}
	br := bufio.NewReader(r)
	var ret []string

	for {
		c, err := br.ReadByte()
		if err != nil && err != io.EOF {
			return ret, err
		}
		if err == nil && c >= 0x20 && c < 0x7f { // BSD isprint(c)
			bs.WriteByte(c)
		} else {
			if bs.Len() >= minsize {
				ret = append(ret, string(bs.Bytes()))
			}
			bs.Reset()
		}
		if err == io.EOF {
			return ret, nil
		}
	}
}

// containsOnly returns true of text contains only letters in chars
func containsOnly(text, chars string) bool {
	for _, r := range text {
		if !strings.ContainsRune(chars, r) {
			return false
		}
	}
	return true
}
