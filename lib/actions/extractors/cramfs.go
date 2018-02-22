package extractors

import (
	"bitbucket.org/vahidi/molly/lib/types"
	"bitbucket.org/vahidi/molly/lib/util"
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const (
	cramMagic     = 0x28cd3d45
	cramSignature = "Compressed ROMFS"
	cramBlkSize   = 4096
	cramZlibSize  = 11
	s_IFMT        = 0170000
	s_IFDIR       = 0040000
	s_IFREG       = 0100000
)

type cramHead struct {
	Magic     uint32
	Size      uint32
	Flags     uint32
	Feature   uint32
	Signature [16]uint8
	Crc       uint32
	Edition   uint32
	Block     uint32
	Files     uint32
	Name      [16]uint8
	Root      cramInode
}

type cramInode struct {
	Mode            uint16
	UID             uint16
	SizeWidth       uint32 // 24-8
	NameWidthOffset uint32 // 6-26
}

// XXX: these may need to be adjusted for endianness
func (c cramInode) Size() uint32    { return (c.SizeWidth & 0xFFFFFF) }
func (c cramInode) NameLen() uint32 { return uint32(c.NameWidthOffset&63) * 4 }
func (c cramInode) Ofsset() uint32  { return uint32(c.NameWidthOffset>>6) * 4 }

// extractStructAt is a helper for loading a struct from a specific offset
func extractStructAt(r io.ReadSeeker, offset int64, endian binary.ByteOrder, data interface{}) error {
	if _, err := r.Seek(offset, os.SEEK_SET); err != nil {
		return err
	}
	return binary.Read(r, endian, data)
}

func uncramInodeDir(e *types.Env, ord binary.ByteOrder, inode *cramInode, name string) error {
	r := e.Reader
	offset := int64(inode.Ofsset())
	end := offset + int64(inode.Size())
	for offset < end {
		var next cramInode
		if err := extractStructAt(r, offset, ord, &next); err != nil {
			return err
		}

		// extract the name, strip the zero padding
		nlen := int(next.NameLen())
		nbuf := make([]byte, nlen)
		if _, err := r.Read(nbuf); err != nil {
			return err
		}
		if n := bytes.IndexByte(nbuf, 0); n != -1 {
			nbuf = nbuf[:n]
		}
		dirname := string(nbuf)

		if err := uncramInode(e, ord, &next, filepath.Join(name, dirname)); err != nil {
			return err
		}
		offset += 12 + int64(nlen)
	}
	return nil
}

func uncramInodeFile(e *types.Env, ord binary.ByteOrder, inode *cramInode, name string) error {
	r := e.Reader
	ptrOffset := int64(inode.Ofsset())
	size := int64(inode.Size())
	nblocks := (size-1)/cramBlkSize + 1

	w, err := e.Create(name)
	if err != nil {
		return err
	}
	defer w.Close()

	ptr := uint32(ptrOffset + nblocks*4) // first block is right at the end of pointers
	buf := make([]byte, cramBlkSize+cramZlibSize)
	for size > 0 {
		if size < cramBlkSize {
			buf = buf[:cramZlibSize+size]
		}
		if err := extractStructAt(r, int64(ptr), ord, &buf); err != nil {
			return err
		}
		zr, err := zlib.NewReader(bytes.NewBuffer(buf))
		if err != nil {
			return err
		}
		defer zr.Close()

		if _, err := io.Copy(w, zr); err != nil {
			return err
		}
		size -= int64(len(buf)) - cramZlibSize
		if size <= 0 {
			break
		}
		// get the pointer to the next page
		if err := extractStructAt(r, ptrOffset, ord, &ptr); err != nil {
			return err
		}
		ptrOffset += 4
	}

	return err
}

func uncramInode(e *types.Env, ord binary.ByteOrder, inode *cramInode, name string) error {
	switch inode.Mode & s_IFMT {
	case s_IFDIR:
		return uncramInodeDir(e, ord, inode, name)
	case s_IFREG:
		return uncramInodeFile(e, ord, inode, name)
	default:
		util.RegisterWarningf("Warning: ignoring unknown file type: %08x\n", inode.Mode)
	}
	return nil
}

// Uncramfs attempts to unpack a cramfs image.
//
// This is a quick and dirty cramfs implementation and the cramfs tools that
// create these files seem to be very buggy so don't be surprised if this
// code fails to handle your images.
func Uncramfs(e *types.Env, prefix string) (string, error) {
	r := e.Reader
	// we don't know the native byte-order, try both:
	for _, order := range []binary.ByteOrder{binary.LittleEndian, binary.BigEndian} {
		var head cramHead
		if err := extractStructAt(r, 0, order, &head); err != nil {
			return "", err
		}
		if head.Magic == cramMagic && string(head.Signature[:]) == cramSignature {
			return prefix, uncramInode(e, order, &head.Root, prefix)
		}
	}
	return "", fmt.Errorf("file is not a cramfs")
}
