// Package ps implements the ps applet: report a snapshot of the current
// processes, read from /proc.
package ps

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the ps applet.
type Command struct{}

// New returns a ps command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "ps" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Report a snapshot of the current processes" }

// procDir is the /proc mount; tests point it at a fixture.
var procDir = "/proc"

// clockTicks is the kernel USER_HZ (jiffies per second).
const clockTicks = 100

// Run executes ps.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "", stdio.Err).WithHelp(command.Help{
		Description: "List every process with its PID, controlling terminal, accumulated CPU time, and " +
			"command name, read from /proc and sorted by PID (like ps -e).",
		Examples: []command.Example{
			{Command: "ps", Explain: "List the running processes."},
		},
		Notes: []string{
			"All processes are listed; the BSD/Unix option syntax (aux, -ef) is not parsed.",
		},
	})
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	_, _ = fmt.Fprintln(stdio.Out, "    PID TTY          TIME CMD")
	for _, pid := range pids() {
		comm, ttyNr, jiffies, ok := readStat(pid)
		if !ok {
			continue
		}
		_, _ = fmt.Fprintf(stdio.Out, "%7d %-8s %s %s\n", pid, ttyName(ttyNr), cpuTime(jiffies), comm)
	}
	return nil
}

func pids() []int {
	entries, err := os.ReadDir(procDir)
	if err != nil {
		return nil
	}
	var out []int
	for _, e := range entries {
		if n, err := strconv.Atoi(e.Name()); err == nil {
			out = append(out, n)
		}
	}
	sort.Ints(out)
	return out
}

// readStat returns a process's comm, tty number, and total CPU jiffies.
func readStat(pid int) (comm string, ttyNr int, jiffies int64, ok bool) {
	data, err := os.ReadFile(filepath.Join(procDir, strconv.Itoa(pid), "stat")) //nolint:gosec // /proc path
	if err != nil {
		return "", 0, 0, false
	}
	line := string(data)
	open := strings.IndexByte(line, '(')
	closeP := strings.LastIndexByte(line, ')')
	if open < 0 || closeP < 0 || closeP < open {
		return "", 0, 0, false
	}
	comm = line[open+1 : closeP]
	f := strings.Fields(line[closeP+1:]) // state ppid pgrp session tty_nr ... utime(11) stime(12)
	if len(f) < 13 {
		return "", 0, 0, false
	}
	ttyNr, _ = strconv.Atoi(f[4])
	utime, _ := strconv.ParseInt(f[11], 10, 64)
	stime, _ := strconv.ParseInt(f[12], 10, 64)
	return comm, ttyNr, utime + stime, true
}

// ttyName decodes a tty device number to a name, or "?" when there is none.
func ttyName(dev int) string {
	if dev == 0 {
		return "?"
	}
	major := (dev >> 8) & 0xfff
	minor := (dev & 0xff) | ((dev >> 12) & 0xfff00)
	switch major {
	case 136:
		return "pts/" + strconv.Itoa(minor)
	case 4:
		return "tty" + strconv.Itoa(minor)
	default:
		return "?"
	}
}

// cpuTime formats accumulated jiffies as HH:MM:SS.
func cpuTime(jiffies int64) string {
	secs := jiffies / clockTicks
	return fmt.Sprintf("%02d:%02d:%02d", secs/3600, (secs%3600)/60, secs%60)
}
