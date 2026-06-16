package chgrp_test

import (
	"bytes"
	"context"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/fileutils/chgrp"
	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: errBuf}
	err := chgrp.New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

// ownGroupName resolves the name of one of the current user's own groups, which
// can be assigned without root privileges.
func ownGroupName(t *testing.T) string {
	t.Helper()
	gids, err := os.Getgroups()
	if err != nil || len(gids) == 0 {
		t.Skipf("cannot determine own groups: %v", err)
	}
	for _, gid := range gids {
		g, err := user.LookupGroupId(strconv.Itoa(gid))
		if err == nil {
			return g.Name
		}
	}
	t.Skip("no resolvable group name for current user")
	return ""
}

func TestChgrpToOwnGroup(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "f.txt")
	if err := os.WriteFile(file, []byte("x\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	group := ownGroupName(t)
	g, err := user.LookupGroup(group)
	if err != nil {
		t.Fatalf("LookupGroup(%q): %v", group, err)
	}
	wantGid, err := strconv.Atoi(g.Gid)
	if err != nil {
		t.Fatalf("Atoi(%q): %v", g.Gid, err)
	}

	_, errOut, err := run(t, group, file)
	if err != nil {
		// Some sandboxes forbid chown entirely (even to one's own group); treat
		// that as unavailable rather than a failure.
		if strings.Contains(errOut, "not permitted") {
			t.Skipf("chown is not permitted in this environment: %s", errOut)
		}
		t.Fatalf("Run error = %v, stderr = %q", err, errOut)
	}
	if errOut != "" {
		t.Errorf("stderr = %q, want empty", errOut)
	}

	var st syscall.Stat_t
	if err := syscall.Stat(file, &st); err != nil {
		t.Fatal(err)
	}
	if int(st.Gid) != wantGid {
		t.Errorf("gid = %d, want %d", st.Gid, wantGid)
	}
}

func TestChgrpInvalidGroup(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "f.txt")
	if err := os.WriteFile(file, []byte("x\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	const bogus = "no_such_group_zzz"
	_, errOut, err := run(t, bogus, file)
	// A non-nil error maps to exit status 1 (command.ExitFailure) in the runner.
	if err == nil {
		t.Fatal("expected error (exit 1) for unknown group")
	}
	want := "chgrp: invalid group: '" + bogus + "'"
	if !strings.Contains(errOut, want) {
		t.Errorf("stderr = %q, want to contain %q", errOut, want)
	}
}

func TestChgrpMissingOperand(t *testing.T) {
	// No operands at all.
	_, errOut, err := run(t)
	if err == nil {
		t.Fatal("expected error for missing operand")
	}
	if !strings.Contains(errOut, "chgrp: missing operand") {
		t.Errorf("stderr = %q, want missing operand error", errOut)
	}

	// Only a group, no file.
	_, errOut, err = run(t, "root")
	if err == nil {
		t.Fatal("expected error for missing file operand")
	}
	if !strings.Contains(errOut, "chgrp: missing operand") {
		t.Errorf("stderr = %q, want missing operand error", errOut)
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
