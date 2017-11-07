package molly

import (
	"fmt"
	"os"
	"path/filepath"

	"bitbucket.org/vahidi/molly/at"
)

type Database struct {
	Classes []*at.Class
}

func NewDatabase() *Database {
	return &Database{}
}

func (db *Database) ScanFile(filename string) error {
	return filepath.Walk(filename, db.scanFile)
}

func (db *Database) scanFile(filename string, info os.FileInfo, err_prev error) error {
	if info.IsDir() {
		return err_prev
	}
	fmt.Printf("\nScaning file %s...\n", filename)
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	env := at.NewEnviorment(f)

	for _, c := range db.Classes {
		env.Reset()
		e, err := c.Eval(env)
		if e {
			fmt.Printf("Eval sucessful: %s\n", c.Id)
			env.Dump()
		}
		if err != nil {
			fmt.Printf("Eval failed: %s\n", err)
		}
	}

	return err_prev
}
