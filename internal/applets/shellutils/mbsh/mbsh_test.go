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

func TestRunExecutesExternalCommand(t *testing.T) {
	out, errOut, err := run(t, "echo hello\nexit\n")
	if err != nil {
		t.Fatalf("Run error = %v (stderr=%q)", err, errOut)
	}
	if !strings.Contains(out, "hello") {
		t.Errorf("stdout = %q, want it to contain %q", out, "hello")
	}
}

func TestPromptShowsCwd(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	out, _, runErr := run(t, "exit\n")
	if runErr != nil {
		t.Fatalf("Run error = %v", runErr)
	}
	if !strings.Contains(out, "mbsh:"+cwd+"> ") {
		t.Errorf("prompt = %q, want it to contain cwd %q", out, cwd)
	}
}

func TestRunEOFEndsLoop(t *testing.T) {
	out, errOut, err := run(t, "echo bye\n")
	if err != nil {
		t.Fatalf("Run error = %v (stderr=%q)", err, errOut)
	}
	if !strings.Contains(out, "bye") {
		t.Errorf("stdout = %q, want it to contain %q", out, "bye")
	}
}

// TestRunFinalLineWithoutNewline exercises the EOF path that still runs a final
// line lacking a trailing newline.
func TestRunFinalLineWithoutNewline(t *testing.T) {
	out, errOut, err := run(t, "echo tail")
	if err != nil {
		t.Fatalf("Run error = %v (stderr=%q)", err, errOut)
	}
	if !strings.Contains(out, "tail") {
		t.Errorf("stdout = %q, want it to contain %q", out, "tail")
	}
}

// TestRunExitOnFinalLineWithoutNewline covers the EOF branch where the trailing
// line is the exit command.
func TestRunExitOnFinalLineWithoutNewline(t *testing.T) {
	_, _, err := run(t, "exit")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
}

func TestRunEmptyInputEndsImmediately(t *testing.T) {
	out, errOut, err := run(t, "")
	if err != nil {
		t.Fatalf("Run error = %v (stderr=%q)", err, errOut)
	}
	if !strings.HasPrefix(out, "mbsh:") || !strings.HasSuffix(out, "> ") {
		t.Errorf("stdout = %q, want a single mbsh prompt", out)
	}
}

func TestCommentLineIgnored(t *testing.T) {
	out, errOut, err := run(t, "# this is a comment\necho ok\nexit\n")
	if err != nil {
		t.Fatalf("Run error = %v (stderr=%q)", err, errOut)
	}
	if !strings.Contains(out, "ok") {
		t.Errorf("stdout = %q", out)
	}
}

func TestLastStatusExpansion(t *testing.T) {
	// "false" exits 1; "$?" must then expand to 1.
	out, errOut, err := run(t, "false\necho $?\nexit\n")
	if err != nil {
		t.Fatalf("Run error = %v (stderr=%q)", err, errOut)
	}
	if !strings.Contains(out, "1") {
		t.Errorf("stdout = %q, want it to contain exit status 1", out)
	}
}

func TestCdBuiltinChangesDirectory(t *testing.T) {
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })

	dir := t.TempDir()
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

func TestCdNoArgGoesHome(t *testing.T) {
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })

	home := t.TempDir()
	t.Setenv("HOME", home)
	wantHome, err := filepath.EvalSymlinks(home)
	if err != nil {
		t.Fatal(err)
	}

	if _, errOut, runErr := run(t, "cd\nexit\n"); runErr != nil {
		t.Fatalf("Run error = %v (stderr=%q)", runErr, errOut)
	}
	got, _ := os.Getwd()
	got, _ = filepath.EvalSymlinks(got)
	if got != wantHome {
		t.Errorf("cd with no arg -> %q, want HOME %q", got, wantHome)
	}
}

func TestCdDashReturnsToPrevious(t *testing.T) {
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })

	start, _ := filepath.EvalSymlinks(orig)
	dir := t.TempDir()

	// cd into dir, then "cd -" must return to the starting directory.
	if _, errOut, runErr := run(t, "cd "+dir+"\ncd -\nexit\n"); runErr != nil {
		t.Fatalf("Run error = %v (stderr=%q)", runErr, errOut)
	}
	got, _ := os.Getwd()
	got, _ = filepath.EvalSymlinks(got)
	if got != start {
		t.Errorf("cd - landed in %q, want %q", got, start)
	}
}

func TestExitAndQuit(t *testing.T) {
	for _, word := range []string{"exit", "quit"} {
		word := word
		t.Run(word, func(t *testing.T) {
			out, _, err := run(t, "echo first\n"+word+"\necho second\n")
			if err != nil {
				t.Fatalf("Run error = %v", err)
			}
			if !strings.Contains(out, "first") {
				t.Errorf("out = %q, want 'first'", out)
			}
			if strings.Contains(out, "second") {
				t.Errorf("%q should have stopped the loop, out = %q", word, out)
			}
		})
	}
}

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
	if strings.Contains(out.String(), "mbsh:") {
		t.Errorf("--help should not print a prompt: %q", out.String())
	}
}
