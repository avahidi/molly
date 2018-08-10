package extractors

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"strconv"

	"bitbucket.org/vahidi/molly/types"
)

func cpioBinaryParser(r io.Reader) (int64, int64, error) {
	var head struct {
		Magic     uint16
		Garbage   [9]uint16
		NameSize  uint16
		FileSizes [2]uint16
	}
	if err := binary.Read(r, binary.LittleEndian, &head); err != nil {
		return 0, 0, err
	}
	if head.Magic != 0x71c7 {
		return 0, 0, fmt.Errorf("Not CPIO header?")
	}
	return int64(head.NameSize),
		int64(head.FileSizes[1]) + (int64(head.FileSizes[0]) << 16), nil
}

func cpioAsciiParser(r io.Reader) (int64, int64, error) {
	var head struct {
		Magic    [6]byte
		Garbage  [53]byte
		NameSize [6]byte
		FileSize [11]byte
	}
	if err := binary.Read(r, binary.LittleEndian, &head); err != nil {
		return 0, 0, err
	}
	if string(head.Magic[:]) != "070707" {
		return 0, 0, fmt.Errorf("Not CPIO header?")
	}
	ns, err := strconv.ParseInt(string(head.NameSize[:]), 8, 64)
	if err != nil {
		return 0, 0, err
	}
	fs, err := strconv.ParseInt(string(head.FileSize[:]), 8, 64)
	return int64(ns), int64(fs), err
}

func Uncpio(e *types.Env, prefix string) (string, error) {
	var parser func(io.Reader) (int64, int64, error)
	mustPad := false
	r := e.Input

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
		namesize, filesize, err := parser(r)
		if err != nil {
			return "", err
		}

		// read name
		name := make([]byte, namesize)
		r.Read(name)
		if mustPad && namesize%2 == 1 {
			r.Read(pad)
		}
		if name[len(name)-1] == 0 {
			name = name[:len(name)-1]
		}
		if string(name) == "TRAILER!!!" {
			return "", nil
		}

		// copy file contents
		if filesize > 0 {
			w, err := e.Create(prefix+string(name), nil)
			if err != nil {
				return "", err
			}
			defer w.Close()

			if _, err = io.CopyN(w, r, filesize); err != nil {
				return "", err
			}
			if mustPad && filesize%2 == 1 {
				r.Read(pad)
			}
		}
	}
}
