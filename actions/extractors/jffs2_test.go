package extractors

import (
	"bytes"
	"testing"
)

func TestCopyRtime(t *testing.T) {
	var testdata = []struct {
		input, output []byte
	}{
		// from stackoverflow
		{[]byte{65, 0, 66, 0, 66, 1, 65, 4}, []byte{65, 66, 66, 66, 65, 66, 66, 66, 65}},
		// linux implementation
		{[]byte{5, 1}, []byte{5, 5}},
		{[]byte{5, 3}, []byte{5, 5, 5, 5}},
		{[]byte{1, 0, 2, 0, 3, 3}, []byte{1, 2, 3, 1, 2, 3}},
		{[]byte{1, 0, 2, 0, 3, 6, 4, 3, 4, 0}, []byte{1, 2, 3, 1, 2, 3, 1, 2, 3, 4, 1, 2, 3, 4}},
	}

	for _, test := range testdata {
		dst := make([]byte, len(test.output))
		copyRtime(dst, test.input)
		if !bytes.Equal(dst, test.output) {
			t.Errorf("copyRtime error: wanted %v, got %v", test.output, dst)
		}
	}
}
