package analyzers

// dalvik bytecode functions,
// see https://source.android.com/devices/tech/dalvik/dalvik-bytecode.html

import (
	"fmt"
	"strings"
)

// javaIsInnerName return true if this is a inner-class or a member of one
func javaIsInnerName(name string) bool {
	return strings.Contains(name, "$")
}

// javaTypeToClassName extracts class complete name from type name
// i.e. "Lcom/example/<someclass>;" to "com.example.someclass"
func javaTypeToClassName(typename string) (string, error) {
	if !strings.HasPrefix(typename, "L") || !strings.HasSuffix(typename, ";") {
		return "", fmt.Errorf("Invalid class type: '%s'", typename)
	}
	return strings.Replace(typename[1:len(typename)-1], "/", ".", -1), nil
}

// javaExtractPackageName extracts package name from class name
// some.package.example.Class -> some.package.example
func javaExtractPackageName(classname string) string {
	n := strings.LastIndex(classname, ".")
	if n == -1 {
		return ""
	}
	return classname[:n]
}

func dalvikType(op uint16) string {
	c := uint8(op & 0xFF)
	switch c {
	case 0x28:
		return "10t"
	case 0x00, 0x0e, 0x73:
		return "10x"
	case 0x12:
		return "11n"
	case 0x0a, 0x0b, 0x0c, 0x0d, 0x0f, 0x10, 0x11, 0x1d, 0x1e, 0x27:
		return "11x"
	case 0x01, 0x04, 0x07, 0x21:
		return "12x"
	case 0x29:
		return "20t"
	case 0x1a, 0x1c, 0x1f, 0x22, 0xfe, 0xff:
		return "21c"
	case 0x15, 0x19:
		return "21h"
	case 0x13, 0x16:
		return "21s"
	case 0x20, 0x23:
		return "22c"
	case 0x02, 0x05, 0x08:
		return "22x"
	case 0x2a:
		return "30t"
	case 0x1b:
		return "31c"
	case 0x14, 0x17:
		return "31i"
	case 0x26, 0x2b, 0x2c:
		return "31t"
	case 0x03, 0x06, 0x09:
		return "32x"
	case 0x24, 0xfc:
		return "35c"
	case 0x25, 0xfd:
		return "3rc"
	case 0xfa, 0xfb:
		return "4rcc"
	case 0x18:
		return "51l"
	}
	switch {
	case c >= 0x3e && c <= 0x43,
		c >= 0x79 && c <= 0x7a,
		c >= 0xe3 && c <= 0xf9:
		return "10x"
	case c >= 0x7b && c <= 0x8f,
		c >= 0xb0 && c <= 0xcf:
		return "12x"
	case c >= 0x60 && c <= 0x6d:
		return "21c"
	case c >= 0x38 && c <= 0x3d:
		return "21t"
	case c >= 0xd8 && c <= 0xe2:
		return "22b"
	case c >= 0x52 && c <= 0x5f:
		return "22c"
	case c >= 0xd0 && c <= 0xd7:
		return "22s"
	case c >= 0x32 && c <= 0x37:
		return "22t"
	case c >= 0x2d && c <= 0x31,
		c >= 0x44 && c <= 0x51,
		c >= 0x90 && c <= 0xaf:
		return "23x"
	case c >= 0x6e && c <= 0x72:
		return "35c"
	case c >= 0x74 && c <= 0x78:
		return "3rc"
	}
	return "" // should not reach this
}

func dalvikAnalyze(insts []uint16, index int) (uint8, int, string) {
	inst := insts[index]
	typ := dalvikType(inst)
	op := uint8(inst & 0xFF)
	size := int(typ[0] - '0')

	// special cases where NOP contains data
	if op == 0x00 {
		// get instruction after this
		count := 0
		if index < len(insts)-2 {
			count = int(insts[index+1])
		}

		switch inst >> 8 {
		// packed-switch-payload format
		case 0x01:
			size = count*2 + 4

		// sparse-switch-payload format
		case 0x02:
			size = count*4 + 2

		case 0x03:
			if index < len(insts)-4 {
				usize := (int(insts[index+2])) + (int(insts[index+3]) << 16)
				size = (usize*count+1)/2 + 4
			}
		}
	}

	// this must be an error
	if size < 1 {
		size = 1
	}

	return op, size, typ
}

// dalvikOpNew returns type index if this is a new() operation
func dalvikOpNew(op uint8, inst []uint16) (bool, uint16) {
	if op == 0x22 {
		return true, inst[1]
	}
	return false, 0
}

// dalvikOpInvoke returns method index if this is a invoke operation
func dalvikOpInvoke(op uint8, inst []uint16) (bool, uint16) {
	switch op {
	case 0x6e, 0x6f, 0x70, 0x71, 0x72: // invoke
		return true, inst[1]
	default:
		return false, 0
	}
}
