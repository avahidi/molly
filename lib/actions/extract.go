package actions

import (
	"fmt"

	"bitbucket.org/vahidi/molly/lib/actions/extractors"
	"bitbucket.org/vahidi/molly/lib/types"
	"bitbucket.org/vahidi/molly/lib/util"
)

// extractor holds both kind of extractor because right now we can not
// combine Seek and Multi readers
type extractor struct {
	full  func(*types.Env, string) (string, error)
	slice func(*types.Env, string, ...uint64) (string, error)
}

var extractorList = map[string]extractor{
	"":       extractor{slice: extractors.BinarySlice},
	"binary": extractor{slice: extractors.BinarySlice},
	"zip":    extractor{full: extractors.Unzip},
	"gz":     extractor{full: extractors.Ungzip},
	"tar":    extractor{full: extractors.Untar},
	"cpio":   extractor{full: extractors.Uncpio},
	"mbrlba": extractor{full: extractors.MbrLba},
	"cramfs": extractor{full: extractors.Uncramfs},
	"jffs2":  extractor{full: extractors.Unjffs2},
}

// ExtractorRegister provides a method to register user extractor functions
func ExtractorRegister(typ string, e func(*types.Env, string) (string, error)) {
	extractorList[typ] = extractor{full: e}
}

// ExtractorSliceRegister provides a method to register user extractor functions that can slice a file
func ExtractorSliceRegister(typ string, e func(*types.Env, string, ...uint64) (string, error)) {
	extractorList[typ] = extractor{slice: e}
}

func ExtractorHelp() {
	fmt.Println("Available extractors:")
	for name, ex := range extractorList {
		if ex.full != nil {
			fmt.Printf("\t%-12s (full file)\n", name)
		} else {
			fmt.Printf("\t%-12s (slice)\n", name)
		}
	}
}

func extractFunction(e *types.Env, format string, name string, positions ...uint64) (string, error) {
	ex, found := extractorList[format]
	if !found {
		ExtractorHelp()
		util.RegisterFatalf("Unknown extractor: '%s'", format)
		return "", fmt.Errorf("Unknown extractor: '%s'", format)
	}

	if ex.slice != nil {
		// parameter sanity checks
		if len(positions)%2 != 0 {
			return "", fmt.Errorf("extract slice data is invalid")
		}
		if len(positions) == 0 {
			positions = []uint64{0, e.GetSize()}
		}
		return ex.slice(e, name, positions...)
	}

	// parameter sanity checks
	if len(positions) != 0 {
		return "", fmt.Errorf("extractor '%s' does not accept a slice", format)
	}
	return ex.full(e, name)
}

func init() {
	ActionRegister("extract", extractFunction)
}
