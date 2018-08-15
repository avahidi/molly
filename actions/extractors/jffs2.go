package extractors

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"fmt"
	"os"
	"path"
	"sort"
	"time"

	"bitbucket.org/vahidi/molly/types"
	"bitbucket.org/vahidi/molly/util"
)

const (
	jffs2Magic        = 0x1985
	jffs2NodeAccurate = 0x2000

	// type feature flags
	jffs2FeatureRwcompatDelete = 0x0000
	jffs2FeatureRwcompatCopy   = 0x4000
	jffs2FeatureRocompat       = 0x8000
	jffs2FeatureIncompat       = 0xC000
	jffs2FeatureMask           = 0xC000

	// types
	jffs2NodetypeDirent      = 0x0001
	jffs2NodetypeInode       = 0x0002
	jffs2NodetypeCleanmarker = 0x0003
	jffs2NodetypeMask        = 0x000F

	// compression
	jffs2ComprNone = 0x00
	jffs2ComprZlib = 0x06
)

// these are in syscall but thats not available on Windows:
const (
	DT_DIR = 0x4
	DT_REG = 0x8
	DT_LNK = 0xa
)

// jinode is our internal representation of an inode
type jinode struct {
	version     uint32
	srcSize     uint32
	dstSize     uint32
	srcOffset   uint32
	dstOffset   uint32
	compression uint8
}

// jdatalist is a list of inodes, which together form contents of a file
type jdatalist []*jinode

// sort.Interface
var _ sort.Interface = (*jdatalist)(nil)

func (j jdatalist) Len() int           { return len(j) }
func (j jdatalist) Less(a, b int) bool { return j[a].version < j[b].version }
func (j jdatalist) Swap(a, b int)      { j[a], j[b] = j[b], j[a] }

func (j jdatalist) size() int {
	var size int
	for _, v := range j {
		if size < int(v.dstSize+v.dstOffset) {
			size = int(v.dstSize + v.dstOffset)
		}
	}
	return size
}

func (j jdatalist) generate(src *jcontext) ([]byte, error) {
	size := j.size()
	data := make([]byte, size)
	for i := 0; i < size; i++ {
		data[i] = 0xFF // flash default
	}

	sort.Sort(j) // sort by version -> old data will be overwritten by new
	for _, v := range j {
		srcData := make([]byte, v.srcSize)
		if err := src.ReadAt(int64(v.srcOffset), srcData); err != nil {
			return nil, err
		}
		switch v.compression {
		case jffs2ComprNone:
			copy(data[v.dstOffset:], srcData)
		case jffs2ComprZlib:
			bin := bytes.NewReader(srcData)
			zr, err := zlib.NewReader(bin)
			if err != nil {
				return nil, err
			}
			defer zr.Close()
			if _, err := zr.Read(data[v.dstOffset : v.dstOffset+v.dstSize-1]); err != nil {
				return nil, err
			}
		default:
			return nil, fmt.Errorf("Unknown jffs2 compression format '%d'", v.compression)
		}
	}

	return data, nil
}

// jdnode is out internal representation of a dir entry
type jdnode struct {
	name     string
	version  uint32
	time     uint32
	typ      uint8
	parent   *jdnode
	children []*jdnode
	file     jdatalist
}

// jcontext is our internal context holder
type jcontext struct {
	util.Structured
	Create  func(string) (*os.File, *types.FileData, error)
	nodemap map[uint32]*jdnode
}

type jffs2Header struct {
	Magic  uint16
	Type   uint16
	Length uint32
	Crc    uint32
}

const jffsDirentSize = 40 // sizeof(jffs2Header) + sizeof(jffs2Dirent)
type jffs2Dirent struct {
	Pino     uint32
	Version  uint32
	Ino      uint32
	Mctime   uint32
	Nsize    uint8
	Type     uint8
	Reserved uint16
	Crc      [2]uint32
}

const jffsDataOffset = 68 // sizeof(jffs2Header) + sizeof(jffs2Inode)
type jffs2Inode struct {
	Ino     uint32
	Version uint32
	Mode    uint32
	Uid     uint16
	Guid    uint16
	Isize   uint32
	ATime   uint32 // access time
	MTime   uint32 // modification time
	CTime   uint32 // change time (??)
	Ofsset  uint32
	CSize   uint32
	DSize   uint32
	Compr   [2]uint8
	Flags   uint16
	Crc     [2]uint32
}

func (c *jcontext) scan(prefix string, offset int64) error {
	var head jffs2Header
	for {
		if err := c.ReadAt(offset, &head); err != nil {
			return nil // not really an error...
		}
		if head.Magic != jffs2Magic {
			return nil
		}

		if (head.Type & jffs2NodeAccurate) == jffs2NodeAccurate {
			switch head.Type & jffs2NodetypeMask {
			case jffs2NodetypeDirent:
				var dirent jffs2Dirent
				if err := c.Read(&dirent); err != nil {
					return err
				}
				name := make([]uint8, dirent.Nsize)
				if err := c.Read(&name); err != nil {
					return err
				}
				// keep it only if its the most recent version
				p, found := c.nodemap[dirent.Ino]
				if !found || p.version < dirent.Version {
					jdnode := &jdnode{
						name:    string(name),
						typ:     dirent.Type,
						time:    dirent.Mctime,
						version: dirent.Version,
					}
					c.nodemap[dirent.Ino] = jdnode
					if p, found := c.nodemap[dirent.Pino]; found {
						jdnode.parent = p
						p.children = append(p.children, jdnode)
					}
				}
			case jffs2NodetypeInode:
				var inode jffs2Inode
				if err := c.Read(&inode); err != nil {
					return err
				}
				p, found := c.nodemap[inode.Ino]
				if !found {
					return fmt.Errorf("inode at offset %08x is missing parent %d", offset, inode.Ino)
				}
				jinode := &jinode{
					version:     inode.Version,
					srcOffset:   uint32(offset + jffsDataOffset),
					dstOffset:   inode.Ofsset,
					srcSize:     inode.CSize,
					dstSize:     inode.DSize,
					compression: inode.Compr[0],
				}
				p.file = append(p.file, jinode)
			case jffs2NodetypeCleanmarker:
				break
			default:
				comp := head.Type & jffs2FeatureMask
				if comp == jffs2FeatureIncompat {
					return fmt.Errorf("JFFS2 node type %04x is unknown", head.Type)
				}
			}
		}
		// update offset to next 32-bit aligned position
		offset = offset + int64(head.Length)
		offset = (offset + 3) & ^int64(3)
	}
}

func (c *jcontext) writeFile(prefix string, j *jdnode) error {
	data, err := j.file.generate(c)
	if err != nil {
		return err
	}
	/*
		if len(data) > 0 { // remove EOF?
			data = data[:len(data)-1]
		}*/
	file, fd, err := c.Create(path.Join(prefix, j.name))
	if err != nil {
		return err
	}
	defer file.Close()

	fd.SetTime(time.Unix(int64(j.time), 0))
	_, err = file.Write(data)
	return err
}

func (c *jcontext) writeLink(prefix string, j *jdnode) error {
	data, err := j.file.generate(c)
	if err != nil {
		return err
	}
	file, _, err := c.Create(path.Join(prefix, j.name+".link"))
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.Write(data)
	return err
}

func (c *jcontext) create(prefix string, j *jdnode) error {
	switch int(j.typ) {
	case DT_DIR:
		for _, ch := range j.children {
			if err := c.create(path.Join(prefix, j.name), ch); err != nil {
				return err
			}
		}
	case DT_REG:
		return c.writeFile(prefix, j)
	case DT_LNK:
		return c.writeLink(prefix, j)
	default:
		fmt.Printf("jff2s: ignoring file of type %08x\n", j.typ)
	}
	return nil
}

// findEndian figures out what endian we have
func (c *jcontext) findEndian() error {
	for _, c.Order = range []binary.ByteOrder{binary.BigEndian, binary.LittleEndian} {
		var head jffs2Header
		if err := c.ReadAt(0, &head); err != nil {
			return err
		}
		if head.Magic == jffs2Magic {
			return nil
		}
	}
	return fmt.Errorf("file is not JFFS2")
}

// findPageSize figures out what page size we have
func (c jcontext) findPageSize(filesize int64) (int64, error) {
	// TODO: parse file to figure out what the page size is
	return 0x10000, nil
}

// Unjffs2 will attempt to unpack a JFFS2 file
//
// It is based on Woodhouse's paper + kernel headers.
//
// NOTE: link files are created as regular files named <name>.link
func Unjffs2(e *types.Env, prefix string) (string, error) {
	ctx := &jcontext{Create: e.Create, nodemap: make(map[uint32]*jdnode)}
	ctx.Reader = e.Reader

	// endian?
	if err := ctx.findEndian(); err != nil {
		return "", err
	}

	// flash page size?
	filesize := int64(e.GetSize())
	pagesize, err := ctx.findPageSize(filesize - jffsDirentSize)
	if err != nil {
		return "", err
	}

	// get the nodes
	for offset := int64(0); offset < filesize-jffsDirentSize; offset += pagesize {
		if err := ctx.scan(prefix, offset); err != nil {
			return "", err
		}
	}

	// start from root nodes and create all files
	for _, v := range ctx.nodemap {
		if v.parent == nil {
			if err := ctx.create(prefix, v); err != nil {
				return "", err
			}
		}
	}

	return "", nil
}
