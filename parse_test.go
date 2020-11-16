package main

import (
	"fmt"
	"reflect"
	"testing"
)

func ruleAttributesNotSet(t *testing.T, r *rule) {
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
	if r.attributes != noAttributes {
		t.Error("rule attributes are not all false", r.attributes)
	}
}

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

	ruleAttributesNotSet(t, &rule)
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

	ruleAttributesNotSet(t, &rule)
}

// Test a mkfile with a single rule. The target has a single
// prerequesite; both are local files. The rule has attributes
// set. Possible attributes are D, E, N, n, Q, R, U, V, X
func TestParseOneRuleWithAttributeLocalFiles(t *testing.T) {
	var attribMap = map[string]string{
		"D": "delFailed",
		"E": "nonstop",
		"n": "nonvirtual",
		"N": "forcedTimestamp",
		"Q": "quiet",
		"R": "regex",
		"U": "update",
		"V": "virtual",
		"X": "exclusive",
	}

	for k, v := range attribMap {
		mkfileAsString := fmt.Sprintf("somefile.txt:%s: a_prereq.csv\n\techo $target", k)
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
		//t.Log(rule.attributes)
		ps := reflect.ValueOf(rule.attributes)
		f := ps.FieldByName(v)

		if f.Bool() != true {
			t.Errorf("The %s attribute has not been set", k)
		}
	}
}
