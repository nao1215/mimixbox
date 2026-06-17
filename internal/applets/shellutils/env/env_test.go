package env_test

import (
	"bytes"
	"context"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/shellutils/env"
	"github.com/nao1215/mimixbox/internal/command"
	"github.com/nao1215/mimixbox/internal/testutil/fakecmd"
)

func run(t *testing.T, stdin string, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(stdin), Out: out, Err: errBuf}
	err := env.New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func TestRunPrintEnviron(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantContain []string
		wantAbsent  []string
	}{
		{
			name:        "assignment operand appears in output",
			args:        []string{"FOO=bar"},
			wantContain: []string{"FOO=bar", "MIMIX_BASE=present"},
		},
		{
			name:        "ignore-environment prints only the assignment",
			args:        []string{"-i", "ONLY=one"},
			wantContain: []string{"ONLY=one"},
			wantAbsent:  []string{"MIMIX_BASE="},
		},
		{
			name:        "unset removes a variable",
			args:        []string{"-u", "MIMIX_BASE"},
			wantContain: []string{},
			wantAbsent:  []string{"MIMIX_BASE="},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("MIMIX_BASE", "present")
			out, _, err := run(t, "", tt.args...)
			if err != nil {
				t.Fatalf("Run error = %v", err)
			}
			for _, want := range tt.wantContain {
				if !strings.Contains(out, want) {
					t.Errorf("output %q does not contain %q", out, want)
				}
			}
			for _, absent := range tt.wantAbsent {
				if strings.Contains(out, absent) {
					t.Errorf("output %q unexpectedly contains %q", out, absent)
				}
			}
		})
	}
}

func TestRunNullSeparator(t *testing.T) {
	out, _, err := run(t, "", "-i", "-0", "A=1", "B=2")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	want := "A=1\x00B=2\x00"
	if out != want {
		t.Errorf("out = %q, want %q", out, want)
	}
	if strings.Contains(out, "\n") {
		t.Errorf("out = %q, want no newline separators", out)
	}
}

func TestRunCommand(t *testing.T) {
	fakecmd.UseOnly(t, "echo")
	echo, err := exec.LookPath("echo")
	if err != nil {
		t.Fatalf("fake echo not found: %v", err)
	}
	out, _, runErr := run(t, "", echo, "hello")
	if runErr != nil {
		t.Fatalf("Run error = %v", runErr)
	}
	if out != "hello\n" {
		t.Errorf("out = %q, want %q", out, "hello\n")
	}
}

func TestRunCommandSeesModifiedEnv(t *testing.T) {
	fakecmd.UseOnly(t, "sh")
	sh, err := exec.LookPath("sh")
	if err != nil {
		t.Fatalf("fake sh not found: %v", err)
	}
	out, _, runErr := run(t, "", "GREETING=hi", sh, "-c", "printf %s \"$GREETING\"")
	if runErr != nil {
		t.Fatalf("Run error = %v", runErr)
	}
	if out != "hi" {
		t.Errorf("out = %q, want %q", out, "hi")
	}
}

func TestRunCommandExitStatus(t *testing.T) {
	fakecmd.UseOnly(t, "sh")
	sh, err := exec.LookPath("sh")
	if err != nil {
		t.Fatalf("fake sh not found: %v", err)
	}
	_, _, runErr := run(t, "", sh, "-c", "exit 3")
	exitErr, ok := runErr.(*command.ExitError)
	if !ok {
		t.Fatalf("error = %v (%T), want *command.ExitError", runErr, runErr)
	}
	if exitErr.Code != 3 {
		t.Errorf("exit code = %d, want 3", exitErr.Code)
	}
}

func TestRunCommandNotFound(t *testing.T) {
	out, errOut, runErr := run(t, "", "no_such_command_xyz")
	exitErr, ok := runErr.(*command.ExitError)
	if !ok {
		t.Fatalf("error = %v (%T), want *command.ExitError", runErr, runErr)
	}
	if exitErr.Code != 127 {
		t.Errorf("exit code = %d, want 127", exitErr.Code)
	}
	if out != "" {
		t.Errorf("out = %q, want empty", out)
	}
	if !strings.Contains(errOut, "env: 'no_such_command_xyz': No such file or directory") {
		t.Errorf("stderr = %q, want not-found message", errOut)
	}
}

// TestRunChdir confirms --chdir makes the launched command run in DIR: pwd
// (via the shell's built-in) reports the requested directory.
func TestRunChdir(t *testing.T) {
	fakecmd.UseOnly(t, "sh")
	sh, err := exec.LookPath("sh")
	if err != nil {
		t.Fatalf("fake sh not found: %v", err)
	}
	dir := t.TempDir()
	// Resolve symlinks so the comparison survives /tmp -> /private/tmp etc.
	want, err := filepath.EvalSymlinks(dir)
	if err != nil {
		t.Fatalf("EvalSymlinks(%q) = %v", dir, err)
	}
	out, _, runErr := run(t, "", "--chdir="+dir, sh, "-c", "pwd -P")
	if runErr != nil {
		t.Fatalf("Run error = %v", runErr)
	}
	if got := strings.TrimSpace(out); got != want {
		t.Errorf("pwd = %q, want %q", got, want)
	}
}

// TestRunChdirShortFlag exercises the -C spelling with a separate argument.
func TestRunChdirShortFlag(t *testing.T) {
	fakecmd.UseOnly(t, "sh")
	sh, err := exec.LookPath("sh")
	if err != nil {
		t.Fatalf("fake sh not found: %v", err)
	}
	dir := t.TempDir()
	want, err := filepath.EvalSymlinks(dir)
	if err != nil {
		t.Fatalf("EvalSymlinks(%q) = %v", dir, err)
	}
	out, _, runErr := run(t, "", "-C", dir, sh, "-c", "pwd -P")
	if runErr != nil {
		t.Fatalf("Run error = %v", runErr)
	}
	if got := strings.TrimSpace(out); got != want {
		t.Errorf("pwd = %q, want %q", got, want)
	}
}

// TestRunChdirNonexistent confirms a directory that cannot be used is a fatal
// error (exit 125, GNU env's failure status) and the command is not run.
func TestRunChdirNonexistent(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "does-not-exist")
	_, errOut, runErr := run(t, "", "--chdir="+missing, "true")
	exitErr, ok := runErr.(*command.ExitError)
	if !ok {
		t.Fatalf("error = %v (%T), want *command.ExitError", runErr, runErr)
	}
	if exitErr.Code != 125 {
		t.Errorf("exit code = %d, want 125", exitErr.Code)
	}
	if !strings.Contains(errOut, "cannot change directory") {
		t.Errorf("stderr = %q, want a chdir failure", errOut)
	}
}

// TestRunSplitStringExpandsArgv shows a single -S string is split into several
// argv entries: the command is the first token and the rest are its arguments.
func TestRunSplitStringExpandsArgv(t *testing.T) {
	fakecmd.UseOnly(t, "printf")
	printf, err := exec.LookPath("printf")
	if err != nil {
		t.Fatalf("fake printf not found: %v", err)
	}
	// "-S" carries the command and two of its arguments in one string.
	out, _, runErr := run(t, "", "-S", printf+" %s-%s a b")
	if runErr != nil {
		t.Fatalf("Run error = %v", runErr)
	}
	if out != "a-b" {
		t.Errorf("out = %q, want %q", out, "a-b")
	}
}

// TestRunSplitStringEscapes verifies the \_ (space) escape keeps a word
// together so it reaches the command as a single argument.
func TestRunSplitStringEscapes(t *testing.T) {
	fakecmd.UseOnly(t, "sh")
	sh, err := exec.LookPath("sh")
	if err != nil {
		t.Fatalf("fake sh not found: %v", err)
	}
	// The \_ produces a literal space inside the final argument, which the
	// shell receives as $0 of "printf %s \"$0\"".
	out, _, runErr := run(t, "", "--split-string="+sh+` -c printf\_%s\_"$0" one\_word`)
	if runErr != nil {
		t.Fatalf("Run error = %v", runErr)
	}
	if out != "one word" {
		t.Errorf("out = %q, want %q", out, "one word")
	}
}

// TestRunIgnoreSignalRejectsBadName confirms --ignore-signal validates names
// and fails (without running a command) on an unknown one.
func TestRunIgnoreSignalRejectsBadName(t *testing.T) {
	_, errOut, runErr := run(t, "", "--ignore-signal=BOGUS", "true")
	exitErr, ok := runErr.(*command.ExitError)
	if !ok {
		t.Fatalf("error = %v (%T), want *command.ExitError", runErr, runErr)
	}
	if exitErr.Code != 125 {
		t.Errorf("exit code = %d, want 125", exitErr.Code)
	}
	if !strings.Contains(errOut, "invalid signal") {
		t.Errorf("stderr = %q, want an invalid-signal message", errOut)
	}
}

// TestRunIgnoreSignalValidNames accepts good names and still runs the command.
func TestRunIgnoreSignalValidNames(t *testing.T) {
	fakecmd.UseOnly(t, "echo")
	echo, err := exec.LookPath("echo")
	if err != nil {
		t.Fatalf("fake echo not found: %v", err)
	}
	out, _, runErr := run(t, "", "--ignore-signal=INT,TERM", echo, "ok")
	if runErr != nil {
		t.Fatalf("Run error = %v", runErr)
	}
	if out != "ok\n" {
		t.Errorf("out = %q, want %q", out, "ok\n")
	}
}

func TestRunHelpAndVersion(t *testing.T) {
	out, _, err := run(t, "", "--help")
	if err != nil {
		t.Fatalf("--help error = %v", err)
	}
	if !strings.Contains(out, "Usage: env") {
		t.Errorf("--help out = %q", out)
	}
	for _, want := range []string{"Examples:", "Exit status:"} {
		if !strings.Contains(out, want) {
			t.Errorf("--help output missing %q:\n%s", want, out)
		}
	}

	out, _, err = run(t, "", "--version")
	if err != nil {
		t.Fatalf("--version error = %v", err)
	}
	if !strings.Contains(out, "env (mimixbox)") {
		t.Errorf("--version out = %q", out)
	}
}
