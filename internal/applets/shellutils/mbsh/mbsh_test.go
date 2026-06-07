package mbsh_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/shellutils/mbsh"
	"github.com/nao1215/mimixbox/internal/command"
)

// run drives the REPL with script as its input and returns stdout, stderr and
// the Run error.
func run(t *testing.T, script string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	stdio := command.IO{In: strings.NewReader(script), Out: out, Err: errBuf}
	err := mbsh.New().Run(context.Background(), stdio, nil)
	return out.String(), errBuf.String(), err
}

// TestRunExecutesExternalCommand feeds the shell a one-line script that runs the
// external "echo" command, then exits. The command output must reach stdout.
func TestRunExecutesExternalCommand(t *testing.T) {
	out, errOut, err := run(t, "echo hello\nexit\n")
	if err != nil {
		t.Fatalf("Run error = %v (stderr=%q)", err, errOut)
	}
	if !strings.Contains(out, "hello") {
		t.Errorf("stdout = %q, want it to contain %q", out, "hello")
	}
	// Two prompts are printed: one before "echo hello" and one before "exit".
	if got := strings.Count(out, "> "); got != 2 {
		t.Errorf("prompt count = %d, want 2 (out=%q)", got, out)
	}
}

// TestRunEOFEndsLoop feeds a script that never says "exit"; the loop must end
// cleanly when stdio.In reaches EOF.
func TestRunEOFEndsLoop(t *testing.T) {
	out, errOut, err := run(t, "echo bye\n")
	if err != nil {
		t.Fatalf("Run error = %v (stderr=%q)", err, errOut)
	}
	if !strings.Contains(out, "bye") {
		t.Errorf("stdout = %q, want it to contain %q", out, "bye")
	}
}

// TestRunEmptyInputEndsImmediately verifies that immediate EOF (empty script)
// returns without hanging or error.
func TestRunEmptyInputEndsImmediately(t *testing.T) {
	out, errOut, err := run(t, "")
	if err != nil {
		t.Fatalf("Run error = %v (stderr=%q)", err, errOut)
	}
	if out != "> " {
		t.Errorf("stdout = %q, want a single prompt", out)
	}
}

// TestCdBuiltinChangesDirectory verifies that the cd built-in changes the
// process working directory through the REPL.
func TestCdBuiltinChangesDirectory(t *testing.T) {
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })

	dir := t.TempDir()
	// Resolve symlinks (macOS /tmp, etc.) so the comparison is exact.
	want, err := filepath.EvalSymlinks(dir)
	if err != nil {
		t.Fatal(err)
	}

	_, errOut, runErr := run(t, "cd "+dir+"\nexit\n")
	if runErr != nil {
		t.Fatalf("Run error = %v (stderr=%q)", runErr, errOut)
	}

	got, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	got, err = filepath.EvalSymlinks(got)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Errorf("working directory = %q, want %q", got, want)
	}
}

// TestCdWithoutPathReportsError verifies that "cd" with no argument writes the
// path-required error to stderr without stopping the loop.
func TestCdWithoutPathReportsError(t *testing.T) {
	_, errOut, err := run(t, "cd\nexit\n")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if !strings.Contains(errOut, "path required") {
		t.Errorf("stderr = %q, want it to mention %q", errOut, "path required")
	}
}

// TestHelp verifies the standard --help flag is wired through the flag set.
// --help is an argument (not REPL input), so it must short-circuit before any
// prompt is printed.
func TestHelp(t *testing.T) {
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	stdio := command.IO{In: strings.NewReader(""), Out: out, Err: errBuf}
	if err := mbsh.New().Run(context.Background(), stdio, []string{"--help"}); err != nil {
		t.Fatalf("--help error = %v", err)
	}
	if !strings.Contains(out.String(), "Usage: mbsh") {
		t.Errorf("--help out = %q", out.String())
	}
	if strings.Contains(out.String(), prompt) {
		t.Errorf("--help should not print a prompt: %q", out.String())
	}
}

// prompt is duplicated from the package under test for the assertion above.
const prompt = "> "
