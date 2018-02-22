package actions

import (
	"fmt"
	"os"

	"bitbucket.org/vahidi/molly/lib/actions/extractors"
	"bitbucket.org/vahidi/molly/lib/types"
)

var extractorList = map[string]func(*types.Env, string) (string, error){
	"zip":    extractors.Unzip,
	"gz":     extractors.Ungzip,
	"tar":    extractors.Untar,
	"cpio":   extractors.Uncpio,
	"mbrlba": extractors.MbrLba,
	"cramfs": extractors.Uncramfs,
}

var slicableExtractorList = map[string]func(*types.Env, string, ...uint64) (string, error){
	"":       extractors.BinarySlice,
	"binary": extractors.BinarySlice,
}

// RegisterExtractor provides a method to register user extractor functions
func RegisterExtractor(typ string, extoractor func(*types.Env, string) (string, error)) {
	extractorList[typ] = extoractor
}

// RegisterExtractorSlicable provides a method to register user extractor functions that can slice a file
func RegisterSlicableExtractor(typ string, extoractor func(*types.Env, string, ...uint64) (string, error)) {
	slicableExtractorList[typ] = extoractor
}

func extractFunction(e *types.Env, format string, name string, positions ...uint64) (string, error) {
	// parameter sanity check 1
	if len(positions)%2 != 0 {
		return "", fmt.Errorf("extract slice data is invalid")
	}

	// rewind so the extractors dont need to handle it
	if _, err := e.Reader.Seek(0, os.SEEK_SET); err != nil {
		return "", err
	}

	// these functions accept a slice:
	sf, found := slicableExtractorList[format]
	if found {
		// parameter sanity check 2
		if len(positions) == 0 {
			positions = []uint64{0, e.GetSize()}
		}
		return sf(e, name, positions...)
	}

	// these functions dont
	nf, found := extractorList[format]
	if found {
		// parameter sanity check 3
		if len(positions) != 0 {
			return "", fmt.Errorf("Unable to extract format '%s' as sliced", format)
		}
		return nf(e, name)
	}

	return "", fmt.Errorf("Unknown extraction format: '%s'", format)
}

func init() {
	types.FunctionRegister("extract", extractFunction)
}
