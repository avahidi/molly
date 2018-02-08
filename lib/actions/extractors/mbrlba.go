package extractors

import (
	"bitbucket.org/vahidi/molly/lib/types"
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

// TODO: this currently handles simplified LBA without looking at the extended partition

// MbrLba extracts a drive based on LBA parameters in the MBR
func MbrLba(e *types.Env, name string) (string, error) {
	var partition struct {
		State    uint8
		Begining [3]uint8
		Typ      uint8
		End      [3]uint8
		LbaStart uint32
		LbaSize  uint32
	}

	filesize := int64(e.GetSize())
	for i := 0; i < 4; i++ {
		if _, err := e.Reader.Seek(int64(0x1BE+i*16), os.SEEK_SET); err != nil {
			return "", err
		}
		if err := binary.Read(e.Reader, binary.LittleEndian, &partition); err != nil {
			return "", err
		}

		start := int64(partition.LbaStart) * 512
		end := start + int64(partition.LbaSize)*512
		if end > start && end <= filesize {
			w, err := e.Create(fmt.Sprintf("%s%d", name, i))
			if err != nil {
				return "", err
			}
			defer w.Close()

			if _, err := e.Reader.Seek(start, os.SEEK_SET); err != nil {
				return "", err
			}
			io.CopyN(w, e.Reader, end-start)

		}
	}
	return "", nil
}
