// Package minips implements the minips applet: a minimal process lister showing
// the PID, owning user, and command, read from /proc.
package minips

import (
	"context"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"text/tabwriter"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the minips applet.
type Command struct{}

// New returns a minips command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "minips" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Minimal process lister (PID, user, command)" }

// procDir is the /proc mount; tests point it at a fixture.
var procDir = "/proc"

// Run executes minips.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "", stdio.Err).WithHelp(command.Help{
		Description: "List every process with its PID, owning user, and command name, read from /proc " +
			"and sorted by PID. This is a minimal subset of ps.",
		Examples: []command.Example{
			{Command: "minips", Explain: "List processes minimally."},
		},
	})
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	tw := tabwriter.NewWriter(stdio.Out, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(tw, "PID\tUSER\tCOMMAND")
	for _, pid := range pids() {
		comm, ok := comm(pid)
		if !ok {
			continue
		}
		_, _ = fmt.Fprintf(tw, "%d\t%s\t%s\n", pid, owner(pid), comm)
	}
	_ = tw.Flush()
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

// comm reads a process's command name from /proc/PID/comm.
func comm(pid int) (string, bool) {
	data, err := os.ReadFile(filepath.Join(procDir, strconv.Itoa(pid), "comm")) //nolint:gosec // /proc path
	if err != nil {
		return "", false
	}
	return strings.TrimSpace(string(data)), true
}

// owner returns the user name owning the process.
func owner(pid int) string {
	info, err := os.Stat(filepath.Join(procDir, strconv.Itoa(pid)))
	if err != nil {
		return "?"
	}
	st, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return "?"
	}
	uid := strconv.FormatUint(uint64(st.Uid), 10)
	if u, err := user.LookupId(uid); err == nil {
		return u.Username
	}
	return uid
}
