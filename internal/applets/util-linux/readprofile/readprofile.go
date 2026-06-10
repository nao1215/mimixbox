// Package readprofile implements the readprofile applet: report a summary of the
// kernel profiling buffer (/proc/profile).
package readprofile

import (
	"context"
	"encoding/binary"
	"fmt"
	"os"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the readprofile applet.
type Command struct{}

// New returns a readprofile command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "readprofile" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Summarize the kernel profiling buffer" }

// profilePath is the kernel profiling buffer; tests point it at a fixture.
var profilePath = "/proc/profile"

// Run executes readprofile.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "", stdio.Err).WithHelp(command.Help{
		Description: "Read the kernel profiling buffer (/proc/profile) and report its profiling step " +
			"and the total number of samples. Per-symbol attribution (which needs the kernel symbol " +
			"map) is not implemented, and the buffer is empty unless the kernel was booted with " +
			"profiling enabled.",
		Examples: []command.Example{
			{Command: "readprofile", Explain: "Show the profiling step and total samples."},
		},
		ExitStatus: "0  success.\n1  the profiling buffer could not be read or is empty.",
	})
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	data, err := os.ReadFile(profilePath) //nolint:gosec // the profiling buffer path
	if err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "readprofile: %s\n", command.FileError(profilePath, err))
		return command.SilentFailure()
	}
	if len(data) < 4 {
		_, _ = fmt.Fprintln(stdio.Err, "readprofile: profiling is not enabled (buffer is empty)")
		return command.SilentFailure()
	}

	step, total := summarize(data)
	_, _ = fmt.Fprintf(stdio.Out, "profiling step: %d\n", step)
	_, _ = fmt.Fprintf(stdio.Out, "total samples: %d\n", total)
	return nil
}

// summarize reads /proc/profile as a uint32 array: the first word is the
// profiling step and the rest are per-bucket sample counts.
func summarize(data []byte) (step uint32, total uint64) {
	step = binary.LittleEndian.Uint32(data[:4])
	for off := 4; off+4 <= len(data); off += 4 {
		total += uint64(binary.LittleEndian.Uint32(data[off : off+4]))
	}
	return step, total
}
