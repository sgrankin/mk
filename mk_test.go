package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sanity-io/litter"
)

type testvector struct {
	input  string
	output string
	errors string
	passes bool
}

// For each test vector, confirm that it matches
func TestBasicMaking(t *testing.T) {
	tests := []testvector{
		{
			// Really basic mk operation.
			input:  "testdata/test1.mk",
			output: "testdata/test1.mk.expected",
			errors: "",
			passes: true,
		},
		{
			// Environment variables are expanded in dependencies
			input:  "testdata/test2.mk",
			output: "testdata/test2.mk.expected",
			errors: "",
			passes: true,
		},
		{
			// Variables defined in the mkfile are expanded in dependencies
			input:  "testdata/test3.mk",
			output: "testdata/test3.mk.expected",
			errors: "",
			passes: true,
		},
		{
			// Pair of dependencies in an variable are expanded.
			input:  "testdata/test4.mk",
			output: "testdata/test4.mk.expected",
			errors: "",
			passes: true,
		},
		{
			// \ can escape newlines.
			input:  "testdata/test5.mk",
			output: "testdata/test5.mk.expected",
			errors: "",
			passes: true,
		},
		{
			// variables can be set included from another file
			input:  "testdata/test6.mk",
			output: "testdata/test6.mk.expected",
			errors: "",
			passes: true,
		},
		{
			// $vars are not expanded in import statements. expected to fail.
			input:  "testdata/test7.mk",
			output: "testdata/test7.mk.expected",
			errors: "",
			passes: true,
		},
		{
			// Variables expanded in recipes.
			input:  "testdata/test8.mk",
			output: "testdata/test8.mk.expected",
			errors: "",
			passes: true,
		},
		{
			// EOF can end a variable if no \n present.
			input:  "testdata/test9.mk",
			output: "testdata/test9.mk.expected",
			errors: "",
			passes: true,
		},
		{
			// External commands can generate variables
			input:  "testdata/test10.mk",
			output: "testdata/test10.mk.expected",
			errors: "",
			passes: true,
		},
		{
			// mkfile variables are expanded in backquote substitution
			input:  "testdata/test11.mk",
			output: "testdata/test11.mk.expected",
			errors: "",
			passes: true,
		},
		{
			// mkfile variables are expanded in backquote substitution
			input:  "testdata/test13.mk",
			output: "testdata/test13.mk.expected",
			errors: "",
			passes: false,
		},
	}

	for _, tv := range tests {
		// TODO(rjk): Validate generated errors.
		got, _, err := startMk("-n", "-f", tv.input)

		if err != nil {
			if !tv.passes {
				t.Logf("%s expected failure", tv.input)
				t.Logf("%s exec failed: %v", tv.input, err)
			} else {
				t.Errorf("%s exec failed: %v", tv.input, err)
			}
		}

		efd, err := os.Open(tv.output)
		if err != nil {
			t.Errorf("%s can't open: %v", tv.input, err)
			continue
		}
		want, err := ioutil.ReadAll(efd)
		if err != nil {
			t.Errorf("%s can't read: %v", tv.input, err)
			continue
		}

		// TODO(rjk): Read expected errors if they exist.
		if diff := cmp.Diff(string(want), string(got)); diff != "" {
			if !tv.passes {
				t.Logf("%s expected failure", tv.input)
				t.Logf("%s: mismatch (-want +got):\n%s", tv.input, diff)
			} else {
				t.Errorf("%s: mismatch (-want +got):\n%s", tv.input, diff)
			}
		}
	}
}

// Make sure that recipes get mk variables as environment.
func TestRecipesHaveEnv(t *testing.T) {
	input := "testdata/test12.mk"
	got, _, err := startMk("-f", input)

	if err != nil {
		t.Errorf("%s exec failed: %v", input, err)
	}

	// Make sure that the output has the right variables in it.
	// got should be the contents of the environment.
	envs := make([]string, 0)
	for _, b := range bytes.Split(got, []byte("\n")) {
		envs = append(envs, string(b))
	}
outer:
	for _, ekv := range []string{
		"bar=thebigness",
		"TEST_MAIN=mk",
		"shell=sh",
	} {
		for _, v := range envs {
			if v == ekv {
				continue outer
			}
		}
		t.Errorf("%s: output missing %s", input, ekv)
	}
}

func TestMain(m *testing.M) {
	switch os.Getenv("TEST_MAIN") {
	case "mk":
		main()
	default:
		e := m.Run()
		os.Exit(e)
	}
}

func startMk(args ...string) ([]byte, []byte, error) {
	outbuffy := new(bytes.Buffer)
	errbuffy := new(bytes.Buffer)

	mkcmd := exec.Command(os.Args[0], args...)
	mkcmd.Env = append(os.Environ(), "TEST_MAIN=mk")

	mkcmd.Stdout = outbuffy
	mkcmd.Stderr = errbuffy

	// log.Println("mkcmd", mkcmd)
	if err := mkcmd.Run(); err != nil {
		return nil, nil, err
	}

	return outbuffy.Bytes(), errbuffy.Bytes(), nil
}

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
		//		{
		//			input:       "\"$targetpath\"",
		//			vars:        map[string][]string{
		//				"targetpath": []string{"./testdata"},
		//			},
		//			expandticks: false,
		//			want:        []string{"\"$targetpath\""},
		//		},
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
