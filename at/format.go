package at

import (
	"encoding/binary"
	"log"
)

type Format int

const (
	UBE Format = iota
	SBE
	ULE
	SLE
)

func (f Format) ByteOrder() binary.ByteOrder {
	if f == UBE || f == SBE {
		return binary.BigEndian
	}
	return binary.LittleEndian
}

func (f Format) Signed() bool {
	return f == SBE || f == SLE
}

func (f Format) Extract(bytes int, data []byte) uint64 {
	bo := f.ByteOrder()
	ret := uint64(0)
	switch bytes {
	case 1:
		ret = uint64(data[0])
	case 2:
		ret = uint64(bo.Uint16(data))
	case 4:
		ret = uint64(bo.Uint32(data))
	case 8:
		ret = bo.Uint64(data)
	default:
		log.Panicf("Internal error: invalid number length: %d %v", bytes, f)
	}

	// sign extend
	if f.Signed() {
		mask := uint64(0xFFFFFFFFFFFFFF80) << uint64(8*(bytes-1))
		if (mask & ret) != 0 {
			ret |= mask
		}
	}

	return ret
}
