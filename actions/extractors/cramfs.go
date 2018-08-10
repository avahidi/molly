package extractors

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"bitbucket.org/vahidi/molly/types"
	"bitbucket.org/vahidi/molly/util"
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

type cramContext struct {
	util.Structured
	Create func(string, *time.Time) (*os.File, error)
}

func (c cramContext) inodeDir(inode *cramInode, name string) error {
	offset := int64(inode.Ofsset())
	end := offset + int64(inode.Size())
	for offset < end {
		var next cramInode
		if err := c.ReadAt(offset, &next); err != nil {
			return err
		}

		// extract the name, strip the zero padding
		nlen := int(next.NameLen())
		nbuf := make([]byte, nlen)
		if _, err := c.Reader.Read(nbuf); err != nil {
			return err
		}
		dirname := util.AsciizToString(nbuf)

		if err := c.inode(&next, filepath.Join(name, dirname)); err != nil {
			return err
		}
		offset += 12 + int64(nlen)
	}
	return nil
}

func (c cramContext) inodeFile(inode *cramInode, name string) error {
	ptrOffset := int64(inode.Ofsset())
	size := int64(inode.Size())
	nblocks := (size-1)/cramBlkSize + 1

	w, err := c.Create(name, nil)
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
		if err := c.ReadAt(int64(ptr), &buf); err != nil {
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
		if err := c.ReadAt(ptrOffset, &ptr); err != nil {
			return err
		}
		ptrOffset += 4
	}

	return err
}

func (c cramContext) inode(inode *cramInode, name string) error {
	switch inode.Mode & s_IFMT {
	case s_IFDIR:
		return c.inodeDir(inode, name)
	case s_IFREG:
		return c.inodeFile(inode, name)
	default:
		util.RegisterWarningf("Warning: ignoring unknown file type: %08x", inode.Mode)
	}
	return nil
}

// Uncramfs attempts to unpack a cramfs image.
//
// This is a quick and dirty cramfs implementation and the cramfs tools that
// create these files seem to be very buggy so don't be surprised if this
// code fails to handle your images.
func Uncramfs(e *types.Env, prefix string) (string, error) {
	ctx := &cramContext{Create: e.Create}
	ctx.Reader = e.Input

	// we don't know the native byte-order, try both:
	for _, ctx.Order = range []binary.ByteOrder{binary.LittleEndian, binary.BigEndian} {
		var head cramHead
		if err := ctx.ReadAt(0, &head); err != nil {
			return "", err
		}
		if head.Magic == cramMagic && string(head.Signature[:]) == cramSignature {
			return prefix, ctx.inode(&head.Root, prefix)
		}
	}
	return "", fmt.Errorf("file is not a cramfs")
}
