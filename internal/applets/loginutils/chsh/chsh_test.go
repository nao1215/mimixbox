package chsh

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

const samplePasswd = "root:x:0:0:root:/root:/bin/bash\n" +
	"alice:x:1000:1000:Alice:/home/alice:/bin/sh\n" +
	"bob:x:1001:1001:Bob:/home/bob:/bin/sh\n"

const sampleShells = "# comment\n/bin/sh\n/bin/bash\n\n/usr/bin/zsh\n"

// setup points the package paths at fixtures and forces a non-root euid so the
// /etc/shells validation path is exercised by default. Individual tests opt
// into root behavior with asRoot.
func setup(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	pw := filepath.Join(dir, "passwd")
	if err := os.WriteFile(pw, []byte(samplePasswd), 0o644); err != nil {
		t.Fatal(err)
	}
	sh := filepath.Join(dir, "shells")
	if err := os.WriteFile(sh, []byte(sampleShells), 0o644); err != nil {
		t.Fatal(err)
	}
	op, os2, oe := passwdPath, shellsPath, geteuid
	passwdPath, shellsPath, geteuid = pw, sh, func() int { return 1000 }
	t.Cleanup(func() { passwdPath, shellsPath, geteuid = op, os2, oe })
	return pw
}

func asRoot(t *testing.T) {
	t.Helper()
	orig := geteuid
	geteuid = func() int { return 0 }
	t.Cleanup(func() { geteuid = orig })
}

func run(t *testing.T, stdin string, args ...string) (string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(stdin), Out: out, Err: &bytes.Buffer{}}
	err := New().Run(context.Background(), io, args)
	return out.String(), err
}

func shellOf(t *testing.T, path, user string) string {
	t.Helper()
	data, _ := os.ReadFile(path)
	for _, line := range strings.Split(strings.TrimRight(string(data), "\n"), "\n") {
		f := strings.Split(line, ":")
		if f[0] == user {
			return f[6]
		}
	}
	t.Fatalf("user %s not found", user)
	return ""
}

func TestNameSynopsis(t *testing.T) {
	t.Parallel()
	c := New()
	if c.Name() != "chsh" {
		t.Errorf("Name() = %q, want chsh", c.Name())
	}
	if c.Synopsis() == "" {
		t.Error("Synopsis() is empty")
	}
}

// TestTargetUserDefaultsToCurrent verifies that with no operand the current
// user is resolved.
func TestTargetUserDefaultsToCurrent(t *testing.T) {
	t.Parallel()
	got, err := targetUser(nil)
	if err != nil {
		t.Fatalf("targetUser(nil) err = %v", err)
	}
	if got == "" {
		t.Error("targetUser(nil) returned an empty user name")
	}
}

// TestTargetUserErrors covers the operand-validation branches.
func TestTargetUserErrors(t *testing.T) {
	t.Parallel()
	if _, err := targetUser([]string{"alice", "bob"}); err == nil {
		t.Error("more than one operand must error")
	}
	if _, err := targetUser([]string{""}); err == nil {
		t.Error("an empty operand must error")
	}
	got, err := targetUser([]string{"alice"})
	if err != nil || got != "alice" {
		t.Errorf("targetUser([alice]) = (%q,%v), want (alice,nil)", got, err)
	}
}

// TestReadShellsMissingFile verifies the open-error path of readShells.
func TestReadShellsMissingFile(t *testing.T) {
	orig := shellsPath
	shellsPath = filepath.Join(t.TempDir(), "no-such-shells")
	t.Cleanup(func() { shellsPath = orig })
	if _, err := readShells(); err == nil {
		t.Error("readShells on a missing file must error")
	}
}

// TestChangeShellReadError verifies changeShell surfaces a read error when the
// passwd database is absent.
func TestChangeShellReadError(t *testing.T) {
	orig := passwdPath
	passwdPath = filepath.Join(t.TempDir(), "no-such-passwd")
	t.Cleanup(func() { passwdPath = orig })
	if err := changeShell("alice", "/bin/sh"); err == nil {
		t.Error("changeShell with a missing passwd file must error")
	}
}

// TestReadLinesEmptyFile verifies an empty passwd file yields no lines and no
// error, exercising the trimmed-empty short-circuit.
func TestReadLinesEmptyFile(t *testing.T) {
	t.Parallel()
	p := filepath.Join(t.TempDir(), "empty")
	if err := os.WriteFile(p, []byte("\n\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	lines, err := readLines(p)
	if err != nil {
		t.Fatalf("readLines err = %v", err)
	}
	if lines != nil {
		t.Errorf("readLines(empty) = %v, want nil", lines)
	}
}

// TestWriteLinesCreateTempError verifies writeLines fails when the temp file
// cannot be created because the target directory does not exist.
func TestWriteLinesCreateTempError(t *testing.T) {
	t.Parallel()
	missing := filepath.Join(t.TempDir(), "no-such-dir", "passwd")
	if err := writeLines(missing, []string{"x"}); err == nil {
		t.Error("writeLines into a non-existent directory must error")
	}
}

func TestChangeOtherUserWithFlag(t *testing.T) {
	pw := setup(t)
	if _, err := run(t, "", "-s", "/bin/bash", "alice"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := shellOf(t, pw, "alice"); got != "/bin/bash" {
		t.Errorf("alice shell = %q, want /bin/bash", got)
	}
	// Other users untouched.
	if got := shellOf(t, pw, "bob"); got != "/bin/sh" {
		t.Errorf("bob should be unchanged, got %q", got)
	}
}

func TestInteractiveReadsShellFromStdin(t *testing.T) {
	pw := setup(t)
	if _, err := run(t, "/usr/bin/zsh\n", "alice"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := shellOf(t, pw, "alice"); got != "/usr/bin/zsh" {
		t.Errorf("alice shell = %q, want /usr/bin/zsh", got)
	}
}

func TestRejectsShellNotInShells(t *testing.T) {
	setup(t)
	if _, err := run(t, "", "-s", "/bin/fish", "alice"); err == nil {
		t.Error("a shell missing from /etc/shells must be rejected for a non-root user")
	}
}

func TestRootMaySetAnyAbsoluteShell(t *testing.T) {
	pw := setup(t)
	asRoot(t)
	if _, err := run(t, "", "-s", "/opt/custom/shell", "alice"); err != nil {
		t.Fatalf("root should be allowed to set any absolute shell: %v", err)
	}
	if got := shellOf(t, pw, "alice"); got != "/opt/custom/shell" {
		t.Errorf("alice shell = %q, want /opt/custom/shell", got)
	}
}

func TestRejectsRelativeShell(t *testing.T) {
	setup(t)
	asRoot(t)
	if _, err := run(t, "", "-s", "bin/bash", "alice"); err == nil {
		t.Error("a relative shell path must be rejected")
	}
}

func TestUnknownUser(t *testing.T) {
	setup(t)
	if _, err := run(t, "", "-s", "/bin/bash", "carol"); err == nil {
		t.Error("an unknown user must fail")
	}
}

func TestEmptyShellRejected(t *testing.T) {
	setup(t)
	if _, err := run(t, "\n", "alice"); err == nil {
		t.Error("an empty shell must be rejected")
	}
}

func TestTooManyArguments(t *testing.T) {
	setup(t)
	if _, err := run(t, "", "-s", "/bin/sh", "alice", "bob"); err == nil {
		t.Error("more than one user operand must fail")
	}
}

func TestListShells(t *testing.T) {
	setup(t)
	out, err := run(t, "", "-l")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, want := range []string{"/bin/sh", "/bin/bash", "/usr/bin/zsh"} {
		if !strings.Contains(out, want) {
			t.Errorf("-l output missing %q; got %q", want, out)
		}
	}
	if strings.Contains(out, "# comment") {
		t.Error("-l must not print comment lines")
	}
}

func TestNoChangeWhenAlreadySet(t *testing.T) {
	pw := setup(t)
	asRoot(t)
	if _, err := run(t, "", "-s", "/bin/sh", "alice"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := shellOf(t, pw, "alice"); got != "/bin/sh" {
		t.Errorf("alice shell = %q, want /bin/sh", got)
	}
}

func TestAtomicWritePreservesOtherEntriesAndMode(t *testing.T) {
	pw := setup(t)
	asRoot(t)
	if _, err := run(t, "", "-s", "/bin/bash", "alice"); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(pw)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o644 {
		t.Errorf("passwd mode = %o, want 0644", info.Mode().Perm())
	}
	// root's full entry must be intact.
	data, _ := os.ReadFile(pw)
	if !strings.Contains(string(data), "root:x:0:0:root:/root:/bin/bash") {
		t.Errorf("root entry was corrupted: %q", string(data))
	}
	entries, _ := os.ReadDir(filepath.Dir(pw))
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".chsh-") {
			t.Errorf("temp file left behind: %s", e.Name())
		}
	}
}

func TestUnknownFlagFails(t *testing.T) {
	setup(t)
	if _, err := run(t, "", "--bogus"); err == nil {
		t.Error("an unknown flag must fail")
	}
}

// A shell value carrying a newline must never be able to forge a new passwd
// line (e.g. a passwordless UID-0 account), even for root.
func TestRejectsNewlineInjection(t *testing.T) {
	pw := setup(t)
	asRoot(t)
	if _, err := run(t, "", "-s", "/bin/sh\nevil::0:0:root:/root:/bin/sh", "alice"); err == nil {
		t.Fatal("a newline in the shell value must be rejected")
	}
	data, _ := os.ReadFile(pw)
	if strings.Contains(string(data), "evil") {
		t.Errorf("passwd database was corrupted by injection: %q", string(data))
	}
}

// A ':' would spill the value into adjacent passwd fields, so it must be
// rejected even for root.
func TestRejectsColonInjection(t *testing.T) {
	setup(t)
	asRoot(t)
	if _, err := run(t, "", "-s", "/bin/sh:0:0", "alice"); err == nil {
		t.Error("a colon in the shell value must be rejected")
	}
}

func TestRejectsControlCharacter(t *testing.T) {
	setup(t)
	asRoot(t)
	if _, err := run(t, "", "-s", "/bin/sh\x01", "alice"); err == nil {
		t.Error("a control character in the shell value must be rejected")
	}
}

// With no -s and immediate EOF on stdin, chsh must fail rather than succeed.
func TestStdinEOFFails(t *testing.T) {
	setup(t)
	if _, err := run(t, "", "alice"); err == nil {
		t.Error("EOF on stdin with no shell given must fail")
	}
}

// A malformed passwd line for the target user must not be matched.
func TestMalformedTargetLineIsUnknownUser(t *testing.T) {
	dir := t.TempDir()
	pw := filepath.Join(dir, "passwd")
	if err := os.WriteFile(pw, []byte("alice:x:1000:1000:Alice:/home/alice\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	op := passwdPath
	passwdPath = pw
	t.Cleanup(func() { passwdPath = op })
	asRoot(t)
	if _, err := run(t, "", "-s", "/bin/sh", "alice"); err == nil {
		t.Error("a malformed (non-7-field) passwd line must not be changed")
	}
}
