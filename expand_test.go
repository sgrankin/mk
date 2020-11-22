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
		//This one differs between p9p mk and mk.
		//Do we want this difference?
		{
			input: "\"$targetpath\"",
			vars: map[string][]string{
				"targetpath": []string{"./testdata"},
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
