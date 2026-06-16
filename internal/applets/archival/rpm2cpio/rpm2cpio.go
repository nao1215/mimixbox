// Package rpm2cpio implements the rpm2cpio applet: extract the cpio payload
// from an RPM package and write it, decompressed, to standard output. The
// result can be piped straight into cpio (e.g. rpm2cpio pkg.rpm | cpio -idmv).
package rpm2cpio

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/nao1215/mimixbox/internal/applets/archival/rpmfile"
	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the rpm2cpio applet.
type Command struct{}

// New returns an rpm2cpio command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "rpm2cpio" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Extract the cpio payload from an RPM package" }

// Run executes rpm2cpio.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[FILE.rpm]", stdio.Err).WithHelp(command.Help{
		Description: "Extract the cpio payload from the RPM package FILE.rpm and write it, " +
			"decompressed, to standard output. With no FILE, or when FILE is '-', the package is " +
			"read from standard input. The result can be piped straight into cpio to unpack the " +
			"package contents.",
		Examples: []command.Example{
			{Command: "rpm2cpio pkg.rpm | cpio -idmv", Explain: "Extract every file from pkg.rpm into the current directory."},
			{Command: "rpm2cpio pkg.rpm > pkg.cpio", Explain: "Save the decompressed cpio payload to a file."},
			{Command: "cat pkg.rpm | rpm2cpio", Explain: "Read the package from standard input."},
		},
		ExitStatus: "0  the payload was extracted successfully.\n1  the file could not be opened or is not a valid RPM package.",
	})
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	r := stdio.In
	if rest := fs.Args(); len(rest) > 0 && rest[0] != "-" {
		f, oerr := os.Open(rest[0]) //nolint:gosec // user-named file
		if oerr != nil {
			_, _ = fmt.Fprintf(stdio.Err, "rpm2cpio: %s\n", command.FileError(rest[0], oerr))
			return command.SilentFailure()
		}
		defer func() { _ = f.Close() }()
		r = f
	}

	rpm, err := rpmfile.Open(r)
	if err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "rpm2cpio: %v\n", err)
		return command.SilentFailure()
	}
	payload, err := rpm.Payload()
	if err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "rpm2cpio: %v\n", err)
		return command.SilentFailure()
	}
	if _, err := io.Copy(stdio.Out, payload); err != nil { //nolint:gosec // decompressing user data
		_, _ = fmt.Fprintf(stdio.Err, "rpm2cpio: %v\n", err)
		return command.SilentFailure()
	}
	return nil
}
