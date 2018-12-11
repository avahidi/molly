package lib

import (
	"testing"

	"bitbucket.org/vahidi/molly/report"
	"bitbucket.org/vahidi/molly/types"
)

func getVar(match *types.Match, name string) (interface{}, bool) {
	for match != nil {
		got, found := match.Vars[name]
		if found {
			return got, found
		}
		match = match.Parent
	}
	return nil, false
}
func matchCheck(t *testing.T, match *types.Match, name string, val interface{}) {
	if val == nil {
		return
	}

	got, found := getVar(match, name)
	if !found {
		t.Errorf("Missing match variable %s", name)
		return
	}
	if got != val {
		t.Errorf("Incorrect data for variable %s: wanted %v (%T) got %v (%T)",
			name, val, val, got, got)
	}
}

func TestScanData(t *testing.T) {
	var testdata = []struct {
		rule    string
		input   []byte
		name    string
		a, b, c interface{}
	}{
		{ // standard bigendian
			"rule test { var a = Byte(0); var b = Short(1); var c = Long(2); }",
			[]byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06},
			"test", int8(0x00), int16(0x0102), int32(0x02030405)},
		{ // little endian
			"rule test (bigendian = false) { var a = Byte(0); var b = Short(1); var c = Long(2); }",
			[]byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06},
			"test", int8(0x00), int16(0x0201), int32(0x05040302)},
		{ // little endian, one big
			"rule test (bigendian = false) { var a = Byte(0); var b = Short(1, bigendian = true); var c = Long(2); }",
			[]byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06},
			"test", int8(0x00), int16(0x0102), int32(0x05040302)},
		{ // two rules, both little endian
			"rule r1 (bigendian = false) { var a = Byte(0); var b = Short(1); } rule test : r1 {var c = Long(2); }",
			[]byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06},
			"test", int8(0x00), int16(0x0201), int32(0x05040302)},
		{ // two rules, different endians endian
			"rule r1 (bigendian = false) { var a = Byte(0); var b = Short(1); } rule test (bigendian = true) : r1 {var c = Long(2); }",
			[]byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06},
			"test", int8(0x00), int16(0x0201), int32(0x02030405)},
	}

	for _, test := range testdata {
		molly, _ := loadRule(t, test.rule, test.name)
		if molly == nil {
			continue
		}
		mr, err := ScanData(molly, test.input)
		if err != nil || len(mr.Files) != 1 {
			t.Errorf("No match in scan data")
			continue
		}
		// get the deepest match:
		match := mr.Files[0].Matches[0]
		for len(match.Children) != 0 {
			match = match.Children[0]
		}
		matchCheck(t, match, "a", test.a)
		matchCheck(t, match, "b", test.b)
		matchCheck(t, match, "c", test.c)
	}
}

func TestScanPass(t *testing.T) {
	ruletext := `
	rule p0 (pass = 0) { }
	rule p2 (pass = 2) { }
	rule p1 (pass = 1) { }
	`

	molly := New("", 0)
	if err := LoadRulesFromText(molly, ruletext); err != nil {
		t.Fatalf("Could not load rule from text: %v", err)
	}

	mr, err := ScanData(molly, []byte{})
	if err != nil {
		t.Fatal(err)
	}

	if len(mr.Files) != 1 || len(mr.Files[0].Matches) != 3 {
		t.Fatalf("Incorrect number of matches")
	}

	ms := mr.Files[0].Matches
	if ms[0].Rule.ID != "p0" || ms[1].Rule.ID != "p1" || ms[2].Rule.ID != "p2" {
		t.Fatalf("Rule pass not respected during scann")
	}
}

func TestScanNum(t *testing.T) {
	ruletext := `
	rule p0 (pass = 0) { var a = $num_matches; }
	rule p1 (pass = 1) { var b = $num_matches; }
	`

	molly := New("", 0)
	if err := LoadRulesFromText(molly, ruletext); err != nil {
		t.Fatalf("Could not load rule from text: %v", err)
	}

	mr, err := ScanData(molly, []byte{})
	if err != nil {
		t.Fatal(err)
	}

	if a, valid := report.FindInReportVarNumber(mr, "", "p0", "a"); !valid || a != 0 {
		t.Errorf("Num match failed (1)")
	}

	if b, valid := report.FindInReportVarNumber(mr, "", "p1", "b"); !valid || b != 1 {
		t.Errorf("Num match failed (2)")
	}
}

func TestScanHas(t *testing.T) {
	ruletext := `
	rule p0 (pass = 0) { var a = has("match", "p0"); }
	rule p1 (pass = 1) { var b = has("match", "p0"); }
	`

	molly := New("", 0)
	if err := LoadRulesFromText(molly, ruletext); err != nil {
		t.Fatalf("Could not load rule from text: %v", err)
	}

	mr, err := ScanData(molly, []byte{})
	if err != nil {
		t.Fatal(err)
	}

	if a, valid := report.FindInReportVarBool(mr, "", "p0", "a"); !valid || a {
		t.Errorf("has match failed (1)")
	}

	if b, valid := report.FindInReportVarBool(mr, "", "p1", "b"); !valid || !b {
		t.Errorf("has match failed (2)")
	}
}
