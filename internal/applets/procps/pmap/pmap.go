// Package pmap implements the pmap applet: report the memory map of a process,
// read from /proc/PID/maps.
package pmap

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the pmap applet.
type Command struct{}

// New returns a pmap command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "pmap" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Report the memory map of a process" }

// procDir is the /proc mount; tests point it at a fixture.
var procDir = "/proc"

// Run executes pmap.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "PID...", stdio.Err).WithHelp(command.Help{
		Description: "Print the memory mappings of each process given by PID: one line per mapping " +
			"(address, size, permissions, and name) and a total, read from /proc/PID/maps.",
		Examples: []command.Example{
			{Command: "pmap 1234", Explain: "Show process 1234's memory map."},
		},
		ExitStatus: "0  every PID was reported.\n1  a PID was invalid or could not be read.",
	})
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	pids := fs.Args()
	if len(pids) == 0 {
		_, _ = fmt.Fprintln(stdio.Err, "pmap: a PID is required")
		return command.SilentFailure()
	}

	failed := false
	for _, p := range pids {
		if _, err := strconv.Atoi(p); err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "pmap: %s: invalid process id\n", p)
			failed = true
			continue
		}
		if err := c.report(stdio, p); err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "pmap: %s: %v\n", p, err)
			failed = true
		}
	}
	if failed {
		return command.SilentFailure()
	}
	return nil
}

// report prints one process's map and total.
func (c *Command) report(stdio command.IO, pid string) error {
	f, err := os.Open(filepath.Join(procDir, pid, "maps")) //nolint:gosec // /proc path
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	_, _ = fmt.Fprintf(stdio.Out, "%s:   %s\n", pid, cmdline(pid))

	var total int64
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		start, size, perms, name, ok := parseLine(sc.Text())
		if !ok {
			continue
		}
		total += size
		_, _ = fmt.Fprintf(stdio.Out, "%016x %7dK %s %s\n", start, size, perms, name)
	}
	if err := sc.Err(); err != nil {
		return err
	}
	_, _ = fmt.Fprintf(stdio.Out, " total %12dK\n", total)
	return nil
}

// parseLine parses one /proc/PID/maps line into its start address, size in KiB,
// reformatted permissions, and display name.
func parseLine(line string) (start uint64, sizeKB int64, perms, name string, ok bool) {
	fields := strings.Fields(line)
	if len(fields) < 2 {
		return 0, 0, "", "", false
	}
	lo, hi, found := strings.Cut(fields[0], "-")
	if !found {
		return 0, 0, "", "", false
	}
	s, err1 := strconv.ParseUint(lo, 16, 64)
	e, err2 := strconv.ParseUint(hi, 16, 64)
	if err1 != nil || err2 != nil {
		return 0, 0, "", "", false
	}
	name = "[ anon ]"
	if len(fields) >= 6 {
		path := strings.Join(fields[5:], " ")
		if strings.HasPrefix(path, "[") {
			name = path
		} else {
			name = filepath.Base(path)
		}
	}
	return s, int64(e-s) / 1024, mode(fields[1]), name, true
}

// mode converts maps permissions ("r-xp") to pmap's "r-x--" form.
func mode(p string) string {
	b := []byte("-----")
	for i := 0; i < 3 && i < len(p); i++ {
		if p[i] != '-' {
			b[i] = p[i]
		}
	}
	if len(p) >= 4 && p[3] == 's' {
		b[3] = 's'
	}
	return string(b)
}

// cmdline returns the process command line, or "[unknown]".
func cmdline(pid string) string {
	data, err := os.ReadFile(filepath.Join(procDir, pid, "cmdline")) //nolint:gosec // /proc path
	if err != nil || len(data) == 0 {
		return "[unknown]"
	}
	return strings.TrimSpace(strings.ReplaceAll(string(data), "\x00", " "))
}
