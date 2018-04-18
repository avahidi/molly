package actions

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"fmt"
	"hash"
	"hash/crc32"
	"hash/crc64"
	"io"
	"os"

	"bitbucket.org/vahidi/molly/lib/types"
)

var hashlist = map[string]func() hash.Hash{
	// hash
	"sha256": sha256.New,
	"sha128": sha1.New,
	"sha1":   sha1.New,
	"md5":    md5.New,

	// CRC
	"crc32":            func() hash.Hash { return crc32.NewIEEE() },
	"crc32-ieee":       func() hash.Hash { return crc32.New(crc32.MakeTable(crc32.IEEE)) },
	"crc32-castagnoli": func() hash.Hash { return crc32.New(crc32.MakeTable(crc32.Castagnoli)) },
	"crc32-koopman":    func() hash.Hash { return crc32.New(crc32.MakeTable(crc32.Koopman)) },
	"crc64":            func() hash.Hash { return crc64.New(crc64.MakeTable(crc64.ISO)) },
	"crc64-iso":        func() hash.Hash { return crc64.New(crc64.MakeTable(crc64.ISO)) },
	"crc64-ecma":       func() hash.Hash { return crc64.New(crc64.MakeTable(crc64.ECMA)) },
}

// ChecksumHelp prints checksum help text
func ChecksumHelp() {
	fmt.Println("Valid checksum types are: ")
	for key, _ := range hashlist {
		fmt.Println("\t", key)
	}
}

// RegisterChecksumFunction provides a method to register user checksum function
func RegisterChecksumFunction(typ string, generator func() hash.Hash) {
	hashlist[typ] = generator
}

// checksumFunction computes checksum over a number of slices of the current file
func checksumFunction(e *types.Env, typ string, positions ...uint64) ([]byte, error) {
	if len(positions)%2 != 0 {
		return nil, fmt.Errorf("Wrong number of parameters in checksum()")
	}
	if len(positions) == 0 {
		positions = []uint64{0, e.GetSize()}
	}

	hnew, found := hashlist[typ]
	if !found {
		ChecksumHelp()
		return nil, fmt.Errorf("Unknown checksum function: '%s'", typ)
	}

	total := e.GetSize()
	hash := hnew()
	hash.Reset()
	for i := 0; i < len(positions); i += 2 {
		start, end := positions[i], positions[i+1]
		if start < 0 || start >= end || end > total {
			return nil, fmt.Errorf("invalid boundaries in checksum(): %d-%v", start, end)
		}
		if _, err := e.Input.Seek(int64(start), os.SEEK_SET); err != nil {
			return nil, err
		}
		if _, err := io.CopyN(hash, e.Input, int64(end-start)); err != nil {
			return nil, err
		}
	}
	return hash.Sum(nil), nil
}

func init() {
	ActionRegister("checksum", checksumFunction)
}
