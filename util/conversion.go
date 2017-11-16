package util

import (
	"bytes"
	"strconv"
	"strings"
)

// parse number from string
func ParseNumber(str string, bits int) (uint64, error) {
	if strings.HasPrefix(str, "0x") || strings.HasPrefix(str, "0X") {
		return strconv.ParseUint(str[2:], 16, bits)
	}
	if strings.HasPrefix(str, "0o") || strings.HasPrefix(str, "0O") {
		return strconv.ParseUint(str[2:], 8, bits)
	}
	if strings.HasPrefix(str, "0b") || strings.HasPrefix(str, "0B") {
		return strconv.ParseUint(str[2:], 2, bits)
	}
	return strconv.ParseUint(str, 10, bits)
}

// file name sanitization
func SanitizeFilename(filename string, filter func(rune) bool) string {
	var buf bytes.Buffer
	if filter == nil {
		filter = func(r rune) bool {
			const badchars = "/\\;<>?*"
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
