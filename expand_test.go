package main

import (
	"reflect"
	"testing"

	"github.com/sanity-io/litter"
)

type expandtv struct {
	name        string
	input       string
	vars        map[string][]string
	expandticks bool
	want        []string
}

func TestExpand(t *testing.T) {
	tests := []expandtv{
		{
			name:        "literal",
			input:       "a",
			vars:        map[string][]string{},
			expandticks: false,
			want:        []string{"a"},
		},
		{
			name: "literal_same_as_var",
			input: "a",
			vars: map[string][]string{
				"a": {"glenda"},
			},
			expandticks: false,
			want:        []string{"a"},
		},
		{
			name: "bare_var",
			input: "$a",
			vars: map[string][]string{
				"a": {"glenda"},
			},
			expandticks: false,
			want:        []string{"glenda"},
		},
		{
			name: "braced_var",
			input: "${a}",
			vars: map[string][]string{
				"a": {"glenda"},
			},
			expandticks: false,
			want:        []string{"glenda"},
		},
		{
			name: "braced_var_multi",
			input: "${a}",
			vars: map[string][]string{
				"a": {"glenda", "gopher"},
			},
			expandticks: false,
			want:        []string{"glenda", "gopher"},
		},
		{
			name: "prefix_var",
			input: "ab$targetpath",
			vars: map[string][]string{
				"targetpath": {"./testdata"},
			},
			expandticks: false,
			want:        []string{"ab./testdata"},
		},
		{
			name: "var_suffix_mismatch",
			input: "$targetpathab",
			vars: map[string][]string{
				"targetpath": {"./testdata"},
			},
			expandticks: false,
			want:        []string{"$targetpathab"},
		},
		{
			name: "var_slash_suffix",
			input: "$targetpath/foo",
			vars: map[string][]string{
				"targetpath": {"./testdata"},
			},
			expandticks: false,
			want:        []string{"./testdata/foo"},
		},
		{
			name: "braced_var_slash_suffix",
			input: "${targetpath}/foo",
			vars: map[string][]string{
				"targetpath": {"./testdata"},
			},
			expandticks: false,
			want:        []string{"./testdata/foo"},
		},
		{
			// Differs from p9p mk: double quotes strip here.
			name: "double_quoted_var",
			input: "\"$targetpath\"",
			vars: map[string][]string{
				"targetpath": {"./testdata"},
			},
			expandticks: false,
			want:        []string{"./testdata"},
		},
		{
			name: "double_quoted_prefix_var",
			input: "\"s3://$targetpath\"",
			vars: map[string][]string{
				"targetpath": {"testdata"},
			},
			expandticks: false,
			want:        []string{"s3://testdata"},
		},
		{
			name: "single_quoted_var",
			input: "'$targetpath'",
			vars: map[string][]string{
				"targetpath": {"./testdata"},
			},
			expandticks: false,
			want:        []string{"$targetpath"},
		},
		{
			name: "multi_var_dot",
			input: "$prefix.$suffix",
			vars: map[string][]string{
				"prefix": {"name"},
				"suffix": {"o"},
			},
			expandticks: false,
			want:        []string{"name.o"},
		},
		{
			name: "namelist_subst",
			input: "${targets:%=$targetpath/%}",
			vars: map[string][]string{
				"targetpath": {"./testdata"},
				"targets":    {"rupert", "ruxpin"},
			},
			expandticks: false,
			want:        []string{"./testdata/rupert", "./testdata/ruxpin"},
		},
		{
			name: "namelist_missing_var",
			input: "${targets:%=%.$novar}",
			vars: map[string][]string{
				"suffixes": {"o", "ab", "b"},
				"targets":  {"rupert", "ruxpin"},
			},
			expandticks: false,
			want:        []string{"rupert.$novar", "ruxpin.$novar"},
		},
		{
			name: "namelist_multi_value",
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
			name: "namelist_single_suffix",
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
		{
			name:        "backtick_literal",
			input:       "`echo hello`",
			vars:        map[string][]string{},
			expandticks: false,
			want:        []string{"`echo hello`"},
		},
		{
			name:        "backtick_with_prefix",
			input:       "prefix`cmd`suffix",
			vars:        map[string][]string{},
			expandticks: false,
			want:        []string{"prefix`cmd`suffix"},
		},
		{
			name:        "escaped_space",
			input:       `\ a`,
			vars:        map[string][]string{},
			expandticks: false,
			want:        []string{" a"},
		},
		{
			name:        "escaped_tab",
			input:       "\\\ta",
			vars:        map[string][]string{},
			expandticks: false,
			want:        []string{"\ta"},
		},
		{
			name:        "dollar_dollar",
			input:       "$$",
			vars:        map[string][]string{},
			expandticks: false,
			want:        []string{"$"},
		},
		{
			name:        "unclosed_brace",
			input:       "${unclosed",
			vars:        map[string][]string{},
			expandticks: false,
			want:        []string{"${unclosed"},
		},
		{
			name:        "namelist_nonexistent_var",
			input:       "${missing:%=%.o}",
			vars:        map[string][]string{},
			expandticks: false,
			want:        []string{},
		},
		{
			name:        "bare_nonexistent_var",
			input:       "$nonexistent_var_xyz",
			vars:        map[string][]string{},
			expandticks: false,
			want:        []string{"$nonexistent_var_xyz"},
		},
		{
			name: "namelist_partial_match",
			input: "${targets:foo%=bar%}",
			vars: map[string][]string{
				"targets": {"fooX", "other"},
			},
			expandticks: false,
			want:        []string{"barX", "other"},
		},
	}

	for _, tv := range tests {
		t.Run(tv.name, func(t *testing.T) {
			got := expand(tv.input, tv.vars, tv.expandticks)
			if !reflect.DeepEqual(got, tv.want) {
				t.Errorf("input: %#v, vars: %s, ticks: %v\n  got  %s\n  want %s",
					tv.input, litter.Sdump(tv.vars), tv.expandticks,
					litter.Sdump(got),
					litter.Sdump(tv.want))
			}
		})
	}
}

func TestExpandRecipeSigils(t *testing.T) {
	tests := []struct {
		name  string
		input string
		vars  map[string][]string
		want  string
	}{
		{
			name:  "braced_var",
			input: "/runs/contition_${stem1}_bowtie_k10/mapping.bam.bai",
			vars:  map[string][]string{"stem1": {"a"}},
			want:  "/runs/contition_a_bowtie_k10/mapping.bam.bai",
		},
		{
			name:  "s3_prefix",
			input: "s3://runs/contition_${stem1}_bowtie_k10/mapping.bam.bai",
			vars:  map[string][]string{"stem1": {"a"}},
			want:  "s3://runs/contition_a_bowtie_k10/mapping.bam.bai",
		},
		{
			name:  "multiline",
			input: "mkdir -p $target\necho $target",
			vars:  map[string][]string{"target": {"a"}},
			want:  "mkdir -p a\necho a",
		},
		{
			name:  "subshell",
			input: "mkdir -p $(dirname $target)\necho $target",
			vars:  map[string][]string{"target": {"a"}},
			want:  "mkdir -p $(dirname a)\necho a",
		},
		{
			name:  "trailing_backslash",
			input: "echo \\",
			vars:  map[string][]string{},
			want:  "echo \\",
		},
		{
			name:  "backslash_non_dollar",
			input: "echo \\n",
			vars:  map[string][]string{},
			want:  "echo \\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := expandRecipeSigils(tt.input, tt.vars)
			if got != tt.want {
				t.Errorf("expandRecipeSigils(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestExpandDoubleQuotedEdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
		wantN int
	}{
		{"unterminated", "abc", "abc", 3},
		{"backslash_at_end", "abc\\", "abc\\", 4},
		{"backslash_escaped_char", "abc\\x\"", "abc\\x", 6},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, n := expandDoubleQuoted(tt.input, map[string][]string{}, false)
			if got != tt.want || n != tt.wantN {
				t.Errorf("expandDoubleQuoted(%q) = (%q, %d), want (%q, %d)",
					tt.input, got, n, tt.want, tt.wantN)
			}
		})
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
	tests := []struct {
		name  string
		input string
		want  []string
		wantN int
	}{
		{
			name:  "unterminated_sh",
			input: "cmd without closing",
			want:  []string{"cmd without closing"},
			wantN: len("cmd without closing"),
		},
		{
			name:  "unterminated_rc",
			input: "{cmd without closing",
			want:  []string{"{cmd without closing"},
			wantN: len("{cmd without closing"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, n := expandBackQuoted(tt.input, map[string][]string{"shell": {"sh"}})
			if !reflect.DeepEqual(got, tt.want) || n != tt.wantN {
				t.Errorf("expandBackQuoted(%q) = (%v, %d), want (%v, %d)",
					tt.input, got, n, tt.want, tt.wantN)
			}
		})
	}
}

func TestExpandSuffixes(t *testing.T) {
	tests := []struct {
		name              string
		input, stem, want string
	}{
		{"percent", "%.o", "foo", "foo.o"},
		{"ampersand", "&.o", "foo", "foo.o"},
		{"escaped_percent", `\%`, "stem", "%"},
		{"escaped_percent_suffix", `\%.o`, "stem", "%.o"},
		{"escaped_ampersand", `\&`, "stem", "&"},
		{"multi_percent", "%.c.%", "foo", "foo.c.foo"},
		{"multi_percent_slash", "%.dir/%", "bar", "bar.dir/bar"},
		{"backslash_n", `\n`, "stem", `\n`},
		{"backslash_dot", `\..o`, "stem", `\..o`},
		{"backslash_mid", `a\.b`, "stem", `a\.b`},
		{"trailing_backslash", `a\`, "stem", `a\`},
		{"only_backslash", `\`, "stem", `\`},
		{"escape_and_subst", `\%=%`, "stem", "%=stem"},
		{"plain", "plain", "stem", "plain"},
		{"empty", "", "stem", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := expandSuffixes(tt.input, tt.stem)
			if got != tt.want {
				t.Errorf("expandSuffixes(%q, %q) = %q, want %q", tt.input, tt.stem, got, tt.want)
			}
		})
	}
}

func TestExpandSigilEnvLookup(t *testing.T) {
	// Valid varname not in vars but in env → env lookup path
	t.Setenv("MK_TEST_ENV_VAR_XYZ", "envvalue")
	got := expand("$MK_TEST_ENV_VAR_XYZ", map[string][]string{}, false)
	if len(got) != 1 || got[0] != "envvalue" {
		t.Errorf("expand env var: got %v, want [envvalue]", got)
	}

	// Bracketed invalid varname in env
	t.Setenv("1bad", "badval")
	got = expand("${1bad}", map[string][]string{}, false)
	if len(got) != 1 || got[0] != "badval" {
		t.Errorf("expand invalid varname in env: got %v, want [badval]", got)
	}

	// Bracketed invalid varname not in env
	got = expand("${2nonexistent_xyzzy}", map[string][]string{}, false)
	if len(got) != 1 || got[0] != "${2nonexistent_xyzzy}" {
		t.Errorf("expand invalid varname not in env: got %v, want [${2nonexistent_xyzzy}]", got)
	}
}

func TestExpandBackQuoted(t *testing.T) {
	tests := []struct {
		name  string
		input string
		vars  map[string][]string
		want  []string
	}{
		{
			name:  "printf",
			input: "printf 'a b c'`",
			vars:  map[string][]string{"shell": {"sh"}},
			want:  []string{"a", "b", "c"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := expandBackQuoted(tt.input, tt.vars)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("expandBackQuoted(%q)\n  got  %s\n  want %s",
					tt.input, litter.Sdump(got), litter.Sdump(tt.want))
			}
		})
	}
}
