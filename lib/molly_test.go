package lib

import (
	"testing"

	"bitbucket.org/vahidi/molly/lib/types"
)

// some helper functions to simplify the code
func loadRule(t *testing.T, text string, name string) (*types.Molly, *types.Rule) {
	molly := New("", "", 0)
	if err := LoadRulesFromText(molly, text); err != nil {
		t.Errorf("Could not load rule from text: %v", err)
		return nil, nil
	}
	rule, _ := molly.Rules.Flat[name]
	return molly, rule
}

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
	totalRules, totalTop := 0, 0

	molly := New("", "", 0)
	for _, test := range testdata {
		var err error
		if err = LoadRulesFromText(molly, test.text); err != nil {
			t.Errorf("Could not load rule from text: %v", err)
			return
		}
		totalRules += test.countRules
		if len(molly.Rules.Flat) != totalRules {
			t.Errorf("rule count mismatch")
		}

		totalTop += test.countTop
		if len(molly.Rules.Top) != totalTop {
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
		_, dut := loadRule(t, test.text, "test")
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
	_, dut := loadRule(t, text, "test")
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

// BenchmarkMollyExpr test effect of lazy evaluation and  early termination
func BenchmarkMollyExpr(b *testing.B) {
	input := []byte{0, 1, 2, 3, 4, 5, 6}
	var testdata = []struct {
		rule  string
		match bool
	}{
		{"rule a { var a = false; var b = Long(0); if a && b > 0; }",
			false,
		},
		{"rule a { var a = true; var b = Long(0); if a && b > 0; }",
			true,
		},
		{`rule a {
			var a = false;
			var b0 = Byte(0);
			var b1 = Byte(1);
			var b2 = Byte(2);
			var b3 = Byte(3);
			var b4 = Byte(4);
			if a && b0 == 0 && b1 == 1 && b2 == 2 && b3 == 3 && b4 == b4;
			}`,
			false,
		},
		{`rule a {
			var a = true;
			var b0 = Byte(0);
			var b1 = Byte(1);
			var b2 = Byte(2);
			var b3 = Byte(3);
			var b4 = Byte(4);
			if a && b0 == 0 && b1 == 1 && b2 == 2 && b3 == 3 && b4 == b4;
			}`,
			true,
		},
	}

	for _, test := range testdata {
		molly := New("", "", 0)
		if err := LoadRulesFromText(molly, test.rule); err != nil {
			b.Errorf("Could not load benchmark rule: %v", err)
		}

		// we want the scan time to dominate, not rule load time, hence this loop
		for n := 0; n < 100000; n++ {
			mr, err := ScanData(molly, input)
			if err != nil || mr == nil { // mr == nil to avoid compiler optimization
				b.Errorf("Could not scan benchmark data: %v", err)
			}
			if (len(mr.Files) != 0) != test.match {
				b.Errorf("Unexpected result in benchmark scan")
			}
		}
	}

}
