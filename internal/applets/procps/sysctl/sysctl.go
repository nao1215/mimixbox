// Package sysctl implements the sysctl applet: read and write kernel parameters
// through /proc/sys.
package sysctl

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the sysctl applet.
type Command struct{}

// New returns a sysctl command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "sysctl" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Read and write kernel parameters at runtime" }

// sysDir is the sysctl tree; tests point it at a fixture.
var sysDir = "/proc/sys"

// Run executes sysctl.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[-a] [-w] NAME[=VALUE]...", stdio.Err).WithHelp(command.Help{
		Description: "Read or write kernel parameters under /proc/sys. Names use dots (kernel.ostype). " +
			"With -a, list every parameter; NAME=VALUE (or -w) writes a value.",
		Examples: []command.Example{
			{Command: "sysctl kernel.ostype", Explain: "Print one parameter."},
			{Command: "sysctl -w net.ipv4.ip_forward=1", Explain: "Set a parameter."},
		},
		ExitStatus: "0  success.\n1  a parameter was missing or could not be set.",
	})
	all := fs.BoolP("all", "a", false, "list all parameters")
	_ = fs.BoolP("write", "w", false, "write a NAME=VALUE setting")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	if *all {
		c.listAll(stdio.Out)
		return nil
	}

	names := fs.Args()
	if len(names) == 0 {
		_, _ = fmt.Fprintln(stdio.Err, "sysctl: a parameter name is required")
		return command.SilentFailure()
	}

	failed := false
	for _, arg := range names {
		if name, value, isSet := strings.Cut(arg, "="); isSet {
			if err := c.write(stdio.Out, name, value); err != nil {
				_, _ = fmt.Fprintf(stdio.Err, "sysctl: %v\n", err)
				failed = true
			}
			continue
		}
		if err := c.read(stdio.Out, arg); err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "sysctl: %v\n", err)
			failed = true
		}
	}
	if failed {
		return command.SilentFailure()
	}
	return nil
}

// read prints one parameter as "name = value".
func (c *Command) read(out io.Writer, name string) error {
	data, err := os.ReadFile(pathFor(name)) //nolint:gosec // sysctl path
	if err != nil {
		return fmt.Errorf("cannot read %s: %w", name, err)
	}
	_, _ = fmt.Fprintf(out, "%s = %s\n", name, strings.TrimRight(string(data), "\n"))
	return nil
}

// pathFor maps a dotted sysctl name to its file path.
func pathFor(name string) string {
	return filepath.Join(sysDir, filepath.FromSlash(strings.ReplaceAll(name, ".", "/")))
}

func (c *Command) write(out io.Writer, name, value string) error {
	if err := os.WriteFile(pathFor(name), []byte(value), 0o644); err != nil { //nolint:gosec // sysctl path
		return fmt.Errorf("cannot set %s: %w", name, err)
	}
	_, _ = fmt.Fprintf(out, "%s = %s\n", name, value)
	return nil
}

// listAll walks the sysctl tree and prints every readable parameter.
func (c *Command) listAll(out io.Writer) {
	_ = filepath.WalkDir(sysDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		data, rerr := os.ReadFile(path) //nolint:gosec // sysctl path
		if rerr != nil {
			return nil
		}
		rel, _ := filepath.Rel(sysDir, path)
		name := strings.ReplaceAll(filepath.ToSlash(rel), "/", ".")
		_, _ = fmt.Fprintf(out, "%s = %s\n", name, strings.TrimRight(string(data), "\n"))
		return nil
	})
}
