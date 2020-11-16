package main

import (
	"testing"
)

// Test a mkfile with a single rule. The target has a single
// prerequesite; both are local files.
func TestParseOneRuleLocalFiles(t *testing.T) {
	mkfileAsString := "somefile.txt: a_prereq.csv\n\techo $target"
	ruleSet := parse(mkfileAsString, "mkfile", "/mkfile")
	if len(ruleSet.rules) != 1 {
		t.Errorf("There should be 1 rule")
	}
	rule := ruleSet.rules[0]
	if len(rule.prereqs) != 1 {
		t.Errorf("There should be 1 prerequesite")
	}
	if rule.prereqs[0] != "a_prereq.csv" {
		t.Errorf("The prerequesites are not a_prereq.csv")
	}

	noAttributes := attribSet{
		delFailed:       false,
		nonstop:         false,
		forcedTimestamp: false,
		nonvirtual:      false,
		quiet:           false,
		regex:           false,
		update:          false,
		virtual:         false,
		exclusive:       false,
	}
	if rule.attributes != noAttributes {
		t.Error("rule attributes are not all false", rule.attributes)
	}
}

// Test a mkfile with a single rule. The target has a two
// prerequesite; both are local files.
func TestParseOneRuleMultiPrereqLocalFiles(t *testing.T) {
	mkfileAsString := "somefile.txt: a_prereq.csv b_prereq.tsv\n\techo $target"
	ruleSet := parse(mkfileAsString, "mkfile", "/mkfile")
	if len(ruleSet.rules) != 1 {
		t.Errorf("There should be 1 rule")
	}
	rule := ruleSet.rules[0]
	if len(rule.prereqs) != 2 {
		t.Errorf("There should be 2 prerequesite")
	}
	if rule.prereqs[0] != "a_prereq.csv" || rule.prereqs[1] != "b_prereq.tsv" {
		t.Errorf("The prerequesites are not 'a_prereq.csv b_prereq.tsv'")
	}

	noAttributes := attribSet{
		delFailed:       false,
		nonstop:         false,
		forcedTimestamp: false,
		nonvirtual:      false,
		quiet:           false,
		regex:           false,
		update:          false,
		virtual:         false,
		exclusive:       false,
	}
	if rule.attributes != noAttributes {
		t.Error("rule attributes are not all false", rule.attributes)
	}
}
