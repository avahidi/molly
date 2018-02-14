package analyzers

import (
	"debug/elf"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// turn a ReadSeeker into a ReaderAt :(
type readseekerat struct {
	r io.ReadSeeker
}

func (rsa readseekerat) ReadAt(p []byte, off int64) (int, error) {
	_, err := rsa.r.Seek(off, os.SEEK_SET)
	if err != nil {
		return 0, err
	}
	return rsa.r.Read(p)
}

// ElfAnalyzer examinies ELF binaries
func ElfAnalyzer(r io.ReadSeeker, w io.Writer, data ...interface{}) error {
	rsa := &readseekerat{r: r}
	file, err := elf.NewFile(rsa)
	if err != nil {
		return err
	}
	defer file.Close()

	// create report
	report := map[string]interface{}{
		"byte-order": file.ByteOrder.String(),
		"entry":      file.Entry,
		"machine":    file.Machine.String(),
		"OSABI":      file.OSABI.String(),
	}


	// extract data from elf structure

	// get imported symbols
	imported := make([]string, 0)
	if syms, err := file.ImportedSymbols(); err == nil {
		for _, s := range syms {
			imported = append(imported, fmt.Sprintf("%s:%s", s.Name, s.Library))
		}
	}
	report["imported"] = imported

	// get all used functions
	functions := make([]string, 0)
	syms1, _ := file.DynamicSymbols()
	syms2, _ := file.Symbols()
	syms := append(syms1, syms2...)
	for _, s := range syms {
		if typ := elf.ST_TYPE(s.Info); typ == elf.STT_FUNC {
			functions = append(functions, fmt.Sprintf("%s:%d:%d", s.Name, s.Value, s.Size))
		}
	}
	report["functions"] = functions

	if libs, err := file.ImportedLibraries(); err == nil {
		report["libraries"] = libs
	}

	// write report as json
	bs, err := json.MarshalIndent(report, "", "\t")
	if err != nil {
		return err
	}
	w.Write(bs)
	return nil
}
