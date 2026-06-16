package printutils

import (
	"fmt"
	"text/tabwriter"

	"github.com/nao1215/mimixbox/internal/command"
)

// runLpq lists the jobs currently queued in the spool backend.
func (c *Command) runLpq(stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.name, "[-S SPOOL]", stdio.Err).WithHelp(command.Help{
		Description: "List the jobs currently queued in the spool directory (-S SPOOL, default " +
			defaultSpool + "), in id order, showing rank, owner, job id, original file name, and size " +
			"in bytes. Reports 'no entries' when the queue is empty.",
		Examples:   []command.Example{{Command: "lpq -S /tmp/spool", Explain: "Show the queue in /tmp/spool."}},
		ExitStatus: "0  success (including an empty queue).\n1  the spool could not be read.",
	})
	spoolDir := fs.StringP("spool", "S", defaultSpool, "spool directory")
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}
	jobs, err := openSpool(*spoolDir).list()
	if err != nil {
		fmt.Fprintf(stdio.Err, "%s: %v\n", c.name, err)
		return command.SilentFailure()
	}
	if len(jobs) == 0 {
		fmt.Fprintln(stdio.Out, "no entries")
		return nil
	}
	tw := tabwriter.NewWriter(stdio.Out, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "Rank\tOwner\tJob\tFiles\tSize")
	for i, j := range jobs {
		fmt.Fprintf(tw, "%d\t%s\t%d\t%s\t%d\n", i+1, j.Owner, j.ID, j.Name, j.Size)
	}
	return tw.Flush()
}
