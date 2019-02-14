package molly

import (
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"

	_ "bitbucket.org/vahidi/molly/operators" // import default actions
	"bitbucket.org/vahidi/molly/types"
	"bitbucket.org/vahidi/molly/util"
)

// suggestBaseName picks a new base name for a file
func suggestBaseName(c *types.Configuration, input *types.FileData) string {
	for i := 0; ; i++ {
		basename := filepath.Join(c.OutDir, util.SanitizeFilename(input.Filename))
		if i != 0 {
			basename = fmt.Sprintf("%s_%04d", basename, i)
		}
		if util.GetPathType(basename) == util.NoFile && util.GetPathType(basename+"_") == util.NoFile {
			return basename
		}
	}
}

// checkDuplicate if a file is duplicate and does all the book keeping
func checkDuplicate(m *types.Molly, file *types.FileData) (bool, error) {
	hash, err := util.HashFile(file.Filename)
	if err != nil {
		return false, err
	}

	hashtxt := hex.EncodeToString(hash)

	// seen it has been done, add the checksum to our analysis
	file.RegisterAnalysis("checksum", hashtxt, nil)

	// check if we have already seen this checksum:
	org, alreadyseen := m.FilesByHash[hashtxt]
	if !alreadyseen {
		m.FilesByHash[hashtxt] = file
		return false, nil
	}

	// record it
	file.DuplicateOf = org
	file.RegisterWarning("duplicate of %s", org.Filename)
	return true, nil
}

// scanFile opens and scans a single file
func scanFile(m *types.Molly, env *types.Env, filename_ string, parent *types.FileData) {

	fl := &util.FileList{}
	fl.Push(filename_)
	for {
		filename, fi, err := fl.Pop()
		if filename == "" {
			return
		}
		fr, found := m.Files[filename]
		if !found {
			fr = types.NewFileData(filename, parent)
			m.Files[filename] = fr

			// update basename to something we can use to create files from
			if fr.Parent == nil {
				fr.FilenameOut = suggestBaseName(m.Config, fr)

				// make sure its path is there and we have a soft link to the real file
				path, _ := filepath.Split(fr.FilenameOut)
				util.SafeMkdir(path)

				// make sure we link to the absolute path
				filename_abs, _ := filepath.Abs(fr.Filename)
				os.Symlink(filename_abs, fr.FilenameOut)
			}
		}

		// if we for some reason have done this one before, just skip it
		if fr.Processed {
			continue
		}
		fr.Processed = true

		// started with an error, no point moving on
		if err != nil {
			fr.RegisterError(err)
			continue
		}

		// record what we know about it so far
		fr.SetTime(fi.ModTime())
		fr.Filesize = fi.Size()

		if m.Config.MaxDepth != 0 && fr.Depth >= m.Config.MaxDepth {
			fr.RegisterErrorf("File depth above %d", m.Config.MaxDepth)
			continue
		}

		alreadyseen, err := checkDuplicate(m, fr)
		if err != nil {
			fr.RegisterError(err)
		}
		if alreadyseen {
			// if we already have this guy, just delete it (assuming its ours)
			// XXX: this does not follow the extraction hierarchy
			if fr.Parent != nil {
				os.Remove(fr.FilenameOut)
				os.Symlink(fr.DuplicateOf.Filename, fr.FilenameOut)
			}
			continue
		}

		reader, err := os.Open(filename)
		if err != nil {
			fr.RegisterError(err)
			continue
		}

		scanInput(m, env, reader, fr)

		// manual Close insted of defer Close, or we will have too many files open
		reader.Close()

		// now that the file is closed, attempt to adjust its time
		if t := fr.GetTime(); t != fi.ModTime() {
			os.Chtimes(fr.FilenameOut, t, t)
		}
	}
}

// ScanFiles scans a set of files for matches.
func ScanFiles(m *types.Molly, files ...string) error {
	env := types.NewEnv(m)
	for _, filename := range files {
		scanFile(m, env, filename, nil)
	}
	return nil
}
