package printutils

import (
	"fmt"
	"os"

	"github.com/nao1215/mimixbox/internal/command"
)

// runLpd drains the spool backend into an output directory, "printing" each job.
func (c *Command) runLpd(stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.name, "[-S SPOOL] -o OUTDIR", stdio.Err).WithHelp(command.Help{
		Description: "Drain the print spool (-S SPOOL, default " + defaultSpool + ") by 'printing' each " +
			"queued job: its data is written into the OUTDIR output directory (-o, created if needed) " +
			"under the job's id and original name, and the job is then removed from the queue. Jobs are " +
			"processed in id order. This is a local, file-based stand-in for the printer daemon; it does " +
			"not open a network socket.",
		Examples: []command.Example{
			{Command: "lpd -S /tmp/spool -o /tmp/printed", Explain: "Print every queued job into /tmp/printed."},
		},
		ExitStatus: "0  the queue was drained.\n1  the spool or output directory could not be accessed.",
		Notes: []string{
			"Network listening is intentionally not implemented; lpd here drains the local spool to a directory.",
		},
	})
	spoolDir := fs.StringP("spool", "S", defaultSpool, "spool directory")
	out := fs.StringP("output", "o", "", "directory to write printed jobs into")
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}
	if *out == "" {
		fmt.Fprintf(stdio.Err, "%s: no output directory; use -o DIR (network printing is not implemented)\n", c.name)
		return command.SilentFailure()
	}
	s := openSpool(*spoolDir)
	jobs, err := s.list()
	if err != nil {
		fmt.Fprintf(stdio.Err, "%s: %v\n", c.name, err)
		return command.SilentFailure()
	}
	if err := os.MkdirAll(*out, 0o700); err != nil {
		fmt.Fprintf(stdio.Err, "%s: %v\n", c.name, err)
		return command.SilentFailure()
	}
	var failed bool
	for _, j := range jobs {
		if err := s.print(*out, j); err != nil {
			fmt.Fprintf(stdio.Err, "%s: %v\n", c.name, err)
			failed = true
			continue
		}
		fmt.Fprintf(stdio.Out, "printed job %d (%s)\n", j.ID, j.Name)
	}
	if failed {
		return command.SilentFailure()
	}
	return nil
}
