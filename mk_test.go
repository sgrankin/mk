package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"testing"

	"github.com/google/go-cmp/cmp"
)

type testvector struct {
	input  string
	output string
	errors string
}

// For each test vector, confirm that it matches
func TestBasicMaking(t *testing.T) {
	tests := []testvector{
		{
			input:  "testdata/test1.mk",
			output: "testdata/test1.mk.expected",
			errors: "",
		},
// failure expected. 
//		{
//			input:  "testdata/test2.mk",
//			output: "testdata/test2.mk.expected",
//			errors: "",
//		},
		{
			input:  "testdata/test3.mk",
			output: "testdata/test3.mk.expected",
			errors: "",
		},
		{
			input:  "testdata/test4.mk",
			output: "testdata/test4.mk.expected",
			errors: "",
		},
// \ doesn't work. expected failure
//		{
//			input:  "testdata/test5.mk",
//			output: "testdata/test5.mk.expected",
//			errors: "",
//		},
		{
			input:  "testdata/test6.mk",
			output: "testdata/test6.mk.expected",
			errors: "",
		},
// $vars are not expanded in import statements. expected to fail.
//		{
//			input:  "testdata/test7.mk",
//			output: "testdata/test7.mk.expected",
//			errors: "",
//		},
// $vars are not expanded in import statements. expected to fail.
//		{
//			input:  "testdata/test7.mk",
//			output: "testdata/test7.mk.expected",
//			errors: "",
//		},
		{
			// Variables expanded in recipes.
			input:  "testdata/test8.mk",
			output: "testdata/test8.mk.expected",
			errors: "",
		},
// expected to fail.
//		{
//			// EOF can end a variable if no \n present.
//			input:  "testdata/test9.mk",
//			output: "testdata/test9.mk.expected",
//			errors: "",
//		},
// expected to fail: <| doesn't work.
//		{
//			// External commands can generate 
//			input:  "testdata/test10.mk",
//			output: "testdata/test10.mk.expected",
//			errors: "",
//		},
	}

	for _, tv := range tests {
		// TODO(rjk): Validate generated errors.
		got, _, err := startMk("-n", "-f", tv.input)

		if err != nil {
			t.Errorf("%s exec failed: %v", tv.input, err)
			continue
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
			t.Errorf("%s: mismatch (-want +got):\n%s", tv.input, diff)
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
