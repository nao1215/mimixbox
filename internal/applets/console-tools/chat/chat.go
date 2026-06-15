// Package chat implements the chat applet: run an expect/send conversation
// (a chat script) over a connection, as used to drive modems and serial links.
package chat

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the chat applet.
type Command struct{}

// New returns a chat command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "chat" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Run an expect/send conversation script" }

// Step is one expect/send pair from a chat script: wait until Expect is seen on
// the link, then write Send (followed by a carriage return unless suppressed).
type Step struct {
	Expect string
	Send   string
}

// ParseScript turns the chat script words into expect/send steps. The script is
// a sequence of alternating expect and send strings, so the first word is an
// expect, the second the reply to send, and so on. An empty expect ("") means
// "send immediately without waiting". A trailing expect with no reply waits for
// that final string. Common escapes (\r \n \t \\ \" and \c to suppress the
// trailing return) are interpreted.
func ParseScript(words []string) ([]Step, error) {
	var steps []Step
	for i := 0; i < len(words); i += 2 {
		expect, err := unescape(words[i])
		if err != nil {
			return nil, fmt.Errorf("in expect string %q: %w", words[i], err)
		}
		step := Step{Expect: expect}
		if i+1 < len(words) {
			send, err := unescape(words[i+1])
			if err != nil {
				return nil, fmt.Errorf("in send string %q: %w", words[i+1], err)
			}
			step.Send = send
		}
		steps = append(steps, step)
	}
	return steps, nil
}

// unescape interprets the backslash escapes chat understands. A literal "\c" is
// preserved as a sentinel so the caller can suppress the trailing return.
func unescape(s string) (string, error) {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] != '\\' {
			b.WriteByte(s[i])
			continue
		}
		i++
		if i >= len(s) {
			return "", fmt.Errorf("dangling backslash")
		}
		switch s[i] {
		case 'r':
			b.WriteByte('\r')
		case 'n':
			b.WriteByte('\n')
		case 't':
			b.WriteByte('\t')
		case '\\':
			b.WriteByte('\\')
		case '"':
			b.WriteByte('"')
		case 'c':
			b.WriteString("\\c") // sentinel: suppress trailing CR
		default:
			b.WriteByte(s[i])
		}
	}
	return b.String(), nil
}

// Conversation runs the steps against the link r/w. For each step it reads from
// r until Expect has been seen (or EOF / timeout), then writes Send. It returns
// an error if an expected string is never seen before the link closes or the
// deadline passes. timeout is the per-step expect timeout; a zero timeout means
// wait indefinitely (bounded only by EOF).
func Conversation(ctx context.Context, r io.Reader, w io.Writer, steps []Step, timeout time.Duration) error {
	br := bufio.NewReader(r)
	for n, step := range steps {
		if step.Expect != "" {
			if err := expect(ctx, br, step.Expect, timeout); err != nil {
				return fmt.Errorf("step %d: waiting for %q: %w", n+1, step.Expect, err)
			}
		}
		if err := send(w, step.Send); err != nil {
			return fmt.Errorf("step %d: sending: %w", n+1, err)
		}
	}
	return nil
}

// expect reads bytes from br until want has appeared as a substring of the
// stream tail, or the context/timeout/EOF ends the wait.
func expect(ctx context.Context, br *bufio.Reader, want string, timeout time.Duration) error {
	var deadline <-chan time.Time
	if timeout > 0 {
		t := time.NewTimer(timeout)
		defer t.Stop()
		deadline = t.C
	}
	var seen strings.Builder
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-deadline:
			return fmt.Errorf("timed out")
		default:
		}
		b, err := br.ReadByte()
		if err != nil {
			if err == io.EOF {
				return fmt.Errorf("connection closed before %q was seen", want)
			}
			return err
		}
		seen.WriteByte(b)
		if strings.Contains(seen.String(), want) {
			return nil
		}
	}
}

// send writes the reply, appending a carriage return unless the reply ended with
// the \c sentinel, which is then stripped.
func send(w io.Writer, reply string) error {
	if strings.HasSuffix(reply, "\\c") {
		_, err := io.WriteString(w, strings.TrimSuffix(reply, "\\c"))
		return err
	}
	_, err := io.WriteString(w, reply+"\r")
	return err
}

// Run executes chat.
func (c *Command) Run(ctx context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[-t TIMEOUT] EXPECT SEND [EXPECT SEND]...", stdio.Err).WithHelp(command.Help{
		Description: "Run an expect/send conversation over the connection on standard input and output. " +
			"The arguments form an alternating sequence of strings to expect from the link and replies " +
			"to send back; an empty expect (\"\") sends immediately, and replies get a trailing carriage " +
			"return unless they end with \\c. Backslash escapes \\r \\n \\t \\\\ \\\" \\c are understood. " +
			"With -t the timeout (in seconds) for each expect is set; the default waits until the link " +
			"closes. This is the modem-dialing chat: connect its standard input/output to a serial " +
			"device (or, for testing, a pipe) to drive the conversation.",
		Examples: []command.Example{
			{Command: "chat '' ATZ OK ATDT5551234 CONNECT ''", Explain: "Reset a modem and dial."},
			{Command: "chat -t 5 ogin: user word: secret", Explain: "Log in, waiting up to 5s per prompt."},
		},
		ExitStatus: "0  the whole script ran.\n" +
			"1  a bad script, or an expected string was never seen before timeout/EOF.",
	})
	timeoutSec := fs.IntP("timeout", "t", 0, "seconds to wait for each expected string (0 = until EOF)")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}
	rest := fs.Args()
	if len(rest) == 0 {
		return command.Failuref("a chat script (EXPECT SEND ...) is required")
	}
	steps, err := ParseScript(rest)
	if err != nil {
		return command.Failuref("%v", err)
	}

	timeout := time.Duration(*timeoutSec) * time.Second
	if err := Conversation(ctx, stdio.In, stdio.Out, steps, timeout); err != nil {
		return command.Failuref("%v", err)
	}
	return nil
}
