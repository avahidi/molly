package molly

import (
	"testing"

	"bitbucket.org/vahidi/molly/types"
)

// some helper functions to simplify the code
func loadRule(t *testing.T, text string, name string) (*types.Molly, *types.Rule) {
	molly := New()
	if err := LoadRulesFromText(molly, text); err != nil {
		t.Errorf("Could not load rule from text: %v", err)
		return nil, nil
	}
	rule, _ := molly.Rules.Flat[name]
	return molly, rule
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

	molly := New()
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
	text := "rule test (tag = \"testing\", pass = 2, bigendian = false)  { }"
	_, dut := loadRule(t, text, "test")
	if dut == nil {
		return
	}

	if len(dut.Actions)+len(dut.Conditions)+len(dut.Variables) != 0 {
		t.Errorf("Unknown actions/conditions/variables in empty rule")
	}

	mt := dut.Metadata
	if val, found := mt.GetBoolean("bigendian", true); val != false || !found {
		t.Errorf("Missing boolean metadata")
	}
	if val, found := mt.Get("pass", int64(0)); !(found && val == int64(2)) {
		t.Errorf("Missing number metadata: %t %v %v", val, found, mt)
	}
	if val, found := mt.GetString("tag", ""); val != "testing" || !found {
		t.Errorf("Missing string metadata")
	}
}
