package main

import (
	"regexp"
	"testing"
)

func TestMatchMetaRule(t *testing.T) {
	regex, err := regexp.Compile(`^data/processed/(\d+)/mapping_k10.bam.bai$`)
	if err != nil {
		t.Error("Failure to compile regex pattern")
	}
	pat := pattern{false, "data/processed/(\\d+)/mapping_k10.bam.bai", regex}
	matches := pat.match("data/processed/12345/mapping_k10.bam.bai")
	want := []string{"data/processed/12345/mapping_k10.bam.bai", "12345"}
	if matches[0] != want[0] || matches[1] != want[1] {
		t.Error("failed to match regular expression")
	}
}

func TestMatchSimplePattern(t *testing.T) {
	// Simple string pattern matching (no regex)
	pat := pattern{spat: "foo"}

	// Exact match returns empty slice
	matches := pat.match("foo")
	if matches == nil {
		t.Error("expected match for 'foo', got nil")
	}
	if len(matches) != 0 {
		t.Errorf("expected empty slice for exact match, got %v", matches)
	}

	// Non-match returns nil
	matches = pat.match("bar")
	if matches != nil {
		t.Errorf("expected nil for non-match, got %v", matches)
	}
}

func TestEquivRecipe(t *testing.T) {
	r1 := &rule{recipe: "build", shell: []string{"sh"}}
	r2 := &rule{recipe: "build", shell: []string{"sh"}}
	r3 := &rule{recipe: "other", shell: []string{"sh"}}
	r4 := &rule{recipe: "build", shell: []string{"bash"}}
	r5 := &rule{recipe: "build", shell: []string{"sh", "-e"}}

	if !r1.equivRecipe(r2) {
		t.Error("identical recipes should be equivalent")
	}
	if r1.equivRecipe(r3) {
		t.Error("different recipes should not be equivalent")
	}
	if r1.equivRecipe(r4) {
		t.Error("different shells should not be equivalent")
	}
	if r1.equivRecipe(r5) {
		t.Error("different shell arg counts should not be equivalent")
	}
}

func TestExecuteAssignment(t *testing.T) {
	// Valid assignment with non-word tokens in value
	rs := &ruleSet{
		vars:        map[string][]string{},
		rules:       []rule{},
		targetrules: map[string][]int{},
	}
	tokens := []token{
		{typ: tokenWord, val: "x"},
		{typ: tokenWord, val: "a"},
		{typ: tokenColon, val: ":"},
		{typ: tokenWord, val: "b"},
	}
	err := rs.executeAssignment(tokens)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rs.vars["x"]) != 1 || rs.vars["x"][0] != "a:b" {
		t.Errorf("expected x=[a:b], got %v", rs.vars["x"])
	}

	// Assignment value starting with non-word token (hits len(input)==0 branch)
	nonWordStart := []token{
		{typ: tokenWord, val: "y"},
		{typ: tokenColon, val: ":"},
		{typ: tokenWord, val: "val"},
	}
	err = rs.executeAssignment(nonWordStart)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rs.vars["y"]) != 1 || rs.vars["y"][0] != ":val" {
		t.Errorf("expected y=[:val], got %v", rs.vars["y"])
	}

	// Invalid variable name
	badName := []token{
		{typ: tokenWord, val: "1bad"},
	}
	err = rs.executeAssignment(badName)
	if err == nil {
		t.Error("expected error for invalid variable name")
	}
}

func TestIsValidVarName(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{"foo", true},
		{"_bar", true},
		{"a1", true},
		{"1bad", false},
		{"a-b", false},
	}
	for _, tt := range tests {
		got := isValidVarName(tt.name)
		if got != tt.want {
			t.Errorf("isValidVarName(%q) = %v, want %v", tt.name, got, tt.want)
		}
	}
}
