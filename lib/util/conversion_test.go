package util

import (
	"bytes"
	"testing"
)

func TestParseNumber(t *testing.T) {
	testdata := map[string]uint64{
		"0x120":    0x120,
		"0o120":    0120,
		"0b100001": 33,
		"123":      123,
	}
	for str, ans := range testdata {
		n, err := ParseNumber(str, 64)
		if err != nil || n != ans {
			t.Errorf("Failed to parse '%s': got %d wanted %d", str, n, ans)
		}
	}
}

func TestDecodeString(t *testing.T) {
	testdata := map[string][]byte{
		"abc":           []byte("abc"),
		"\\\\\\n\\r\\t": {'\\', '\n', '\r', '\t'},
		"\\xAB\\xC0":    {0xab, 0xc0},
		"\\123\\055":    {0123, 055},
	}

	for str, ans := range testdata {
		bs, err := DecodeString([]byte(str))
		if err != nil || bytes.Compare(bs, ans) != 0 {
			t.Errorf("Failed to parse '%s': got %v wanted %v", str, bs, ans)
		}
	}
}

func TestAsciizToString(t *testing.T) {
	testdata := map[string][]byte{
		"abc": []byte{'a', 'b', 'c'},
		"xyz": []byte{'x', 'y', 'z', 0},
		"123": []byte{'1', '2', '3', 0},
		"":    []byte{0, 'g', 'n', 'u'},
	}

	for str, data := range testdata {
		ret := AsciizToString(data)
		if str != ret {
			t.Errorf("AsciizToString failed: got %v wanted %v", ret, str)
		}
	}
}
