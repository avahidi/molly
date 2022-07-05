package analyzers

import (
	"testing"
)

func TestGitrefVersion(t *testing.T) {
	testdata := map[string]bool{
		"8173055926cdb8534fbaed517a792bd45aed8377":     true,
		"  8173055926cdb8534fbaed517a792bd45aed8377$$": true,  // start/end
		"8173055926cdb8534fbaed517a792bd45aeg8377":     false, // invalid G
		"8173055926cdb8534fbaed517a792bd45aeg837":      false, // one short
		"8173055926cdb8534fbaed517a792bd45aeg83777":    false, // one long
		"": false,
	}

	for str, ret := range testdata {
		got := stringIsGitref(str)
		if got != ret {
			t.Errorf("gitref detection error on %s", str)
		}
	}
}

func TestVersionVersion(t *testing.T) {
	testdata := map[string]bool{
		"version 1.2.3": true,
		"version 1.2b":  true,
		"v1.2.3":        true,
		"v1.2b":         true,
		"v10":           false,
		"v10.bb":        false,
		"127.0.0.1":     false,
		"8.8.8.8:53":    false,
		"10.0.0.0/24":   false,
		"":              false,
	}

	for str, ret := range testdata {
		got := stringIsVersion(str)
		if got != ret {
			t.Errorf("version detection error on %s", str)
		}
	}
}

func TestVersionCopyright(t *testing.T) {
	testdata := map[string]bool{
		"Crashware 10.7 Copyright evilcorp 2013-2022":               true,
		"Copyleft civilcorp 2000":                                   false,
		"COPYRIGHT (C) 2016 Free Software Foundation, Inc.":         true,
		"copyright (c) 2016 Free Software Foundation, Inc.":         true,
		"(c) 2077-2016 Free Stuff Company":                          true,
		"BusyBox is copyrighted by many authors between 1998-2012,": true,
		"Copyright (c) 2001-3 Shane Hyde and others":                true,
		"copyright.not": false,
		"(C)(D)":        false,
		"":              false,
	}

	for str, ret := range testdata {
		got := stringIsCopyright(str)
		if got != ret {
			t.Errorf("version detection error on %s", str)
		}
	}
}
func TestVersionFilenameVersion(t *testing.T) {
	testdata := map[string]bool{
		"":                        false,
		"bash":                    false,
		"gcc-3":                   false,
		"gcc-3-3":                 false,
		"libstuff-0.2.3.so":       true,
		"libstuff-0.2.3.a":        true,
		"libstuff_0.2x.a":         true,
		"libstuff_10.29.0-rc23.a": true,
	}
	for filename, ret := range testdata {
		got := filenameIsVersion(filename)
		if got != ret {
			t.Errorf("filenameversion detection error on %s", filename)
		}
	}
}
