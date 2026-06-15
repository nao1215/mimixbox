// Package popmaildir implements the popmaildir applet: move messages out of a
// Maildir's new/ directory into a destination directory (or concatenate them to
// standard output). The workflow is entirely local-file based; no POP3 network
// access is involved.
package popmaildir

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the popmaildir applet.
type Command struct{}

// New returns a popmaildir command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "popmaildir" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Move messages from a Maildir's new/ directory" }

// Run executes popmaildir.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[-d DESTDIR] MAILDIR", stdio.Err).WithHelp(command.Help{
		Description: "Process the messages in MAILDIR/new. With -d DESTDIR each message file is moved into " +
			"DESTDIR (which is created if needed). Without -d each message is concatenated to standard " +
			"output, separated by a blank line, and removed from the Maildir. Messages are processed " +
			"in sorted filename order. This is a local Maildir workflow; POP3 network retrieval is not " +
			"implemented.",
		Examples: []command.Example{
			{Command: "popmaildir -d /tmp/out ~/Maildir", Explain: "Move all new messages into /tmp/out."},
			{Command: "popmaildir ~/Maildir", Explain: "Print and drain all new messages."},
		},
		ExitStatus: "0  success.\n1  the Maildir is missing or a message could not be moved/printed.",
		Notes: []string{
			"Network (POP3) retrieval is intentionally not implemented; only local Maildir processing is supported.",
		},
	})
	dest := fs.StringP("dest", "d", "", "move messages into this directory instead of printing them")
	keep := fs.BoolP("keep", "k", false, "keep (do not remove) processed messages")
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	rest := fs.Args()
	if len(rest) != 1 {
		fmt.Fprintf(stdio.Err, "%s: exactly one MAILDIR is required\n", c.Name())
		return command.SilentFailure()
	}
	maildir := rest[0]
	newDir := filepath.Join(maildir, "new")

	names, err := listMessages(newDir)
	if err != nil {
		fmt.Fprintf(stdio.Err, "%s: %v\n", c.Name(), err)
		return command.SilentFailure()
	}

	if *dest != "" {
		if err := os.MkdirAll(*dest, 0o700); err != nil {
			fmt.Fprintf(stdio.Err, "%s: %v\n", c.Name(), err)
			return command.SilentFailure()
		}
	}

	var failed bool
	for _, name := range names {
		src := filepath.Join(newDir, name)
		if *dest != "" {
			if err := os.Rename(src, filepath.Join(*dest, name)); err != nil {
				fmt.Fprintf(stdio.Err, "%s: %v\n", c.Name(), err)
				failed = true
			}
			continue
		}
		data, err := os.ReadFile(src) //nolint:gosec // path under given maildir
		if err != nil {
			fmt.Fprintf(stdio.Err, "%s: %v\n", c.Name(), err)
			failed = true
			continue
		}
		_, _ = stdio.Out.Write(data)
		_, _ = fmt.Fprintln(stdio.Out)
		if !*keep {
			if err := os.Remove(src); err != nil {
				fmt.Fprintf(stdio.Err, "%s: %v\n", c.Name(), err)
				failed = true
			}
		}
	}
	if failed {
		return command.SilentFailure()
	}
	return nil
}

// listMessages returns the sorted regular-file names in a Maildir new/ dir.
func listMessages(newDir string) ([]string, error) {
	entries, err := os.ReadDir(newDir)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		names = append(names, e.Name())
	}
	sort.Strings(names)
	return names, nil
}
