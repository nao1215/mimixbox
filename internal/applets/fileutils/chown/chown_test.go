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
