package extractors

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/avahidi/molly/types"
)

type cpioFileHead struct {
	namesize int
	filesize int64
	mtime    int64
}

func cpioBinaryParser(r io.Reader) (*cpioFileHead, error) {
	var head struct {
		Magic     uint16
		Garbage   [7]uint16
		MTime     [2]uint16
		NameSize  uint16
		FileSizes [2]uint16
	}
	if err := binary.Read(r, binary.LittleEndian, &head); err != nil {
		return nil, err
	}
	if head.Magic != 0x71c7 {
		return nil, fmt.Errorf("Not CPIO header?")
	}

	return &cpioFileHead{
		namesize: int(head.NameSize),
		filesize: int64(head.FileSizes[1]) + (int64(head.FileSizes[0]) << 16),
		mtime:    int64(head.MTime[1]) + (int64(head.MTime[0]) << 16),
	}, nil
}

func cpioAsciiParser(r io.Reader) (*cpioFileHead, error) {
	var head struct {
		Magic    [6]byte
		Garbage  [42]byte
		MTime    [11]byte
		NameSize [6]byte
		FileSize [11]byte
	}
	if err := binary.Read(r, binary.LittleEndian, &head); err != nil {
		return nil, err
	}
	if string(head.Magic[:]) != "070707" {
		return nil, fmt.Errorf("Not CPIO header?")
	}
	ns, err := strconv.ParseInt(string(head.NameSize[:]), 8, 64)
	if err != nil {
		return nil, err
	}
	fs, err := strconv.ParseInt(string(head.FileSize[:]), 8, 64)
	if err != nil {
		return nil, err
	}

	// cpio ascii mdate is octal ascii :(
	mtime, err := strconv.ParseInt(string(head.MTime[:]), 8, 32)

	return &cpioFileHead{
		namesize: int(ns),
		filesize: int64(fs),
		mtime:    mtime,
	}, err
}

func Uncpio(e *types.Env, prefix string) (string, error) {
	var parser func(io.Reader) (*cpioFileHead, error)
	mustPad := false
	r := e.Reader

	// vad type is this?
	magic := make([]byte, 6)
	if _, err := r.Read(magic); err != nil {
		return "", err
	}
	if string(magic) == "070707" {
		parser = cpioAsciiParser
	} else if string(magic) == "070701" || string(magic) == "070702" {
		// parser = cpioAsciiParser ???
	} else if magic[0] == 0xC7 && magic[1] == 0x71 {
		parser = cpioBinaryParser
		mustPad = true
	}
	if parser == nil {
		return "", fmt.Errorf("Unknown cpio format: magic=%v", magic)
	}

	// start parsing the file
	if _, err := r.Seek(0, os.SEEK_SET); err != nil {
		return "", err
	}
	pad := make([]byte, 1)
	for {
		fh, err := parser(r)
		if err != nil {
			return "", err
		}

		// read name
		name := make([]byte, fh.namesize)
		r.Read(name)
		if mustPad && fh.namesize%2 == 1 {
			r.Read(pad)
		}
		if name[len(name)-1] == 0 {
			name = name[:len(name)-1]
		}
		if string(name) == "TRAILER!!!" {
			return "", nil
		}

		// copy file contents
		if fh.filesize > 0 {

			w, d, err := e.Create(prefix + string(name))
			if err != nil {
				return "", err
			}
			defer w.Close()

			d.SetTime(time.Unix(fh.mtime, 0))
			if _, err = io.CopyN(w, r, fh.filesize); err != nil {
				return "", err
			}
			if mustPad && fh.filesize%2 == 1 {
				r.Read(pad)
			}
		}
	}
}
