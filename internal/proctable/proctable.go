// Package proctable is the shared process-matching backend used by the pgrep,
// pkill, and pidof applets. Each of those applets walks /proc, reads every
// process's comm name, and selects PIDs in some way; keeping the enumeration and
// the two selection strategies here is what stops them from drifting apart.
//
// The proc mount root is a parameter so tests (and callers) can point it at a
// fixture tree of <pid>/comm files instead of the host's real /proc.
package proctable

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// DefaultProcDir is the conventional proc mount root.
const DefaultProcDir = "/proc"

// Process is one running process as far as the matchers care.
type Process struct {
	PID  int
	Name string // the comm name, trimmed of trailing whitespace
}

// List enumerates the running processes under procDir. Non-numeric entries and
// processes whose comm cannot be read are skipped. The result is in ascending
// PID order. If procDir cannot be read, a nil slice and the error are returned.
func List(procDir string) ([]Process, error) {
	entries, err := os.ReadDir(procDir)
	if err != nil {
		return nil, err
	}
	var procs []Process
	for _, e := range entries {
		pid, err := strconv.Atoi(e.Name())
		if err != nil {
			continue
		}
		data, err := os.ReadFile(filepath.Join(procDir, e.Name(), "comm")) //nolint:gosec // /proc path
		if err != nil {
			continue
		}
		procs = append(procs, Process{PID: pid, Name: strings.TrimSpace(string(data))})
	}
	sort.Slice(procs, func(i, j int) bool { return procs[i].PID < procs[j].PID })
	return procs, nil
}

// MatchRegexp returns the PIDs of processes under procDir whose name matches re,
// in ascending PID order. This is the selection pgrep and pkill use. If procDir
// cannot be read, a nil slice is returned (matching the historical pgrep
// behavior of treating an unreadable /proc as "nothing matched").
func MatchRegexp(procDir string, re *regexp.Regexp) []int {
	procs, err := List(procDir)
	if err != nil {
		return nil
	}
	var pids []int
	for _, p := range procs {
		if re.MatchString(p.Name) {
			pids = append(pids, p.PID)
		}
	}
	return pids
}

// MatchNames returns the PIDs of processes whose name equals the base name of
// any requested program, newest (highest PID) first. With single, only the
// first match is returned. This is the selection pidof uses; the supplied procs
// let pidof keep its injectable process source. The result is sorted by PID
// descending regardless of the input ordering.
func MatchNames(procs []Process, names []string, single bool) []int {
	want := make(map[string]bool, len(names))
	for _, n := range names {
		want[filepath.Base(n)] = true
	}
	var pids []int
	for _, p := range procs {
		if want[p.Name] {
			pids = append(pids, p.PID)
		}
	}
	sort.Sort(sort.Reverse(sort.IntSlice(pids)))
	if single && len(pids) > 1 {
		pids = pids[:1]
	}
	return pids
}
