// Package lsof implements the lsof applet: list open files of processes, read
// from /proc.
package lsof

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

// Command is the lsof applet.
type Command struct{}

// New returns a lsof command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "lsof" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "List open files of processes" }

// procDir is the /proc mount; tests point it at a fixture.
var procDir = "/proc"

// Run executes lsof.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[-p PID]", stdio.Err).WithHelp(command.Help{
		Description: "List the open files of processes, one row per file, with the command, PID, owner, " +
			"the file descriptor (cwd/rtd/txt or a number), and the path. With -p, restrict the " +
			"listing to one process.",
		Examples: []command.Example{
			{Command: "lsof -p 1234", Explain: "List process 1234's open files."},
		},
		ExitStatus: "0  success.\n1  the requested PID could not be read.",
	})
	pid := fs.IntP("pid", "p", 0, "list only this PID's open files")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	var pids []int
	if fs.Changed("pid") {
		pids = []int{*pid}
	} else {
		pids = allPids()
	}

	tw := tabwriter.NewWriter(stdio.Out, 0, 0, 1, ' ', 0)
	_, _ = fmt.Fprintln(tw, "COMMAND\tPID\tUSER\tFD\tNAME")
	found := false
	for _, p := range pids {
		if listProcess(tw, p) {
			found = true
		}
	}
	_ = tw.Flush()

	if fs.Changed("pid") && !found {
		return command.SilentFailure()
	}
	return nil
}

// listProcess writes one process's open files; it reports whether the process
// could be read.
func listProcess(tw *tabwriter.Writer, pid int) bool {
	dir := filepath.Join(procDir, strconv.Itoa(pid))
	comm := field(dir, "comm")
	if comm == "" {
		return false
	}
	owner := ownerOf(dir)

	row := func(fd, name string) {
		if name != "" {
			_, _ = fmt.Fprintf(tw, "%s\t%d\t%s\t%s\t%s\n", comm, pid, owner, fd, name)
		}
	}
	row("cwd", readlink(filepath.Join(dir, "cwd")))
	row("rtd", readlink(filepath.Join(dir, "root")))
	row("txt", readlink(filepath.Join(dir, "exe")))

	fds, err := os.ReadDir(filepath.Join(dir, "fd"))
	if err == nil {
		var nums []int
		for _, e := range fds {
			if n, err := strconv.Atoi(e.Name()); err == nil {
				nums = append(nums, n)
			}
		}
		sort.Ints(nums)
		for _, n := range nums {
			row(strconv.Itoa(n), readlink(filepath.Join(dir, "fd", strconv.Itoa(n))))
		}
	}
	return true
}

func allPids() []int {
	entries, err := os.ReadDir(procDir)
	if err != nil {
		return nil
	}
	var pids []int
	for _, e := range entries {
		if n, err := strconv.Atoi(e.Name()); err == nil {
			pids = append(pids, n)
		}
	}
	sort.Ints(pids)
	return pids
}

func field(dir, name string) string {
	data, err := os.ReadFile(filepath.Join(dir, name)) //nolint:gosec // /proc path
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func readlink(path string) string {
	dest, err := os.Readlink(path)
	if err != nil {
		return ""
	}
	return dest
}

// ownerOf returns the user name owning the process directory.
func ownerOf(dir string) string {
	info, err := os.Stat(dir)
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
