package printutils

import (
	"fmt"

	"github.com/nao1215/mimixbox/internal/command"
)

// runLpr enqueues each FILE (or stdin) into the spool backend.
func (c *Command) runLpr(stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.name, "[-S SPOOL] [FILE]...", stdio.Err).WithHelp(command.Help{
		Description: "Queue files for printing by copying them into the spool directory (-S SPOOL, " +
			"default " + defaultSpool + ") and writing a control file for each job. With no FILE, or " +
			"'-', the job body is read from standard input. Each job is assigned an increasing numeric " +
			"id; use lpq to list the queue and lpd to drain it. No real printer is contacted.",
		Examples: []command.Example{
			{Command: "lpr -S /tmp/spool document.txt", Explain: "Queue document.txt into /tmp/spool."},
			{Command: "echo hi | lpr -S /tmp/spool", Explain: "Queue standard input."},
		},
		ExitStatus: "0  the job was queued.\n1  the spool or a file could not be accessed.",
	})
	spoolDir := fs.StringP("spool", "S", defaultSpool, "spool directory")
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	s := openSpool(*spoolDir)
	if err := s.init(); err != nil {
		fmt.Fprintf(stdio.Err, "%s: %v\n", c.name, err)
		return command.SilentFailure()
	}

	files := fs.Args()
	if len(files) == 0 {
		files = []string{"-"}
	}
	var failed bool
	for _, f := range files {
		if err := s.enqueue(f, stdio.In); err != nil {
			fmt.Fprintf(stdio.Err, "%s: %v\n", c.name, err)
			failed = true
		}
	}
	if failed {
		return command.SilentFailure()
	}
	return nil
}
