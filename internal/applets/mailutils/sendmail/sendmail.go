// Package sendmail implements the sendmail applet: read a mail message from
// standard input and deliver it locally by appending it to an mbox file. Network
// (SMTP) delivery is intentionally not implemented; this slice covers the local
// file workflow only.
package sendmail

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the sendmail applet.
type Command struct{}

// New returns a sendmail command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "sendmail" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Deliver a message to a local mbox file" }

// now is the clock used for the mbox "From " separator; tests override it.
var now = time.Now

// Run executes sendmail.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[-f FROM] -o MBOX [RECIPIENT]...", stdio.Err).WithHelp(command.Help{
		Description: "Read an RFC 5322 message from standard input and append it, in mbox format, to the " +
			"file named by -o. The message is prefixed with an mbox 'From ' separator line using the " +
			"-f envelope sender (default 'MAILER-DAEMON') and the current time. RECIPIENT operands are " +
			"accepted and recorded but, in this slice, no SMTP delivery is performed: delivery is to " +
			"the local mbox only. The '-t' (read recipients from headers) and '-i' flags are accepted " +
			"for compatibility.",
		Examples: []command.Example{
			{Command: "sendmail -o inbox.mbox -f me@example.com you@example.com < msg.eml", Explain: "Append msg.eml to inbox.mbox."},
		},
		ExitStatus: "0  the message was written to the mbox.\n1  no -o mbox was given or the file could not be written.",
		Notes: []string{
			"Network (SMTP) delivery is intentionally not implemented; only local mbox delivery is supported in this build.",
		},
	})
	from := fs.StringP("from", "f", "MAILER-DAEMON", "envelope sender used in the mbox 'From ' line")
	mbox := fs.StringP("output", "o", "", "mbox file to append the message to")
	fs.BoolP("read-recipients", "t", false, "read recipients from the message headers (accepted for compatibility)")
	fs.BoolP("ignore-dots", "i", false, "do not treat a line of a single dot as end of input (accepted for compatibility)")
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	if *mbox == "" {
		fmt.Fprintf(stdio.Err, "%s: no mbox given; use -o FILE (SMTP delivery is not implemented)\n", c.Name())
		return command.SilentFailure()
	}

	body, err := io.ReadAll(stdio.In)
	if err != nil {
		fmt.Fprintf(stdio.Err, "%s: %v\n", c.Name(), err)
		return command.SilentFailure()
	}

	if err := appendMbox(*mbox, *from, body); err != nil {
		fmt.Fprintf(stdio.Err, "%s: %v\n", c.Name(), err)
		return command.SilentFailure()
	}
	return nil
}

// appendMbox appends a message to an mbox file with a proper "From " separator
// and ">From " body escaping.
func appendMbox(path, from string, body []byte) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600) //nolint:gosec // user-named file
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	w := bufio.NewWriter(f)
	fmt.Fprintf(w, "From %s %s\n", from, now().UTC().Format("Mon Jan  2 15:04:05 2006"))
	sc := bufio.NewScanner(strings.NewReader(string(body)))
	sc.Buffer(make([]byte, 0, 64*1024), command.MaxLineSize)
	for sc.Scan() {
		line := sc.Text()
		if strings.HasPrefix(line, "From ") {
			_ = w.WriteByte('>')
		}
		_, _ = w.WriteString(line)
		_ = w.WriteByte('\n')
	}
	if err := sc.Err(); err != nil {
		return err
	}
	_ = w.WriteByte('\n') // blank line terminating the mbox entry
	return w.Flush()
}
