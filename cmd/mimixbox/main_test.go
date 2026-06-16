package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets"
	"github.com/nao1215/mimixbox/internal/command"
)

func newIO() (command.IO, *bytes.Buffer, *bytes.Buffer) {
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	return command.IO{In: strings.NewReader(""), Out: out, Err: errBuf}, out, errBuf
}

func TestRunHelpExitsZeroOnStdout(t *testing.T) {
	io, out, errBuf := newIO()
	if code := run([]string{"mimixbox", "--help"}, io); code != command.ExitSuccess {
		t.Fatalf("exit = %d, want 0", code)
	}
	if !strings.Contains(out.String(), "Usage: mimixbox") {
		t.Errorf("help should go to stdout, got %q", out.String())
	}
	if errBuf.Len() != 0 {
		t.Errorf("stderr should be empty, got %q", errBuf.String())
	}
}

func TestRunVersionExitsZero(t *testing.T) {
	io, out, _ := newIO()
	if code := run([]string{"mimixbox", "--version"}, io); code != command.ExitSuccess {
		t.Fatalf("exit = %d, want 0", code)
	}
	if !strings.Contains(out.String(), "mimixbox") {
		t.Errorf("version output = %q", out.String())
	}
}

func TestRunListExitsZeroOnStdout(t *testing.T) {
	io, out, _ := newIO()
	if code := run([]string{"mimixbox", "--list"}, io); code != command.ExitSuccess {
		t.Fatalf("exit = %d, want 0", code)
	}
	if !strings.Contains(out.String(), "cat") {
		t.Errorf("list should include applets like cat, got %q", out.String()[:80])
	}
}

func TestRunUnknownCommand(t *testing.T) {
	io, out, errBuf := newIO()
	if code := run([]string{"mimixbox", "frobnicate"}, io); code != command.ExitFailure {
		t.Fatalf("exit = %d, want 1", code)
	}
	if out.Len() != 0 {
		t.Errorf("stdout must stay clean on error, got %q", out.String())
	}
	if !strings.Contains(errBuf.String(), "not a mimixbox command") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}

func TestRunBareIsUsageError(t *testing.T) {
	io, _, errBuf := newIO()
	if code := run([]string{"mimixbox"}, io); code != command.ExitFailure {
		t.Fatalf("exit = %d, want 1", code)
	}
	if !strings.Contains(errBuf.String(), "Usage: mimixbox") {
		t.Errorf("bare invocation should print usage to stderr, got %q", errBuf.String())
	}
}

func TestRunDispatchesAppletByName(t *testing.T) {
	var gotName string
	var gotArgs []string
	orig := runApplet
	runApplet = func(name string, args []string, _ command.IO) int {
		gotName, gotArgs = name, args
		return 0
	}
	t.Cleanup(func() { runApplet = orig })

	io, _, _ := newIO()
	run([]string{"mimixbox", "cat", "-n", "file.txt"}, io)
	if gotName != "cat" {
		t.Errorf("dispatched to %q, want cat", gotName)
	}
	if strings.Join(gotArgs, ",") != "-n,file.txt" {
		t.Errorf("applet args = %v, want [-n file.txt]", gotArgs)
	}
}

func TestRunSymlinkDispatch(t *testing.T) {
	var gotName string
	var gotArgs []string
	orig := runApplet
	runApplet = func(name string, args []string, _ command.IO) int {
		gotName, gotArgs = name, args
		return 0
	}
	t.Cleanup(func() { runApplet = orig })

	// Invoked through a symlink named "cat".
	io, _, _ := newIO()
	run([]string{"/usr/local/bin/cat", "file.txt"}, io)
	if gotName != "cat" || strings.Join(gotArgs, ",") != "file.txt" {
		t.Errorf("symlink dispatch = %q %v", gotName, gotArgs)
	}
}

func TestRunRenamedBinaryKeepsTopLevelOptions(t *testing.T) {
	// The binary is renamed/copied/wrapped under a name that is not "mimixbox"
	// and is not a known applet. Top-level options must keep working instead of
	// being treated as an unknown applet (issue #949).
	cases := []struct {
		name string
		argv []string
		want string // substring expected on stdout
	}{
		{"version", []string{"/opt/pkg/mimixbox-review", "--version"}, "mimixbox"},
		{"help", []string{"/opt/pkg/mimixbox-review", "--help"}, "Usage: mimixbox"},
		{"list", []string{"./mimixbox-review", "--list"}, "cat"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			io, out, errBuf := newIO()
			if code := run(tc.argv, io); code != command.ExitSuccess {
				t.Fatalf("exit = %d, want 0 (stderr=%q)", code, errBuf.String())
			}
			if !strings.Contains(out.String(), tc.want) {
				t.Errorf("stdout = %q, want substring %q", out.String(), tc.want)
			}
		})
	}
}

func TestRunRenamedBinaryDispatchesAppletByName(t *testing.T) {
	// A renamed binary must still dispatch applets given as the first argument,
	// e.g. "mimixbox-review cat file.txt" runs cat.
	var gotName string
	var gotArgs []string
	orig := runApplet
	runApplet = func(name string, args []string, _ command.IO) int {
		gotName, gotArgs = name, args
		return 0
	}
	t.Cleanup(func() { runApplet = orig })

	io, _, _ := newIO()
	run([]string{"/opt/pkg/mimixbox-review", "cat", "file.txt"}, io)
	if gotName != "cat" || strings.Join(gotArgs, ",") != "file.txt" {
		t.Errorf("renamed-binary applet dispatch = %q %v", gotName, gotArgs)
	}
}

func TestRunAppletFlagsReachApplet(t *testing.T) {
	// "mimixbox cp -f a b": -f must be passed to cp, not parsed as
	// --full-install.
	var gotName string
	var gotArgs []string
	orig := runApplet
	runApplet = func(name string, args []string, _ command.IO) int {
		gotName, gotArgs = name, args
		return 0
	}
	t.Cleanup(func() { runApplet = orig })

	io, _, _ := newIO()
	run([]string{"mimixbox", "cp", "-f", "a", "b"}, io)
	if gotName != "cp" || strings.Join(gotArgs, ",") != "-f,a,b" {
		t.Errorf("cp -f dispatch = %q %v", gotName, gotArgs)
	}
}

func TestFullInstallAndRemove(t *testing.T) {
	dir := t.TempDir()
	io, _, _ := newIO()

	if code := run([]string{"mimixbox", "--full-install", dir}, io); code != command.ExitSuccess {
		t.Fatalf("full-install exit = %d", code)
	}
	// A representative applet symlink should now exist.
	if _, err := os.Lstat(filepath.Join(dir, "cat")); err != nil {
		t.Fatalf("expected cat symlink: %v", err)
	}

	io2, _, _ := newIO()
	if code := run([]string{"mimixbox", "--remove", dir}, io2); code != command.ExitSuccess {
		t.Fatalf("remove exit = %d", code)
	}
	if _, err := os.Lstat(filepath.Join(dir, "cat")); !os.IsNotExist(err) {
		t.Errorf("cat symlink should have been removed")
	}
}

func TestInstallTargetMaySharedAppletBasename(t *testing.T) {
	// A directory whose basename collides with an applet ("cat") must still be a
	// valid install target; this was previously rejected by the dispatch hacks.
	parent := t.TempDir()
	dir := filepath.Join(parent, "cat")
	if err := os.Mkdir(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	io, _, _ := newIO()
	if code := run([]string{"mimixbox", "--full-install", dir}, io); code != command.ExitSuccess {
		t.Fatalf("install into a dir named 'cat' exit = %d", code)
	}
	if _, err := os.Lstat(filepath.Join(dir, "true")); err != nil {
		t.Errorf("expected applet symlinks inside the 'cat' directory: %v", err)
	}
}

func TestInstallRequiresDirOperand(t *testing.T) {
	io, _, errBuf := newIO()
	if code := run([]string{"mimixbox", "--install"}, io); code != command.ExitFailure {
		t.Fatalf("install with no dir exit = %d, want 1", code)
	}
	if !strings.Contains(errBuf.String(), "single DIRECTORY operand") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}

func TestRunDispatchesRealAppletInMemory(t *testing.T) {
	// A concrete applet (echo) runs through the real dispatch path, and its
	// output is captured entirely in memory via the injected IO — no os.Args
	// mutation and no process stdio. This is the end-to-end property issue #273
	// asks for.
	io, out, errBuf := newIO()
	if code := run([]string{"mimixbox", "echo", "hello", "world"}, io); code != command.ExitSuccess {
		t.Fatalf("echo exit = %d, want 0", code)
	}
	if got := out.String(); got != "hello world\n" {
		t.Errorf("echo stdout = %q, want %q", got, "hello world\n")
	}
	if errBuf.Len() != 0 {
		t.Errorf("echo stderr = %q, want empty", errBuf.String())
	}
}

func TestRunRealAppletErrorIsCapturedInMemory(t *testing.T) {
	// A failing applet (cat of a missing file) reports its error through the
	// injected stderr and a non-zero exit code, again without touching process
	// globals.
	io, out, errBuf := newIO()
	code := run([]string{"mimixbox", "cat", "/no/such/file/mimixbox-test"}, io)
	if code == command.ExitSuccess {
		t.Errorf("cat of a missing file should fail, got exit 0")
	}
	if out.Len() != 0 {
		t.Errorf("cat stdout = %q, want empty", out.String())
	}
	if !strings.Contains(errBuf.String(), "cat:") {
		t.Errorf("cat stderr = %q, want it to mention the applet", errBuf.String())
	}
}

func TestRunDoesNotMutateOSArgs(t *testing.T) {
	// Dispatching an applet must leave the process-global os.Args untouched.
	saved := append([]string(nil), os.Args...)
	io, _, _ := newIO()
	_ = run([]string{"mimixbox", "echo", "x"}, io)
	if strings.Join(os.Args, " ") != strings.Join(saved, " ") {
		t.Errorf("os.Args mutated by dispatch: got %v, want %v", os.Args, saved)
	}
}

func TestRunUnsupportedAppletViaRunApplet(t *testing.T) {
	// runApplet itself reports an unknown applet (e.g. a stale symlink).
	io, _, errBuf := newIO()
	if code := runApplet("no-such-applet", nil, io); code != command.ExitFailure {
		t.Fatalf("exit = %d, want 1", code)
	}
	if !strings.Contains(errBuf.String(), "not a mimixbox command") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}

func TestRunUnknownOption(t *testing.T) {
	// A "--foo" style token is not an applet, so it falls through to the option
	// parser's default branch and is reported as unsupported on stderr.
	io, out, errBuf := newIO()
	if code := run([]string{"mimixbox", "--definitely-not-an-option"}, io); code != command.ExitFailure {
		t.Fatalf("exit = %d, want 1", code)
	}
	if out.Len() != 0 {
		t.Errorf("stdout must stay clean on error, got %q", out.String())
	}
	if !strings.Contains(errBuf.String(), "is not a mimixbox command or option") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}

func TestFullInstallCreatesOneSymlinkPerApplet(t *testing.T) {
	dir := t.TempDir()
	io, _, errBuf := newIO()

	if code := run([]string{"mimixbox", "--full-install", dir}, io); code != command.ExitSuccess {
		t.Fatalf("full-install exit = %d (stderr: %s)", code, errBuf.String())
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := len(entries), len(applets.SortApplet()); got != want {
		t.Errorf("full-install created %d entries, want %d (one per applet)", got, want)
	}
}

func TestRemoveOnlyDeletesMimixBoxSymlinks(t *testing.T) {
	// --remove must delete only the symlinks that point at the exact running
	// MimixBox binary, leaving foreign symlinks (even ones whose target merely
	// contains "mimixbox") and real files untouched.
	self, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()

	owned := filepath.Join(dir, "cat")
	if err := os.Symlink(self, owned); err != nil {
		t.Fatal(err)
	}
	// Foreign: target contains "mimixbox" but is a different binary, so the old
	// substring check would have wrongly deleted it.
	foreign := filepath.Join(dir, "pidof")
	if err := os.Symlink("/opt/not-the-same-mimixbox-wrapper", foreign); err != nil {
		t.Fatal(err)
	}
	realFile := filepath.Join(dir, "echo")
	if err := os.WriteFile(realFile, []byte("real"), 0o644); err != nil {
		t.Fatal(err)
	}

	io, _, errBuf := newIO()
	if code := run([]string{"mimixbox", "--remove", dir}, io); code != command.ExitSuccess {
		t.Fatalf("remove exit = %d (stderr: %s)", code, errBuf.String())
	}

	if _, err := os.Lstat(owned); !os.IsNotExist(err) {
		t.Errorf("MimixBox-owned symlink %q should have been removed", owned)
	}
	if _, err := os.Lstat(foreign); err != nil {
		t.Errorf("foreign symlink %q should have been left in place", foreign)
	}
	if _, err := os.Lstat(realFile); err != nil {
		t.Errorf("real file %q should have been left in place", realFile)
	}
}

func TestInstallTargetsExactInvokedBinary(t *testing.T) {
	// Even when another "mimixbox" sits earlier on PATH, --install must link to
	// the exact binary that is running (os.Executable), never to a PATH lookup.
	const wantTarget = "/custom/install/source/mimixbox"
	orig := osExecutable
	osExecutable = func() (string, error) { return wantTarget, nil }
	t.Cleanup(func() { osExecutable = orig })

	dir := t.TempDir()
	io, _, errBuf := newIO()
	if code := run([]string{"mimixbox", "--full-install", dir}, io); code != command.ExitSuccess {
		t.Fatalf("full-install exit = %d (stderr: %s)", code, errBuf.String())
	}

	target, err := os.Readlink(filepath.Join(dir, "cat"))
	if err != nil {
		t.Fatal(err)
	}
	if target != wantTarget {
		t.Errorf("symlink target = %q, want the invoked binary %q", target, wantTarget)
	}
}
