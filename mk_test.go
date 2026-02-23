package main

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestMain(m *testing.M) {
	switch os.Getenv("TEST_MAIN") {
	case "mk":
		main()
	default:
		e := m.Run()
		os.Exit(e)
	}
}

func startMkWithStdin(stdin string, args ...string) ([]byte, []byte, error) {
	outbuffy := new(bytes.Buffer)
	errbuffy := new(bytes.Buffer)

	mkcmd := exec.Command(os.Args[0], args...)
	mkcmd.Env = append(os.Environ(), "TEST_MAIN=mk")

	mkcmd.Stdin = strings.NewReader(stdin)

	mkcmd.Stdout = outbuffy
	mkcmd.Stderr = errbuffy

	err := mkcmd.Run()
	return outbuffy.Bytes(), errbuffy.Bytes(), err
}

func TestInteractiveMode(t *testing.T) {
	t.Parallel()
	// Leading whitespace covers the whitespace-skip path in the interactive loop
	got, _, err := startMkWithStdin(" y\n", "-I", "-n", "-p", "1", "-f", "testdata/interactive.mk")
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
	got, _, err := startMkWithStdin("n\n", "-I", "-n", "-p", "1", "-f", "testdata/interactive.mk")
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
	got, _, _ := startMkWithStdin("", "-I", "-n", "-p", "1", "-f", "testdata/interactive.mk")

	output := string(got)
	want := "dep: build dep\nProceed? "
	if output != want {
		t.Errorf("mismatch:\n  got:  %q\n  want: %q", output, want)
	}
}
