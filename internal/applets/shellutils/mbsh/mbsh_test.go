package mbsh_test

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/shellutils/mbsh"
	"github.com/nao1215/mimixbox/internal/command"
)

// run drives the REPL with script delivered through an os.Pipe, so stdio.In is
// a real *os.File and a command launched by the shell shares the same file
// descriptor — the configuration a piped script has in production. (A strings
// reader would not work here: os/exec read-aheads a non-*os.File stdin, racily
// consuming the script before a non-reading command exits.) It returns stdout,
// stderr and the Run error.
func run(t *testing.T, script string) (string, string, error) {
	t.Helper()
	pr, pw, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		_, _ = pw.WriteString(script)
		_ = pw.Close()
	}()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	stdio := command.IO{In: pr, Out: out, Err: errBuf}
	rerr := mbsh.New().Run(context.Background(), stdio, nil)
	_ = pr.Close()
	return out.String(), errBuf.String(), rerr
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
	if !strings.Contains(out.String(), "Exit status:") {
		t.Errorf("--help out missing exit status section = %q", out.String())
	}
}

func TestNameSynopsis(t *testing.T) {
	c := mbsh.New()
	if c.Name() != "mbsh" {
		t.Errorf("Name() = %q, want %q", c.Name(), "mbsh")
	}
	if c.Synopsis() == "" {
		t.Error("Synopsis() is empty")
	}
}

// TestTokenizeErrorReported drives execInput's tokenize-error branch with an
// unterminated quote, which must report on stderr and continue the loop.
func TestTokenizeErrorReported(t *testing.T) {
	out, errOut, err := run(t, "echo 'unterminated\necho recovered\nexit\n")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if !strings.Contains(errOut, "mbsh:") {
		t.Errorf("stderr = %q, want a tokenize error", errOut)
	}
	// The shell kept running after the bad line.
	if !strings.Contains(out, "recovered") {
		t.Errorf("stdout = %q, want it to contain 'recovered'", out)
	}
}

// TestSyntaxErrorReported drives execInput's parse-error branch with a pipeline
// that has no command after the pipe operator.
func TestSyntaxErrorReported(t *testing.T) {
	out, errOut, err := run(t, "echo hi |\necho recovered\nexit\n")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if !strings.Contains(errOut, "mbsh:") {
		t.Errorf("stderr = %q, want a syntax error", errOut)
	}
	if !strings.Contains(out, "recovered") {
		t.Errorf("stdout = %q, want recovery after syntax error", out)
	}
}

// TestStatusTwoAfterSyntaxError verifies that a shell-level syntax error sets
// $? to 2, observable on the following line.
func TestStatusTwoAfterSyntaxError(t *testing.T) {
	requireCmd(t, "echo")
	out, _, err := run(t, "echo hi |\necho $?\nexit\n")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if !strings.Contains(out, "2") {
		t.Errorf("stdout = %q, want $? == 2 after syntax error", out)
	}
}

func requireCmd(t *testing.T, name string) {
	t.Helper()
	if _, err := exec.LookPath(name); err != nil {
		t.Skipf("%s not on PATH: %v", name, err)
	}
}

func TestCatConsumesRemainingScriptInput(t *testing.T) {
	requireCmd(t, "cat")
	out, errOut, _ := run(t, "cat\nhello\nexit\n")
	if !strings.Contains(out, "hello") {
		t.Errorf("cat should print the remaining input; stdout = %q", out)
	}
	// The remaining input must be consumed by cat, not re-parsed as commands.
	if strings.Contains(errOut, "not a mimixbox command") || strings.Contains(errOut, "command not found") {
		t.Errorf("remaining stdin was reparsed as commands; stderr = %q", errOut)
	}
}

func TestNonStdinCommandsRunInSequence(t *testing.T) {
	requireCmd(t, "echo")
	out, _, _ := run(t, "echo first\necho second\nexit\n")
	if !strings.Contains(out, "first") || !strings.Contains(out, "second") {
		t.Errorf("a command that ignores stdin must not swallow the next line; stdout = %q", out)
	}
}

func TestPartialStdinConsumptionNoByteLoss(t *testing.T) {
	requireCmd(t, "echo")
	requireCmd(t, "cat")
	// echo ignores stdin and runs; then cat consumes exactly the remaining
	// lines with no loss or duplication.
	out, _, _ := run(t, "echo top\ncat\nr1\nr2\nexit\n")
	for _, want := range []string{"top", "r1", "r2"} {
		if !strings.Contains(out, want) {
			t.Errorf("stdout = %q, want it to contain %q", out, want)
		}
	}
}
