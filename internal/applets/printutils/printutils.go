// Package printutils implements the classic line-printer applets lpr, lpq, and
// lpd against a local spool directory, so they can be exercised end to end
// without a real printer or network daemon.
//
// The spool directory holds one control file per queued job plus the data
// payload. The shared queue backend lives in spool.go; lpr (lpr.go) enqueues,
// lpq (lpq.go) lists, and lpd (lpd.go) "prints" by draining the queue into an
// output directory. This file holds only the shared Command type and the
// name-based dispatch tables.
package printutils

import (
	"context"

	"github.com/nao1215/mimixbox/internal/command"
)

// defaultSpool is the spool directory used when -S/SPOOL is not given. Tests set
// the spool via the -S flag instead, so this stays a documented default rather
// than a hidden host write.
const defaultSpool = "/var/spool/mimixbox-lpd"

// Command is one print applet, distinguished by name.
type Command struct {
	name string
}

// Name returns the command name.
func (c *Command) Name() string { return c.name }

// NewLpr returns the lpr applet (enqueue a print job).
func NewLpr() *Command { return &Command{name: "lpr"} }

// NewLpq returns the lpq applet (list the print queue).
func NewLpq() *Command { return &Command{name: "lpq"} }

// NewLpd returns the lpd applet (drain the print queue).
func NewLpd() *Command { return &Command{name: "lpd"} }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string {
	switch c.name {
	case "lpr":
		return "Queue files for printing into a local spool"
	case "lpq":
		return "Show the local print queue"
	case "lpd":
		return "Drain the local print spool to an output directory"
	}
	return "Line-printer utility"
}

// Run dispatches to the per-applet implementation.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	switch c.name {
	case "lpr":
		return c.runLpr(stdio, args)
	case "lpq":
		return c.runLpq(stdio, args)
	case "lpd":
		return c.runLpd(stdio, args)
	}
	return command.Failuref("%s: unknown print applet", c.name)
}
