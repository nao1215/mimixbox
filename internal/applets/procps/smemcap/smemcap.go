// Package smemcap implements the smemcap applet: capture the /proc data that
// smem needs into a tar archive written to standard output.
package smemcap

import (
	"archive/tar"
	"context"
	"os"
	"path/filepath"
	"sort"
	"strconv"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the smemcap applet.
type Command struct{}

// New returns a smemcap command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "smemcap" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Capture /proc memory data into a tar for smem" }

// Injected so the capture is testable.
var (
	procDir     = "/proc"
	meminfoPath = "/proc/meminfo"
	versionPath = "/proc/version"
)

// Run executes smemcap.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "", stdio.Err).WithHelp(command.Help{
		Description: "Write a tar archive to standard output containing the system's memory-usage data " +
			"(/proc/meminfo and /proc/version) and, for each process, its smaps, stat, and cmdline. " +
			"The archive is meant to be read later by 'smem -S'.",
		Examples: []command.Example{
			{Command: "smemcap > capture.tar", Explain: "Capture the current memory state."},
		},
		ExitStatus: "0  success.\n1  an error occurred (e.g. /proc could not be read or the archive could not be written).",
	})
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	tw := tar.NewWriter(stdio.Out)
	add(tw, "meminfo", meminfoPath)
	add(tw, "version", versionPath)
	for _, pid := range pids() {
		dir := filepath.Join(procDir, pid)
		add(tw, pid+"/smaps", filepath.Join(dir, "smaps"))
		add(tw, pid+"/stat", filepath.Join(dir, "stat"))
		add(tw, pid+"/cmdline", filepath.Join(dir, "cmdline"))
	}
	if err := tw.Close(); err != nil {
		return command.Failuref("%v", err)
	}
	return nil
}

// add writes one file into the archive, skipping it if it cannot be read.
func add(tw *tar.Writer, name, srcPath string) {
	data, err := os.ReadFile(srcPath) //nolint:gosec // /proc path
	if err != nil {
		return
	}
	hdr := &tar.Header{Name: name, Mode: 0o644, Size: int64(len(data))}
	if tw.WriteHeader(hdr) != nil {
		return
	}
	_, _ = tw.Write(data)
}

// pids returns the numeric /proc entries in sorted order.
func pids() []string {
	entries, err := os.ReadDir(procDir)
	if err != nil {
		return nil
	}
	var nums []int
	for _, e := range entries {
		if n, err := strconv.Atoi(e.Name()); err == nil {
			nums = append(nums, n)
		}
	}
	sort.Ints(nums)
	out := make([]string, len(nums))
	for i, n := range nums {
		out[i] = strconv.Itoa(n)
	}
	return out
}
