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

func TestMkPrintRecipeEmptyRecipe(t *testing.T) {
	// Cover mkPrintRecipe with empty recipe and quiet=false.
	// This path is effectively dead in normal execution (rules with empty recipes
	// are filtered before reaching dorecipe), but exercise it for coverage.
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	mkPrintRecipe("target", "", false)
	w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	buf.ReadFrom(r)
	got := buf.String()
	if want := "target: \n"; got != want {
		t.Errorf("mkPrintRecipe empty recipe: got %q, want %q", got, want)
	}
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

func TestMkNodeAlreadyClaimed(t *testing.T) {
	// Calling mkNode on a node that's already been claimed (status != Ready/Nop)
	// should return immediately without doing any work.
	u := &node{
		name:   "already-started",
		status: nodeStatusStarted,
	}
	g := &graph{nodes: map[string]*node{u.name: u}}
	mkNode(g, u, nil, nil, true, false)
	u.mutex.Lock()
	got := u.status
	u.mutex.Unlock()
	if got != nodeStatusStarted {
		t.Errorf("status changed to %v, expected it to remain nodeStatusStarted", got)
	}
}
