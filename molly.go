package molly

import (
	"bufio"
	"bytes"
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

func (db Database) String() string {
	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)
	for _, c := range db.Classes {
		fmt.Fprintf(w, "%s\n", c)
	}
	w.Flush()
	return buf.String()
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

	for _, c := range db.Classes {
		env := at.NewEnvironmentFromFile(f, uint64(info.Size()), filename)
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
