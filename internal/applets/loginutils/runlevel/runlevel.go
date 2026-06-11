// Package runlevel implements the runlevel applet: print the previous and
// current system runlevel, read from the utmp run-level record.
package runlevel

import (
	"context"
	"encoding/binary"
	"fmt"
	"os"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the runlevel applet.
type Command struct{}

// New returns a runlevel command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "runlevel" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Print the previous and current runlevel" }

// Linux utmp layout (x86_64): 384-byte records. The run-level record stores the
// two runlevels in the ut_pid field as the current and previous ASCII bytes.
const (
	recordSize = 384
	typeOffset = 0
	pidOffset  = 4
	runLevel   = 1 // RUN_LVL
)

// utmpPath is the utmp database; tests point it at a fixture.
var utmpPath = "/var/run/utmp"

// Run executes runlevel.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "", stdio.Err).WithHelp(command.Help{
		Description: "Print the previous and current system runlevel, separated by a space, read from " +
			"the run-level record in the utmp database. The previous runlevel is 'N' if the system " +
			"has not changed runlevel since boot. Prints 'unknown' if no run-level record is found.",
		Examples: []command.Example{
			{Command: "runlevel", Explain: "Print e.g. 'N 5'."},
		},
		ExitStatus: "0  a runlevel was printed.\n1  no run-level record was found.",
	})
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	prev, cur, ok := readRunlevel()
	if !ok {
		_, _ = fmt.Fprintln(stdio.Out, "unknown")
		return command.SilentFailure()
	}
	prevStr := "N"
	if prev != 0 {
		prevStr = string(rune(prev))
	}
	_, _ = fmt.Fprintf(stdio.Out, "%s %c\n", prevStr, rune(cur))
	return nil
}

// readRunlevel returns the previous and current runlevel bytes from the utmp
// run-level record, and whether one was found.
func readRunlevel() (prev, cur byte, ok bool) {
	data, err := os.ReadFile(utmpPath)
	if err != nil {
		return 0, 0, false
	}
	for off := 0; off+recordSize <= len(data); off += recordSize {
		rec := data[off : off+recordSize]
		if binary.LittleEndian.Uint16(rec[typeOffset:]) != runLevel {
			continue
		}
		pid := binary.LittleEndian.Uint32(rec[pidOffset:])
		return byte(pid >> 8), byte(pid & 0xFF), true
	}
	return 0, 0, false
}
