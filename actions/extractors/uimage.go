package extractors

import (
	"encoding/binary"
	"fmt"
	"io"

	"bitbucket.org/vahidi/molly/types"
	"bitbucket.org/vahidi/molly/util"
)

const (
	uimageMagic      = 0x27051956
	uimageHeaderSize = 64
	uimageTypeKernel = 2
	uimageTypeMulti  = 4
)

var uimageComp = map[uint8]string{1: "gz", 2: "bz2", 3: "lzma", 4: "lzo"}
var uimageOs = map[uint8]string{
	5: "linux", 13: "lynxos", 14: "vxworks", 16: "qnx", 17: "uboot", 22: "ose"}
var uimageArch = map[uint8]string{2: "arm", 3: "i386", 4: "ia64", 5: "mips"}

// UnUimage extracts u-boot uimage files
func UnUimage(e *types.Env, prefix string) (string, error) {
	img := util.Structured{Reader: e.Input, Order: binary.BigEndian}
	var head struct {
		Magic    uint32
		Hcrc     uint32
		Time     uint32
		Size     uint32
		LoadAdr  uint32
		EntryAdr uint32
		Dcrc     uint32
		OS       uint8
		Arch     uint8
		Type     uint8
		Comp     uint8
		Name     [32]byte
	}

	if err := img.ReadAt(0, &head); err != nil {
		return "", err
	}
	if head.Magic != uimageMagic {
		return "", fmt.Errorf("uimage: file is not an uimage")
	}

	// decide filename
	name := prefix + util.AsciizToString(head.Name[:])
	if arch, found := uimageArch[head.Arch]; found {
		name = name + "_" + arch
	}
	if os, found := uimageOs[head.OS]; found {
		name = name + "_" + os
	}
	if head.Type != uimageTypeMulti && head.LoadAdr != 0 {
		name = fmt.Sprintf("%s_%08x", name, head.LoadAdr)
	}
	if ext, found := uimageComp[head.Comp]; found {
		name = name + "." + ext
	}

	// multi-image?
	if head.Type == uimageTypeMulti {
		return "", fmt.Errorf("TODO: implement multi image uimage")
	}

	// single image
	w, err := e.Create(name)
	if err != nil {
		return "", err
	}

	_, err = io.CopyN(w, img.Reader, int64(head.Size))
	return "", err
}
