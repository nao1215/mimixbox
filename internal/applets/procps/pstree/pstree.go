// Package pstree implements the pstree applet: show the running processes as a
// tree, read from /proc.
package pstree

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the pstree applet.
type Command struct{}

// New returns a pstree command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "pstree" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Display the process tree" }

// procDir is the /proc mount; tests point it at a fixture.
var procDir = "/proc"

type proc struct {
	pid  int
	comm string
	ppid int
}

// Run executes pstree.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "", stdio.Err).WithHelp(command.Help{
		Description: "Show the running processes as a tree, each node printed as 'name(pid)' with its " +
			"children indented beneath it, read from /proc.",
		Examples: []command.Example{
			{Command: "pstree", Explain: "Print the process tree."},
		},
		Notes: []string{
			"This is the indented-tree form; the compacted classic pstree layout is not reproduced.",
		},
		ExitStatus: "0  success.\n1  an error occurred (e.g. /proc could not be read or an argument was invalid).",
	})
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	procs := readProcs()
	children := map[int][]int{}
	known := map[int]bool{}
	for _, p := range procs {
		known[p.pid] = true
	}
	byPid := map[int]proc{}
	for _, p := range procs {
		byPid[p.pid] = p
		children[p.ppid] = append(children[p.ppid], p.pid)
	}

	// Roots: processes whose parent is not itself a known process (pid 1, or
	// orphans whose ppid is 0).
	var roots []int
	for _, p := range procs {
		if !known[p.ppid] || p.ppid == p.pid {
			roots = append(roots, p.pid)
		}
	}
	sort.Ints(roots)
	for _, r := range children {
		sort.Ints(r)
	}

	for _, r := range roots {
		printTree(stdio.Out, byPid, children, r, "", true)
	}
	return nil
}

// printTree prints the subtree rooted at pid.
func printTree(out io.Writer, byPid map[int]proc, children map[int][]int, pid int, prefix string, root bool) {
	p := byPid[pid]
	if root {
		_, _ = fmt.Fprintf(out, "%s(%d)\n", p.comm, p.pid)
	}
	kids := children[pid]
	for i, kid := range kids {
		last := i == len(kids)-1
		connector, childPrefix := "├─", prefix+"│ "
		if last {
			connector, childPrefix = "└─", prefix+"  "
		}
		kp := byPid[kid]
		_, _ = fmt.Fprintf(out, "%s%s%s(%d)\n", prefix, connector, kp.comm, kp.pid)
		printTree(out, byPid, children, kid, childPrefix, false)
	}
}

// readProcs reads every process's pid, comm and ppid from /proc/*/stat.
func readProcs() []proc {
	entries, err := os.ReadDir(procDir)
	if err != nil {
		return nil
	}
	var procs []proc
	for _, e := range entries {
		pid, err := strconv.Atoi(e.Name())
		if err != nil {
			continue
		}
		data, err := os.ReadFile(filepath.Join(procDir, e.Name(), "stat")) //nolint:gosec // /proc path
		if err != nil {
			continue
		}
		comm, ppid, ok := parseStat(string(data))
		if !ok {
			continue
		}
		procs = append(procs, proc{pid: pid, comm: comm, ppid: ppid})
	}
	return procs
}

// parseStat extracts the comm (in parentheses) and the ppid (the field after
// the state) from a /proc/PID/stat line.
func parseStat(line string) (comm string, ppid int, ok bool) {
	open := strings.IndexByte(line, '(')
	closeP := strings.LastIndexByte(line, ')')
	if open < 0 || closeP < 0 || closeP < open {
		return "", 0, false
	}
	comm = line[open+1 : closeP]
	rest := strings.Fields(line[closeP+1:])
	if len(rest) < 2 { // state, ppid, ...
		return "", 0, false
	}
	ppid, err := strconv.Atoi(rest[1])
	if err != nil {
		return "", 0, false
	}
	return comm, ppid, true
}
