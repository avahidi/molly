package util

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"io"
	"os"
)

// ReadUntil reads from stream until it sees byte "until" or
// gathers "maxsize" bytes (ignored if zero or less)
func ReadUntil(r io.Reader, until byte, maxsize int) ([]byte, bool, error) {
	var bw bytes.Buffer
	br := bufio.NewReaderSize(r, 1)
	for i := 0; i < maxsize; i++ {
		c, err := br.ReadByte()
		if err != nil {
			return nil, false, err
		}
		if c == until {
			return bw.Bytes(), true, nil
		}
		bw.WriteByte(c)
	}
	return bw.Bytes(), false, nil
}

type readeratwrapper struct {
	r io.ReadSeeker
}

func (rsa readeratwrapper) ReadAt(p []byte, off int64) (int, error) {
	_, err := rsa.r.Seek(off, os.SEEK_SET)
	if err != nil {
		return 0, err
	}
	return rsa.r.Read(p)
}

// NewReaderAt turns a ReadSeeker into a ReaderAt for the rare cases its needed
func NewReaderAt(r io.ReadSeeker) io.ReaderAt {
	return &readeratwrapper{r: r}
}

// BufreaderAt is a helper for creating a buffered reader at a position
func BufreaderAt(r io.ReadSeeker, offset int64) (*bufio.Reader, error) {
	if _, err := r.Seek(offset, os.SEEK_SET); err != nil {
		return nil, err
	}
	return bufio.NewReader(r), nil
}

// Structured define a file that has a stream and a byteorder
type Structured struct {
	Reader io.ReadSeeker
	Order  binary.ByteOrder
}

// ReadAt reads structred data from a given offset
func (bf Structured) ReadAt(offset int64, data interface{}) error {
	if _, err := bf.Reader.Seek(offset, os.SEEK_SET); err != nil {
		return err
	}
	return bf.Read(data)
}

// Read reads structred data from the stream
func (bf Structured) Read(data interface{}) error {
	return binary.Read(bf.Reader, bf.Order, data)
}

// Process will fall p on each byte until it returns false
func Process(r io.ByteReader, p func(b uint8, n int) bool) (int, error) {
	for i := 0; ; i++ {
		b, err := r.ReadByte()
		if err != nil {
			return i, err
		}
		if !p(b, i) {
			return i, nil
		}
	}
}
