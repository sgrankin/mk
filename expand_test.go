package main

import (
	"reflect"
	"testing"

	"github.com/sanity-io/litter"
)

type expandtv struct {
	input       string
	vars        map[string][]string
	expandticks bool
	want        []string
}

func TestExpand(t *testing.T) {
	tests := []expandtv{
		{
			input:       "a",
			vars:        map[string][]string{},
			expandticks: false,
			want:        []string{"a"},
		},
		{
			input: "a",
			vars: map[string][]string{
				"a": {"glenda"},
			},
			expandticks: false,
			want:        []string{"a"},
		},
		{
			input: "$a",
			vars: map[string][]string{
				"a": {"glenda"},
			},
			expandticks: false,
			want:        []string{"glenda"},
		},
		{
			input: "${a}",
			vars: map[string][]string{
				"a": {"glenda"},
			},
			expandticks: false,
			want:        []string{"glenda"},
		},
		{
			input: "${a}",
			vars: map[string][]string{
				"a": {"glenda", "gopher"},
			},
			expandticks: false,
			want:        []string{"glenda", "gopher"},
		},
		{
			input: "ab$targetpath",
			vars: map[string][]string{
				"targetpath": {"./testdata"},
			},
			expandticks: false,
			want:        []string{"ab./testdata"},
		},
		{
			input: "$targetpathab",
			vars: map[string][]string{
				"targetpath": {"./testdata"},
			},
			expandticks: false,
			want:        []string{"$targetpathab"},
		},
		{
			input: "$targetpath/foo",
			vars: map[string][]string{
				"targetpath": {"./testdata"},
			},
			expandticks: false,
			want:        []string{"./testdata/foo"},
		},
		{
			input: "${targetpath}/foo",
			vars: map[string][]string{
				"targetpath": {"./testdata"},
			},
			expandticks: false,
			want:        []string{"./testdata/foo"},
		},
		// This one differs between p9p mk and mk.
		// Do we want this difference?
		{
			input: "\"$targetpath\"",
			vars: map[string][]string{
				"targetpath": {"./testdata"},
			},
			expandticks: false,
			want:        []string{"./testdata"},
		},
		{
			input: "\"s3://$targetpath\"",
			vars: map[string][]string{
				"targetpath": {"testdata"},
			},
			expandticks: false,
			want:        []string{"s3://testdata"},
		},
		{
			input: "'$targetpath'",
			vars: map[string][]string{
				"targetpath": {"./testdata"},
			},
			expandticks: false,
			want:        []string{"$targetpath"},
		},
		{
			input: "$prefix.$suffix",
			vars: map[string][]string{
				"prefix": {"name"},
				"suffix": {"o"},
			},
			expandticks: false,
			want:        []string{"name.o"},
		},
		{
			input: "${targets:%=$targetpath/%}",
			vars: map[string][]string{
				"targetpath": {"./testdata"},
				"targets":    {"rupert", "ruxpin"},
			},
			expandticks: false,
			want:        []string{"./testdata/rupert", "./testdata/ruxpin"},
		},
		{
			input: "${targets:%=%.$novar}",
			vars: map[string][]string{
				"suffixes": {"o", "ab", "b"},
				"targets":  {"rupert", "ruxpin"},
			},
			expandticks: false,
			want:        []string{"rupert.$novar", "ruxpin.$novar"},
		},
		{
			input: "${targets:%=%.$suffixes}",
			vars: map[string][]string{
				"suffixes": {"teddy", "ab", "b"},
				"targets":  {"rupert", "ruxpin"},
			},
			expandticks: false,
			want: []string{
				"rupert.teddy",
				"ab",
				"b",
				"ruxpin.teddy",
				"ab",
				"b",
			},
		},
		{
			input: "${targets:%=%.$suffixes}",
			vars: map[string][]string{
				"suffixes": {"adventure"},
				"targets":  {"rupert bear", "ruxpin bear"},
			},
			expandticks: false,
			want: []string{
				"rupert bear.adventure",
				"ruxpin bear.adventure",
			},
		},
		// Backtick with expandBackticks=false → literal
		{
			input:       "`echo hello`",
			vars:        map[string][]string{},
			expandticks: false,
			want:        []string{"`echo hello`"},
		},
		// Escaped space
		{
			input:       `\ a`,
			vars:        map[string][]string{},
			expandticks: false,
			want:        []string{" a"},
		},
		// Escaped tab
		{
			input:       "\\\ta",
			vars:        map[string][]string{},
			expandticks: false,
			want:        []string{"\ta"},
		},
		// $$ → literal $
		{
			input:       "$$",
			vars:        map[string][]string{},
			expandticks: false,
			want:        []string{"$"},
		},
		// ${unclosed — missing }
		{
			input:       "${unclosed",
			vars:        map[string][]string{},
			expandticks: false,
			want:        []string{"${unclosed"},
		},
		// ${missing:%=.o} — namelist with nonexistent var
		{
			input:       "${missing:%=%.o}",
			vars:        map[string][]string{},
			expandticks: false,
			want:        []string{},
		},
		// $nonexistent — bare var not in vars or env
		{
			input:       "$nonexistent_var_xyz",
			vars:        map[string][]string{},
			expandticks: false,
			want:        []string{"$nonexistent_var_xyz"},
		},
		// Namelist with pattern that doesn't match all values (else branch)
		{
			input: "${targets:foo%=bar%}",
			vars: map[string][]string{
				"targets": {"fooX", "other"},
			},
			expandticks: false,
			want:        []string{"barX", "other"},
		},
	}

	//	failing := tests[11:]

	for i, tv := range tests {
		got := expand(tv.input, tv.vars, tv.expandticks)

		if !reflect.DeepEqual(got, tv.want) {
			t.Errorf("%d: input: %#v, vars: %s, ticks: %v. got %s, want %s",
				i,
				tv.input, litter.Sdump(tv.vars), tv.expandticks,
				litter.Sdump(got),
				litter.Sdump(tv.want))
		}
	}
}

func TestExpandRecipeSigils(t *testing.T) {
	tests := []expandtv{
		{
			input: "/runs/contition_${stem1}_bowtie_k10/mapping.bam.bai",
			vars: map[string][]string{
				"stem1": {"a"},
			},
			expandticks: false,
			want:        []string{"/runs/contition_a_bowtie_k10/mapping.bam.bai"},
		},
		{
			input: "s3://runs/contition_${stem1}_bowtie_k10/mapping.bam.bai",
			vars: map[string][]string{
				"stem1": {"a"},
			},
			expandticks: false,
			want:        []string{"s3://runs/contition_a_bowtie_k10/mapping.bam.bai"},
		},
		{
			input: "mkdir -p $target\necho $target",
			vars: map[string][]string{
				"target": {"a"},
			},
			expandticks: false,
			want:        []string{"mkdir -p a\necho a"},
		},
		{
			input: "mkdir -p $(dirname $target)\necho $target",
			vars: map[string][]string{
				"target": {"a"},
			},
			expandticks: false,
			want:        []string{"mkdir -p $(dirname a)\necho a"},
		},
		// backslash + non-$ char → preserved
		{
			input: "echo \\n",
			vars:  map[string][]string{},
			want:  []string{"echo \\n"},
		},
	}

	for i, tv := range tests {
		got := expandRecipeSigils(tv.input, tv.vars)

		if !reflect.DeepEqual(got, tv.want[0]) {
			t.Errorf("%d: input: %#v, vars: %s. got %s, want %s",
				i,
				tv.input, litter.Sdump(tv.vars),
				litter.Sdump(got),
				litter.Sdump(tv.want[0]))
		}
	}
}

func TestExpandDoubleQuotedEdgeCases(t *testing.T) {
	tests := []struct {
		input string
		want  string
		wantN int
	}{
		// Unterminated double quote
		{"abc", "abc", 3},
		// Backslash at end of input
		{"abc\\", "abc\\", 4},
		// Backslash-escaped char inside (not at end), terminated by "
		{"abc\\x\"", "abc\\x", 6},
	}

	for i, tt := range tests {
		got, n := expandDoubleQuoted(tt.input, map[string][]string{}, false)
		if got != tt.want || n != tt.wantN {
			t.Errorf("%d: expandDoubleQuoted(%q) = (%q, %d), want (%q, %d)",
				i, tt.input, got, n, tt.want, tt.wantN)
		}
	}
}

func TestExpandSingleQuotedEdgeCases(t *testing.T) {
	// Unterminated single quote
	got, n := expandSingleQuoted("abc")
	if got != "abc" || n != 3 {
		t.Errorf("expandSingleQuoted(%q) = (%q, %d), want (%q, %d)",
			"abc", got, n, "abc", 3)
	}
}

func TestExpandBackQuotedEdgeCases(t *testing.T) {
	// Unterminated sh-style backtick
	got, n := expandBackQuoted("cmd without closing", map[string][]string{
		"shell": {"sh"},
	})
	if len(got) != 1 || got[0] != "cmd without closing" || n != len("cmd without closing") {
		t.Errorf("expandBackQuoted(unterminated sh) = (%v, %d), want ([%q], %d)",
			got, n, "cmd without closing", len("cmd without closing"))
	}

	// Unterminated rc-style backtick: `{cmd with no closing brace
	input := "{cmd without closing"
	got, n = expandBackQuoted(input, map[string][]string{
		"shell": {"sh"},
	})
	if len(got) != 1 || got[0] != input || n != len(input) {
		t.Errorf("expandBackQuoted(unterminated rc) = (%v, %d), want ([%q], %d)",
			got, n, input, len(input))
	}
}

func TestExpandSuffixes(t *testing.T) {
	tests := []struct {
		input, stem, want string
	}{
		// Basic substitution
		{"%.o", "foo", "foo.o"},
		{"&.o", "foo", "foo.o"},
		// Escaped % → literal %
		{`\%`, "stem", "%"},
		{`\%.o`, "stem", "%.o"},
		// Escaped & → literal &
		{`\&`, "stem", "&"},
		// Multiple substitutions (exercises j>0 iterations)
		{"%.c.%", "foo", "foo.c.foo"},
		{"%.dir/%", "bar", "bar.dir/bar"},
		// Backslash followed by non-meta char → preserved literally
		{`\n`, "stem", `\n`},
		{`\..o`, "stem", `\..o`},
		{`a\.b`, "stem", `a\.b`},
		// Trailing backslash → preserved literally
		{`a\`, "stem", `a\`},
		{`\`, "stem", `\`},
		// Mix of escapes and substitutions
		{`\%=%`, "stem", "%=stem"},
		// No special chars
		{"plain", "stem", "plain"},
		// Empty input
		{"", "stem", ""},
	}

	for _, tt := range tests {
		got := expandSuffixes(tt.input, tt.stem)
		if got != tt.want {
			t.Errorf("expandSuffixes(%q, %q) = %q, want %q", tt.input, tt.stem, got, tt.want)
		}
	}
}

func TestExpandSigilEnvLookup(t *testing.T) {
	// Valid varname not in vars but in env → env lookup path (lines 213-215)
	t.Setenv("MK_TEST_ENV_VAR_XYZ", "envvalue")
	got := expand("$MK_TEST_ENV_VAR_XYZ", map[string][]string{}, false)
	if len(got) != 1 || got[0] != "envvalue" {
		t.Errorf("expand env var: got %v, want [envvalue]", got)
	}

	// Bracketed invalid varname in env → lines 221-223
	t.Setenv("1bad", "badval")
	got = expand("${1bad}", map[string][]string{}, false)
	if len(got) != 1 || got[0] != "badval" {
		t.Errorf("expand invalid varname in env: got %v, want [badval]", got)
	}

	// Bracketed invalid varname not in env → line 225
	got = expand("${2nonexistent_xyzzy}", map[string][]string{}, false)
	if len(got) != 1 || got[0] != "${2nonexistent_xyzzy}" {
		t.Errorf("expand invalid varname not in env: got %v, want [${2nonexistent_xyzzy}]", got)
	}
}

func TestExpandBackQuoted(t *testing.T) {
	tests := []expandtv{
		{
			input: "seq 1 5`",
			vars: map[string][]string{
				"shell": {"sh"},
			},
			expandticks: false,
			want:        []string{"1", "2", "3", "4", "5"},
		},
	}

	for i, tv := range tests {
		got, _ := expandBackQuoted(tv.input, tv.vars)

		if !reflect.DeepEqual(got, tv.want) {
			t.Errorf("%d: input: %#v, vars: %s. got %s, want %s",
				i,
				tv.input, litter.Sdump(tv.vars),
				litter.Sdump(got),
				litter.Sdump(tv.want))
		}
	}
}
