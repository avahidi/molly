package extractors

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"

	"github.com/avahidi/molly/types"
)

// TODO: this currently handles simplified LBA without looking at the extended partition

// MbrLba extracts a drive based on LBA parameters in the MBR
func MbrLba(e *types.Env, name string) (string, error) {
	var partition struct {
		State     uint8
		Beginning [3]uint8
		Typ       uint8
		End       [3]uint8
		LbaStart  uint32
		LbaSize   uint32
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
		if end > start && start < filesize {
			filename := fmt.Sprintf("%s%d_%x_%x_%02x", name, i+1, start, end, partition.Typ)
			w, _, err := e.Create(filename)
			if err != nil {
				return "", err
			}
			defer w.Close()

			if _, err := e.Reader.Seek(start, os.SEEK_SET); err != nil {
				return "", err
			}
			// XXX: if some parts are missing, ignore them
			if end > filesize {
				end = filesize
			}
			io.CopyN(w, e.Reader, end-start)
		}
	}
	return "", nil
}
