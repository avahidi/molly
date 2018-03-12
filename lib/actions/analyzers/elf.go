package analyzers

import (
	"bitbucket.org/vahidi/molly/lib/util"
	"debug/elf"
	"fmt"
	"io"
)

// ElfAnalyzer examinies ELF binaries
func ElfAnalyzer(r io.ReadSeeker,
	gen func(name string, typ string, data interface{}),
	data ...interface{}) error {

	rsa := util.NewReaderAt(r)
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
		if typ := elf.ST_TYPE(s.Info); typ == elf.STT_FUNC && s.Size > 0 {
			functions = append(functions, fmt.Sprintf("%s:%d:%d", s.Name, s.Value, s.Size))
		}
	}
	report["functions"] = functions

	if libs, err := file.ImportedLibraries(); err == nil && libs != nil {
		report["libraries"] = libs
	}

	gen("", "json", report)
	return nil
}
