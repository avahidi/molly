package util

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

// ParseNumber parses number from string, handles 0x, 0o and 0b prefixes
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

func DecodeString(bs []byte) ([]byte, error) {
	ret, len := make([]byte, 0), len(bs)

	for i := 0; i < len; i++ {
		if bs[i] == '\\' {
			i++
			if i == len {
				return nil, fmt.Errorf("end if string after '\\': %s", string(bs))
			}
			switch bs[i] {
			case '\\':
				ret = append(ret, '\\')
			case 't':
				ret = append(ret, '\t')
			case 'r':
				ret = append(ret, '\r')
			case 'n':
				ret = append(ret, '\n')

			case 'x':
				if i+2 >= len {
					return nil, fmt.Errorf("bytes missing in hex byte: %s", bs)
				}
				n, err := strconv.ParseUint(string(bs[i+1:i+3]), 16, 8)
				if err != nil {
					return nil, err
				}
				ret = append(ret, byte(n))
				i += 2
			case '0', '1', '2', '3', '4', '5', '6', '7':
				if i+2 >= len {
					return nil, fmt.Errorf("bytes missing in octal byte: %s", bs)
				}
				n, err := strconv.ParseUint(string(bs[i:i+3]), 8, 8)
				if err != nil {
					return nil, err
				}
				ret = append(ret, byte(n))
				i += 2
			}
		} else {
			ret = append(ret, bs[i])
		}
	}

	return ret, nil
}

// AsciizToString converts a null terminated byte array to a string
func AsciizToString(data []byte) string {
	if n := bytes.IndexByte(data, 0); n != -1 {
		data = data[:n]
	}
	return string(data)
}
