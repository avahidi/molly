package lib

import (
	"testing"

	"bitbucket.org/vahidi/molly/lib/types"
)

// some helper functions to simplify the code
func loadRule(t *testing.T, text string, name string) (*types.Rule, *types.RuleSet) {
	rs, err := LoadRuleText(nil, text)
	if err != nil {
		t.Errorf("Could not load rule from text: %v", err)
		return nil, nil
	}
	rule, _ := rs.Flat[name]
	return rule, rs
}

func getVar(match *types.MatchEntry, name string) (interface{}, bool) {
	for match != nil {
		got, found := match.Vars[name]
		if found {
			return got, found
		}
		match = match.Parent
	}
	return nil, false
}
func matchCheck(t *testing.T, match *types.MatchEntry, name string, val interface{}) {
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

func TestLoadRuleSeries(t *testing.T) {
	var testdata = []struct {
		text       string
		countRules int
		countTop   int
	}{
		{"rule base { var a = Long(0); }", 1, 1},
		{"rule r00 : base { var b = Long(0); }", 1, 0},
		{"rule r01 : r00 { var c = Long(0); } rule r02 : r03 { } rule r03 { }", 3, 1},
	}
	rs := (*types.RuleSet)(nil)
	totalRules, totalTop := 0, 0

	for _, test := range testdata {
		var err error
		rs, err = LoadRuleText(rs, test.text)
		if err != nil {
			t.Errorf("Could not load rule from text: %v", err)
			return
		}
		totalRules += test.countRules
		if len(rs.Flat) != totalRules {
			t.Errorf("rule count mismatch")
		}

		totalTop += test.countTop
		if len(rs.Top) != totalTop {
			t.Errorf("top rule count mismatch")
		}
	}
}

func TestLoadRule(t *testing.T) {
	var testdata = []struct {
		text            string
		countVariables  int
		countConditions int
		countActions    int
	}{
		{
			`rule test {
				var v1 = 12;
				var v2 = v1 + 1;
			}`,
			2, 0, 0},
		{
			`rule test {
				var v1 = 12;
				if (v1 > 12) || true == false;
				system("echo hello world");
			}`,
			1, 1, 1},
	}

	for _, test := range testdata {
		dut, _ := loadRule(t, test.text, "test")
		if dut == nil {
			continue
		}
		if test.countVariables != len(dut.Variables) {
			t.Errorf("Wanted %d variables got %d", test.countVariables, len(dut.Variables))
		}
		if test.countConditions != len(dut.Conditions) {
			t.Errorf("Wanted %d conditions got %d", test.countConditions, len(dut.Conditions))
		}
		if test.countActions != len(dut.Actions) {
			t.Errorf("Wanted %d actions got %d", test.countActions, len(dut.Actions))
		}
	}
}

func TestLoadRuleMetadata(t *testing.T) {
	text := "rule test (name = \"joe\", age = 99, dead = false)  { }"
	dut, _ := loadRule(t, text, "test")
	if dut == nil {
		return
	}

	if len(dut.Actions)+len(dut.Conditions)+len(dut.Variables) != 0 {
		t.Errorf("Unknown actions/conditions/variables in empty rule")
	}

	mt := dut.Metadata
	if val, found := mt.GetBoolean("dead", true); val != false || !found {
		t.Errorf("Missing boolean metadata")
	}
	if val, found := mt.GetNumber("age", 0); val != 99 || !found {
		t.Errorf("Missing number metadata: %v %v %v", val, found, mt)
	}
	if val, found := mt.GetString("name", "nobody"); val != "joe" || !found {
		t.Errorf("Missing string metadata")
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
		_, rs := loadRule(t, test.rule, test.name)
		if rs == nil {
			continue
		}
		mr, err := ScanData(nil, rs, test.input)
		if err != nil || len(mr.MatchTree) != 1 {
			t.Errorf("No match in scan data")
			continue
		}

		// get the deepest match:
		match := mr.MatchTree[0]
		for len(match.Children) != 0 {
			match = match.Children[0]
		}
		matchCheck(t, match, "a", test.a)
		matchCheck(t, match, "b", test.b)
		matchCheck(t, match, "c", test.c)
	}
}