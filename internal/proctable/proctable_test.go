package proctable

import (
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"testing"
)

// fixtureProc builds a fake /proc tree of <pid>/comm files and returns its root.
func fixtureProc(t *testing.T, procs map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for pid, comm := range procs {
		pdir := filepath.Join(dir, pid)
		if err := os.MkdirAll(pdir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(pdir, "comm"), []byte(comm+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	// A non-numeric entry must be ignored.
	if err := os.MkdirAll(filepath.Join(dir, "self"), 0o755); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestListAscendingAndSkipsNonNumeric(t *testing.T) {
	dir := fixtureProc(t, map[string]string{"30": "sshd", "10": "bash", "20": "sshd"})
	procs, err := List(dir)
	if err != nil {
		t.Fatalf("List error = %v", err)
	}
	want := []Process{{10, "bash"}, {20, "sshd"}, {30, "sshd"}}
	if !reflect.DeepEqual(procs, want) {
		t.Errorf("List = %v, want %v", procs, want)
	}
}

func TestListUnreadableDir(t *testing.T) {
	if _, err := List(filepath.Join(t.TempDir(), "nope")); err == nil {
		t.Error("expected error for missing proc dir")
	}
}

func TestMatchRegexpAscending(t *testing.T) {
	dir := fixtureProc(t, map[string]string{"10": "sshd", "20": "bash", "30": "sshd"})
	got := MatchRegexp(dir, regexp.MustCompile("sshd"))
	want := []int{10, 30}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("MatchRegexp = %v, want %v", got, want)
	}
}

func TestMatchRegexpUnreadableIsNil(t *testing.T) {
	if got := MatchRegexp(filepath.Join(t.TempDir(), "nope"), regexp.MustCompile(".")); got != nil {
		t.Errorf("MatchRegexp on unreadable dir = %v, want nil", got)
	}
}

func TestMatchNamesNewestFirst(t *testing.T) {
	procs := []Process{{100, "init"}, {150, "nginx"}, {200, "bash"}, {300, "nginx"}}
	got := MatchNames(procs, []string{"nginx"}, false)
	want := []int{300, 150}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("MatchNames = %v, want %v", got, want)
	}
}

func TestMatchNamesBasenameAndSingle(t *testing.T) {
	procs := []Process{{150, "nginx"}, {300, "nginx"}}
	got := MatchNames(procs, []string{"/usr/sbin/nginx"}, true)
	if want := []int{300}; !reflect.DeepEqual(got, want) {
		t.Errorf("MatchNames single = %v, want %v", got, want)
	}
}

func TestMatchNamesNoMatch(t *testing.T) {
	procs := []Process{{1, "init"}}
	if got := MatchNames(procs, []string{"none"}, false); got != nil {
		t.Errorf("MatchNames no match = %v, want nil", got)
	}
}
