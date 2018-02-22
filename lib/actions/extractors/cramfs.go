package extractors

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	//	"compress/zlib"
	"bitbucket.org/vahidi/molly/lib/types"
	"bitbucket.org/vahidi/molly/lib/util"
)

type cramContext struct {
	Head  cramHead
	Order binary.ByteOrder
}
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
	Root      creamInode
}
type creamInode struct {
	Mode            uint16
	UID             uint16
	SizeWidth       uint32 // 24-8
	NameWidthOffset uint32 // 6-26
}

// XXX: these may need to be adjusted for endianness
func (c creamInode) Size() uint32    { return (c.SizeWidth & 0xFFFFFF) }
func (c creamInode) GID() uint32     { return uint32(c.SizeWidth >> 24) }
func (c creamInode) NameLen() uint32 { return uint32(c.NameWidthOffset&63) * 4 }
func (c creamInode) Ofsset() uint32  { return uint32(c.NameWidthOffset>>6) * 4 }

func extractStructAt(r io.ReadSeeker, offset int64, endian binary.ByteOrder, data interface{}) error {
	if _, err := r.Seek(offset, os.SEEK_SET); err != nil {
		return err
	}
	return binary.Read(r, endian, data)
}

func uncreamInodeDir(e *types.Env, ctx *cramContext, inode *creamInode, name string) error {
	r := e.Reader
	offset := int64(inode.Ofsset())
	end := offset + int64(inode.Size())
	for offset < end {
		var next creamInode
		if err := extractStructAt(r, offset, ctx.Order, &next); err != nil {
			return err
		}
		dirname := make([]byte, next.NameLen())
		r.Read(dirname)
		uncreamInode(e, ctx, &next, filepath.Join(name, string(dirname)))
		offset += 12 + int64(next.NameLen())
	}
	return nil
}

func uncreamInodeFile(e *types.Env, ctx *cramContext, inode *creamInode, name string) error {
	r := e.Reader
	offset := int64(inode.Ofsset())
	if _, err := r.Seek(offset, os.SEEK_SET); err != nil {
		return err
	}
	w, err := e.Create(name)
	if err != nil {
		return err
	}
	defer w.Close()
	_, err = io.CopyN(w, r, int64(inode.Size()))
	return err
}

func uncreamInode(e *types.Env, ctx *cramContext, inode *creamInode, name string) error {
	const S_IFMT uint16 = 0170000
	const S_IFDIR uint16 = 0040000
	const S_IFREG uint16 = 0100000

	switch inode.Mode & S_IFMT {
	case S_IFDIR:
		return uncreamInodeDir(e, ctx, inode, name)
	case S_IFREG:
		return uncreamInodeFile(e, ctx, inode, name)
	default:
		util.RegisterWarningf("Warning: ignoring unknown file type: %08x\n", inode.Mode)
	}
	return nil
}

// Uncramfs attempts to unpack a cramfs image
func Uncramfs(e *types.Env, prefix string) (string, error) {
	const Magic = 0x28cd3d45
	const Signature = "Compressed ROMFS"

	r := e.Reader
	ctx := &cramContext{Order: binary.LittleEndian}

	// get header while figuring out byte order:
	if err := extractStructAt(r, 0, ctx.Order, &ctx.Head); err != nil {
		return "", err
	}
	if ctx.Head.Magic != Magic {
		ctx.Order = binary.BigEndian
		if err := extractStructAt(r, 0, ctx.Order, &ctx.Head); err != nil {
			return "", err
		}
	}
	// valid cramfs file?
	if ctx.Head.Magic != Magic || string(ctx.Head.Signature[:]) != Signature {
		return "", fmt.Errorf("file is not a cramfs")
	}
	return prefix, uncreamInode(e, ctx, &ctx.Head.Root, prefix)
}
