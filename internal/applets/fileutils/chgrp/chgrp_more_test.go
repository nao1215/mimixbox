package chgrp_test

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/fileutils/chgrp"
)

// TestSynopsis covers the Synopsis metadata accessor.
func TestSynopsis(t *testing.T) {
	c := chgrp.New()
	if c.Synopsis() != "Change the group of each FILE to GROUP" {
		t.Errorf("Synopsis() = %q", c.Synopsis())
	}
}

// TestChgrpReportMissingFile drives the report() and changeGroup() error paths
// by chgrp'ing a file that does not exist: os.Chown fails, report writes a
// GNU-style diagnostic, and Run exits non-zero.
func TestChgrpReportMissingFile(t *testing.T) {
	dir := t.TempDir()
	missing := filepath.Join(dir, "nope.txt")

	_, errOut, err := run(t, "0", missing)
	if err == nil {
		t.Fatal("expected error (exit 1) for missing file")
	}
	if !strings.Contains(errOut, "changing group of '"+missing+"'") {
		t.Errorf("stderr = %q, want changing-group diagnostic", errOut)
	}
}

// TestChgrpRecursiveOwnGroup drives changeGroupRecursive over a small tree,
// assigning every entry to one of the caller's own groups (no root needed).
func TestChgrpRecursiveOwnGroup(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "sub")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	f1 := filepath.Join(dir, "a.txt")
	f2 := filepath.Join(sub, "b.txt")
	for _, f := range []string{f1, f2} {
		if err := os.WriteFile(f, []byte("x\n"), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	group := ownGroupName(t)
	wantGid := gidOf(t, group)

	out, errOut, err := run(t, "-Rv", group, dir)
	if err != nil {
		if strings.Contains(errOut, "not permitted") {
			t.Skipf("chown is not permitted in this environment: %s", errOut)
		}
		t.Fatalf("Run -Rv error = %v, stderr = %q", err, errOut)
	}
	// -v reports each processed path, including the nested one.
	if !strings.Contains(out, "changed group of '"+f2+"'") {
		t.Errorf("verbose stdout = %q, want nested file reported", out)
	}

	for _, p := range []string{dir, sub, f1, f2} {
		var st syscall.Stat_t
		if err := syscall.Stat(p, &st); err != nil {
			t.Fatal(err)
		}
		if int(st.Gid) != wantGid {
			t.Errorf("%s gid = %d, want %d", p, st.Gid, wantGid)
		}
	}
}

// TestChgrpRecursiveMissingPath drives the filepath.Walk error branch (and the
// report path inside it) by recursing into a path that does not exist.
func TestChgrpRecursiveMissingPath(t *testing.T) {
	dir := t.TempDir()
	missing := filepath.Join(dir, "ghost")

	_, errOut, err := run(t, "-R", "0", missing)
	if err == nil {
		t.Fatal("expected error for missing recursive path")
	}
	if !strings.Contains(errOut, "chgrp:") {
		t.Errorf("stderr = %q, want chgrp diagnostic", errOut)
	}
}

// TestChgrpNumericGidWithoutEntry covers lookupGid's numeric fallback: a gid
// with no /etc/group entry still resolves. Use a high, almost-certainly-unused
// gid so LookupGroup fails but the numeric branch succeeds.
func TestChgrpNumericGidWithoutEntry(t *testing.T) {
	if _, ok := chgrp.LookupGidForTest("4000000000"); !ok {
		t.Error("LookupGid did not fall back to a numeric gid")
	}
	if _, ok := chgrp.LookupGidForTest("definitely_not_a_group"); ok {
		t.Error("LookupGid resolved a non-existent group name")
	}
}

func gidOf(t *testing.T, group string) int {
	t.Helper()
	gid, ok := chgrp.LookupGidForTest(group)
	if !ok {
		// Fall back to numeric parse if the helper somehow cannot resolve it.
		n, err := strconv.Atoi(group)
		if err != nil {
			t.Fatalf("cannot resolve gid for %q", group)
		}
		return n
	}
	return gid
}
