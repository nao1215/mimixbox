// Package pager implements the more and less applets: page through files (or
// standard input) when standard output is a terminal, and stream straight
// through otherwise so they are safe in pipelines.
//
// The paging engine lives in core.go; the more and less commands here are thin
// front-ends that only differ in their continuation prompt, so each is just a
// small configuration of the shared core.
package pager

import (
	"context"

	"github.com/nao1215/mimixbox/internal/command"
)

// config describes the few points where the more and less front-ends differ.
// Everything else - opening inputs, terminal detection, scrolling, rendering
// and key handling - is shared by the core engine.
type config struct {
	// name is the applet name (more or less).
	name string
	// prompt is the string shown at the bottom of each screenful while waiting
	// for the reader to advance or quit.
	prompt string
}

// Command is the more or less applet. It is a thin front-end over the shared
// paging core, holding only the configuration that distinguishes the two.
type Command struct {
	cfg config
}

// NewMore returns the more applet.
func NewMore() *Command { return &Command{cfg: config{name: "more", prompt: "--More--"}} }

// NewLess returns the less applet.
func NewLess() *Command { return &Command{cfg: config{name: "less", prompt: ":"}} }

// Name returns the command name.
func (c *Command) Name() string { return c.cfg.name }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Page through text one screen at a time" }

// Run executes the pager. It parses the flags shared by both front-ends and
// then hands the work to the core engine.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[FILE]...", stdio.Err).WithHelp(command.Help{
		Description: "Display FILEs (or standard input) one screen at a time when standard output is " +
			"a terminal. When standard output is not a terminal (a pipe or file), the input is " +
			"copied straight through unchanged.",
		Examples: []command.Example{
			{Command: c.Name() + " file.txt", Explain: "Page through file.txt."},
			{Command: "ls -l | " + c.Name(), Explain: "Page command output on a terminal, or pass it through in a pipe."},
		},
		ExitStatus: "0  success.\n1  an error occurred.",
		Notes: []string{
			"Paging keys: Enter advances one screen, a line starting with q quits. Input is line-buffered, so keys take effect on Enter; raw-mode single-key paging and backward scrolling are not implemented.",
		},
	})
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	return core{cfg: c.cfg}.run(stdio, fs.Args())
}
