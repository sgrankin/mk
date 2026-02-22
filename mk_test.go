package main

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

type testvector struct {
	input  string
	output string
	errors string
	passes bool
	args   []string // extra arguments to pass to mk
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
		{
			// Rules can be created by pipeing commands
			input:  "testdata/test14.mk",
			output: "testdata/test14.mk.expected",
			errors: "",
			passes: true,
		},
		{
			// Test alternative recipe shell
			input:  "testdata/test15.mk",
			output: "testdata/test15.mk.expected",
			errors: "",
			passes: true,
		},
		{
			// Test alternative recipe shell
			input:  "testdata/test16.mk",
			output: "testdata/test16.mk.expected",
			errors: "",
			passes: true,
		},
		{
			// Test
			input:  "testdata/test17.mk",
			output: "testdata/test17.mk.expected",
			errors: "",
			passes: true,
		},
		{
			// Test multiline backtick expansion; -p 1 for deterministic order
			input:  "testdata/test18.mk",
			output: "testdata/test18.mk.expected",
			errors: "",
			passes: true,
			args:   []string{"-p", "1"},
		},
		{
			// Suffix/percent metarules
			input:  "testdata/test19.mk",
			output: "testdata/test19.mk.expected",
			errors: "",
			passes: true,
			args:   []string{"-p", "1"},
		},
		{
			// $$ and \$ recipe escapes
			input:  "testdata/test20.mk",
			output: "testdata/test20.mk.expected",
			errors: "",
			passes: true,
		},
		{
			// Multi-target rules
			input:  "testdata/test21.mk",
			output: "testdata/test21.mk.expected",
			errors: "",
			passes: true,
			args:   []string{"-p", "1"},
		},
		{
			// Quiet Q attribute
			input:  "testdata/test22.mk",
			output: "testdata/test22.mk.expected",
			errors: "",
			passes: true,
		},
		{
			// "Nothing to mk" — only meta-rules present
			input:  "testdata/test23.mk",
			output: "testdata/test23.mk.expected",
			errors: "",
			passes: true,
		},
		{
			// -q quiet flag
			input:  "testdata/test24.mk",
			output: "testdata/test24.mk.expected",
			errors: "",
			passes: true,
			args:   []string{"-q", "-p", "1"},
		},
		{
			// -a rebuild-all flag
			input:  "testdata/test25.mk",
			output: "testdata/test25.mk.expected",
			errors: "",
			passes: true,
			args:   []string{"-a", "-p", "1"},
		},
		{
			// -r shallow-rebuild flag
			input:  "testdata/test26.mk",
			output: "testdata/test26.mk.expected",
			errors: "",
			passes: true,
			args:   []string{"-r", "-p", "1", "dep"},
		},
		{
			// Concrete rule overrides meta-rule
			input:  "testdata/test27.mk",
			output: "testdata/test27.mk.expected",
			errors: "",
			passes: true,
			args:   []string{"-p", "1"},
		},
		{
			// shell variable in mkfile
			input:  "testdata/test28.mk",
			output: "testdata/test28.mk.expected",
			errors: "",
			passes: true,
		},
		{
			// Suffix rule with prefix (foo%.o)
			input:  "testdata/test29.mk",
			output: "testdata/test29.mk.expected",
			errors: "",
			passes: true,
			args:   []string{"-p", "1"},
		},
		{
			// Assignment with non-word tokens (colon) in value
			input:  "testdata/test30.mk",
			output: "testdata/test30.mk.expected",
			errors: "",
			passes: true,
		},
		{
			// Adjacent quotes in bare word
			input:  "testdata/test31.mk",
			output: "testdata/test31.mk.expected",
			errors: "",
			passes: true,
		},
		{
			// -F flag (don't drop shell arguments)
			input:  "testdata/test33.mk",
			output: "testdata/test33.mk.expected",
			errors: "",
			passes: true,
			args:   []string{"-F"},
		},
		{
			// Regex rule with prereqs
			input:  "testdata/test34.mk",
			output: "testdata/test34.mk.expected",
			errors: "",
			passes: true,
			args:   []string{"-p", "1"},
		},
		{
			// Single quotes in bare word (lexer coverage)
			input:  "testdata/test35.mk",
			output: "testdata/test35.mk.expected",
			errors: "",
			passes: true,
		},
		{
			// Pipe include with = token in command (parser coverage)
			input:  "testdata/test36.mk",
			output: "testdata/test36.mk.expected",
			errors: "",
			passes: true,
		},
	}

	for _, tv := range tests {
		t.Run(tv.input, func(t *testing.T) {
			t.Parallel()
			// TODO(rjk): Validate generated errors.
			args := append([]string{"-n", "-f", tv.input}, tv.args...)
			got, _, err := startMk(args...)
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
				t.Fatalf("%s can't open: %v", tv.input, err)
			}
			want, err := io.ReadAll(io.Reader(efd))
			if err != nil {
				t.Fatalf("%s can't read: %v", tv.input, err)
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
		})
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

func TestDotOutput(t *testing.T) {
	t.Parallel()
	got, _, err := startMk("-dot", "-f", "testdata/test5.mk")
	if err != nil {
		t.Fatalf("exec failed: %v", err)
	}

	output := string(got)

	// Check structure: header, footer, and expected edges.
	if !strings.HasPrefix(output, "digraph mk {\n") {
		t.Errorf("missing digraph header, got: %s", output)
	}
	if !strings.HasSuffix(output, "}\n") {
		t.Errorf("missing digraph footer, got: %s", output)
	}

	wantEdges := []string{
		`"" -> "test3.mk.o";`,
		`"test3.mk.o" -> "one";`,
		`"test3.mk.o" -> "two";`,
	}
	for _, edge := range wantEdges {
		if !strings.Contains(output, edge) {
			t.Errorf("missing edge %q in output:\n%s", edge, output)
		}
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
	return startMkWithStdin("", args...)
}

func startMkWithStdin(stdin string, args ...string) ([]byte, []byte, error) {
	outbuffy := new(bytes.Buffer)
	errbuffy := new(bytes.Buffer)

	mkcmd := exec.Command(os.Args[0], args...)
	mkcmd.Env = append(os.Environ(), "TEST_MAIN=mk")

	mkcmd.Stdin = strings.NewReader(stdin)

	mkcmd.Stdout = outbuffy
	mkcmd.Stderr = errbuffy

	// log.Println("mkcmd", mkcmd)
	err := mkcmd.Run()
	return outbuffy.Bytes(), errbuffy.Bytes(), err
}

func TestInteractiveMode(t *testing.T) {
	t.Parallel()
	// Leading whitespace covers the whitespace-skip path in the interactive loop
	got, _, err := startMkWithStdin(" y\n", "-i", "-n", "-p", "1", "-f", "testdata/test32.mk")
	if err != nil {
		t.Fatalf("exec failed: %v", err)
	}

	output := string(got)
	want := "dep: build dep\nProceed? dep: build dep\n"
	if output != want {
		t.Errorf("mismatch:\n  got:  %q\n  want: %q", output, want)
	}
}

func TestInteractiveModeDecline(t *testing.T) {
	t.Parallel()
	got, _, err := startMkWithStdin("n\n", "-i", "-n", "-p", "1", "-f", "testdata/test32.mk")
	if err != nil {
		t.Fatalf("exec failed: %v", err)
	}

	output := string(got)
	want := "dep: build dep\nProceed? "
	if output != want {
		t.Errorf("mismatch:\n  got:  %q\n  want: %q", output, want)
	}
}

func TestInteractiveModeEOF(t *testing.T) {
	t.Parallel()
	// Empty stdin → EOF on ReadRune → returns without building
	got, _, _ := startMkWithStdin("", "-i", "-n", "-p", "1", "-f", "testdata/test32.mk")

	output := string(got)
	want := "dep: build dep\nProceed? "
	if output != want {
		t.Errorf("mismatch:\n  got:  %q\n  want: %q", output, want)
	}
}

func TestErrorCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		args        []string
		errContains string
	}{
		{
			name:        "cycle detection",
			args:        []string{"-n", "-f", "testdata/test_err_cycle.mk"},
			errContains: "cycle",
		},
		{
			name:        "missing target",
			args:        []string{"-n", "-f", "testdata/test_err_missing.mk"},
			errContains: "don't know how to make",
		},
		{
			name:        "bad -C directory",
			args:        []string{"-C", "/nonexistent_xyz", "-n", "-f", "testdata/test1.mk"},
			errContains: "changing directory",
		},
		{
			name:        "missing mkfile",
			args:        []string{"-n", "-f", "testdata/nonexistent.mk"},
			errContains: "no mkfile found",
		},
		{
			name:        "parse syntax error",
			args:        []string{"-n", "-f", "testdata/test_err_syntax.mk"},
			errContains: "syntax error",
		},
		{
			name:        "invalid attribute",
			args:        []string{"-n", "-f", "testdata/test_err_invalid_attr.mk"},
			errContains: "attribute",
		},
		{
			name:        "invalid varname assignment",
			args:        []string{"-n", "-f", "testdata/test_err_invalid_varname.mk"},
			errContains: "not a valid variable name",
		},
		{
			name:        "include nonexistent file",
			args:        []string{"-n", "-f", "testdata/test_err_include_nonexistent.mk"},
			errContains: "cannot open",
		},
		{
			name:        "empty pipe include",
			args:        []string{"-n", "-f", "testdata/test_err_empty_pipe.mk"},
			errContains: "empty pipe include",
		},
		{
			name:        "failed pipe include",
			args:        []string{"-n", "-f", "testdata/test_err_failed_pipe.mk"},
			errContains: "failed",
		},
		{
			name:        "unterminated quote",
			args:        []string{"-n", "-f", "testdata/test_err_unterminated_quote.mk"},
			errContains: "end of file",
		},
		{
			name:        "ambiguous rules",
			args:        []string{"-n", "-f", "testdata/test_err_ambiguous.mk"},
			errContains: "ambiguous",
		},
		{
			name:        "regex rule bad pattern",
			args:        []string{"-n", "-f", "testdata/test_err_bad_regex.mk"},
			errContains: "regular expression",
		},
		{
			name:        "unexpected token in target",
			args:        []string{"-n", "-f", "testdata/test_err_unexpected_token.mk"},
			errContains: "syntax error",
		},
		{
			name:        "targets without colon",
			args:        []string{"-n", "-f", "testdata/test_err_targets_no_colon.mk"},
			errContains: "targets",
		},
		{
			name:        "bad token in attributes",
			args:        []string{"-n", "-f", "testdata/test_err_attr_bad_token.mk"},
			errContains: "attributes",
		},
		{
			name:        "bad token in prerequisites",
			args:        []string{"-n", "-f", "testdata/test_err_prereq_bad_token.mk"},
			errContains: "prerequisites",
		},
		{
			name:        "bad token in redirect include",
			args:        []string{"-n", "-f", "testdata/test_err_redir_bad_token.mk"},
			errContains: "parsing include",
		},
		{
			name:        "multi-value include filename",
			args:        []string{"-n", "-f", "testdata/test_err_multivalue_include.mk"},
			errContains: "single value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, stderr, err := startMk(tt.args...)
			if err == nil {
				t.Fatalf("expected error but got none")
			}
			got := string(stderr)
			if !strings.Contains(got, tt.errContains) {
				t.Errorf("stderr = %q, want substring %q", got, tt.errContains)
			}
		})
	}
}
