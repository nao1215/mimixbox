package chown

import (
	"bytes"
	"context"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func currentIDs(t *testing.T) (uid, gid int, name string) {
	t.Helper()
	u, err := user.Current()
	if err != nil {
		t.Fatalf("user.Current() = %v", err)
	}
	uid, err = strconv.Atoi(u.Uid)
	if err != nil {
		t.Fatalf("parse uid %q: %v", u.Uid, err)
	}
	gid, err = strconv.Atoi(u.Gid)
	if err != nil {
		t.Fatalf("parse gid %q: %v", u.Gid, err)
	}
	return uid, gid, u.Username
}

func TestParseOwner(t *testing.T) {
	t.Parallel()
	uid, gid, name := currentIDs(t)

	tests := []struct {
		title   string
		spec    string
		wantUID int
		wantGID int
		wantErr bool
	}{
		{"owner name", name, uid, -1, false},
		{"owner:group name", name + ":" + name, uid, gid, false},
		{"group only", ":" + name, -1, gid, false},
		{"numeric owner:group", strconv.Itoa(uid) + ":" + strconv.Itoa(gid), uid, gid, false},
		{"numeric owner", strconv.Itoa(uid), uid, -1, false},
		{"invalid user", "definitely-no-such-user-xyz", -1, -1, true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.title, func(t *testing.T) {
			t.Parallel()
			gotUID, gotGID, err := parseOwner(tt.spec)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("parseOwner(%q) err = nil, want error", tt.spec)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseOwner(%q) err = %v", tt.spec, err)
			}
			if gotUID != tt.wantUID || gotGID != tt.wantGID {
				t.Errorf("parseOwner(%q) = (%d, %d), want (%d, %d)",
					tt.spec, gotUID, gotGID, tt.wantUID, tt.wantGID)
			}
		})
	}
}

func run(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: errBuf}
	err := New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func TestRunChownToSelfSucceeds(t *testing.T) {
	t.Parallel()
	_, _, name := currentIDs(t)

	dir := t.TempDir()
	file := filepath.Join(dir, "f.txt")
	if err := os.WriteFile(file, []byte("x\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	// Changing to the current user's own owner:group is a no-op that needs no
	// root privilege.
	_, errOut, err := run(t, name+":"+name, file)
	if err != nil {
		t.Fatalf("Run error = %v (stderr=%q)", err, errOut)
	}
	if errOut != "" {
		t.Errorf("stderr = %q, want empty", errOut)
	}
}

func TestRunInvalidUser(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	file := filepath.Join(dir, "f.txt")
	if err := os.WriteFile(file, []byte("x\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: errBuf}
	// Execute maps the command's error to a process exit code, so an invalid
	// user must surface as ExitFailure (1).
	code := command.Execute(context.Background(), New(), io, []string{"definitely-no-such-user-xyz", file})
	if code != command.ExitFailure {
		t.Errorf("exit code = %d, want %d", code, command.ExitFailure)
	}
	if out.String() != "" {
		t.Errorf("stdout = %q, want empty", out.String())
	}
	if !strings.Contains(errBuf.String(), "chown: invalid user: 'definitely-no-such-user-xyz'") {
		t.Errorf("stderr = %q, want invalid-user message", errBuf.String())
	}
}

func TestRunMissingOperand(t *testing.T) {
	t.Parallel()
	out, errOut, err := run(t)
	if err == nil {
		t.Fatal("expected error for missing operand")
	}
	if out != "" {
		t.Errorf("stdout = %q, want empty", out)
	}
	if !strings.Contains(errOut, "chown: missing operand") {
		t.Errorf("stderr = %q, want missing-operand message", errOut)
	}
}

func TestSynopsis(t *testing.T) {
	t.Parallel()
	c := New()
	if c.Name() != "chown" {
		t.Errorf("Name() = %q, want chown", c.Name())
	}
	if c.Synopsis() == "" {
		t.Error("Synopsis() is empty")
	}
}

// TestKeep verifies that the first error is preserved across multiple failing
// operands and that a nil prior error is converted into a SilentFailure.
func TestKeep(t *testing.T) {
	t.Parallel()
	first := keep(nil)
	if first == nil {
		t.Fatal("keep(nil) = nil, want a non-nil failure error")
	}
	if got := keep(first); got != first {
		t.Errorf("keep(existing) = %v, want the existing error preserved", got)
	}
}

// TestLookupGIDNumericAndUnknown exercises the numeric fallback and the
// unknown-group error path that name resolution alone does not reach.
func TestLookupGIDNumericAndUnknown(t *testing.T) {
	t.Parallel()
	gid, err := lookupGID("12345")
	if err != nil {
		t.Fatalf("lookupGID(numeric) err = %v", err)
	}
	if gid != 12345 {
		t.Errorf("lookupGID(\"12345\") = %d, want 12345", gid)
	}
	if _, err := lookupGID("definitely-no-such-group-xyz"); err == nil {
		t.Error("lookupGID on an unknown non-numeric group must error")
	}
}

// TestLookupUIDNumericAndUnknown mirrors the GID test for the uid path.
func TestLookupUIDNumericAndUnknown(t *testing.T) {
	t.Parallel()
	uid, err := lookupUID("54321")
	if err != nil {
		t.Fatalf("lookupUID(numeric) err = %v", err)
	}
	if uid != 54321 {
		t.Errorf("lookupUID(\"54321\") = %d, want 54321", uid)
	}
	if _, err := lookupUID("definitely-no-such-user-xyz"); err == nil {
		t.Error("lookupUID on an unknown non-numeric user must error")
	}
}

// TestParseOwnerInvalidGroup exercises the "invalid group" branch, which the
// existing table only reaches for users.
func TestParseOwnerInvalidGroup(t *testing.T) {
	t.Parallel()
	if _, _, err := parseOwner("0:definitely-no-such-group-xyz"); err == nil {
		t.Error("parseOwner with an unknown group must error")
	}
}

// TestRunVerboseToSelf drives the verbose diagnostic path of chownOne by
// chowning a file to the current user's own ids (a no-op needing no root).
func TestRunVerboseToSelf(t *testing.T) {
	t.Parallel()
	uid, gid, _ := currentIDs(t)

	dir := t.TempDir()
	file := filepath.Join(dir, "v.txt")
	if err := os.WriteFile(file, []byte("x\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	out, errOut, err := run(t, "-v", strconv.Itoa(uid)+":"+strconv.Itoa(gid), file)
	if err != nil {
		t.Fatalf("Run error = %v (stderr=%q)", err, errOut)
	}
	if !strings.Contains(out, "ownership of '"+file+"' retained") {
		t.Errorf("verbose stdout = %q, want a retained-ownership diagnostic", out)
	}
}

// TestRunRecursiveToSelf drives the recursive filepath.Walk path of apply over
// a directory tree, again a no-op chown to the caller's own ids.
func TestRunRecursiveToSelf(t *testing.T) {
	t.Parallel()
	uid, gid, _ := currentIDs(t)

	dir := t.TempDir()
	sub := filepath.Join(dir, "sub")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, p := range []string{filepath.Join(dir, "a.txt"), filepath.Join(sub, "b.txt")} {
		if err := os.WriteFile(p, []byte("x\n"), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	_, errOut, err := run(t, "-R", strconv.Itoa(uid)+":"+strconv.Itoa(gid), dir)
	if err != nil {
		t.Fatalf("recursive Run error = %v (stderr=%q)", err, errOut)
	}
	if errOut != "" {
		t.Errorf("stderr = %q, want empty", errOut)
	}
}

// TestRunRecursiveMissingPath drives the Walk error branch: the walk function's
// err argument is non-nil for a path that does not exist.
func TestRunRecursiveMissingPath(t *testing.T) {
	t.Parallel()
	missing := filepath.Join(t.TempDir(), "no-such-dir")
	out, errOut, err := run(t, "-R", "0", missing)
	if err == nil {
		t.Fatal("recursive chown of a missing path must fail")
	}
	if out != "" {
		t.Errorf("stdout = %q, want empty", out)
	}
	if !strings.Contains(errOut, "chown: ") {
		t.Errorf("stderr = %q, want a chown error", errOut)
	}
}

// TestRunMissingOperandAfterSpec covers the "missing operand after SPEC" branch
// reached when an owner spec is given with no files.
func TestRunMissingOperandAfterSpec(t *testing.T) {
	t.Parallel()
	_, errOut, err := run(t, "0")
	if err == nil {
		t.Fatal("expected error when no file operand follows the spec")
	}
	if !strings.Contains(errOut, "missing operand after '0'") {
		t.Errorf("stderr = %q, want missing-operand-after message", errOut)
	}
}

// TestHelpSections verifies that --help renders both the Examples and the
// Exit status sections supplied through WithHelp.
func TestHelpSections(t *testing.T) {
	out, _, err := run(t, "--help")
	if err != nil {
		t.Fatalf("--help err = %v", err)
	}
	for _, want := range []string{"Examples:", "Exit status:"} {
		if !strings.Contains(out, want) {
			t.Errorf("--help missing %q section:\n%s", want, out)
		}
	}
}
