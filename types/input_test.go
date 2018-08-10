package types

import (
	"testing"
	"time"
)

func TestInputGet(t *testing.T) {
	tid := time.Now()
	i1 := NewInput(nil, "/dir1/dir2/filename.c", 1023, tid)
	i2 := NewInput(i1, "some.file.go", 555, tid)
	i3 := NewInput(i2, "new file", 0, tid)

	var testdata = []struct {
		target *Input
		name   string
		data   interface{}
	}{
		{i1, "filename", "/dir1/dir2/filename.c"},
		{i1, "filesize", int64(1023)},
		{i1, "shortname", "filename.c"},
		{i1, "dirname", "/dir1/dir2/"},
		{i1, "ext", ".c"},
		{i1, "basename", "filename"},
		{i1, "depth", 0},
		{i1, "parent", ""},

		{i2, "filename", "some.file.go"},
		{i2, "filesize", int64(555)},
		{i2, "shortname", "some.file.go"},
		{i2, "dirname", ""},
		{i2, "ext", ".go"},
		{i2, "basename", "some.file"},
		{i2, "depth", 1},
		{i2, "parent", i1.Filename},

		{i3, "filename", "new file"},
		{i3, "shortname", "new file"},
		{i3, "dirname", ""},
		{i3, "ext", ""},
		{i3, "parent", i2.Filename},
		{i3, "time", tid},
	}

	for _, test := range testdata {
		data, found := test.target.Get(test.name)
		if !found {
			t.Errorf("Input: could not Get('%s') ", test.name)
		} else if data != test.data {
			t.Errorf("Input Get('%s') = '%v', expected '%v'",
				test.name, data, test.data)
		}
	}
}
