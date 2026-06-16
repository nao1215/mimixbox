// Package cowthink implements the cowthink applet: print a message inside a
// thought bubble above a cow drawn in ASCII art. It is the thinking cousin of
// cowsay: the bubble is drawn with parentheses and the cow uses "o" connectors.
package cowthink

import (
	"bufio"
	"context"
	"fmt"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// bubbleWidth is the column at which the message is wrapped inside the bubble.
const bubbleWidth = 60

// border is the line drawn above and below the message to form the bubble.
const border = "------------------------------------------------------------"

// cow is the ASCII art drawn below the thought bubble; "o" connectors join the
// cow to the bubble, marking it as a thought rather than speech.
const cow = `   o
    o   ^__^
        (oo)\_______
        (__)\       )\/\
            ||----w |
            ||     ||`

// Command is the cowthink applet.
type Command struct{}

// New returns a cowthink command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "cowthink" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Print message in a cow's thought bubble" }

// Run executes cowthink.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]... [MESSAGE]", stdio.Err).WithHelp(command.Help{
		Description: "Print MESSAGE inside a thought bubble above a cow drawn in ASCII art. The message is taken " +
			"from the operands, or from standard input when no operands are given.",
		Examples: []command.Example{
			{Command: "cowthink Hmm, maybe", Explain: "Print \"Hmm, maybe\" in the cow's thought bubble."},
			{Command: "echo Moo | cowthink", Explain: "Read the message from standard input."},
		},
		ExitStatus: "0  success.\n1  the message could not be read or written.",
	})
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	var message string
	if operands := fs.Args(); len(operands) > 0 {
		message = strings.Join(operands, " ")
	} else {
		message, err = readMessage(stdio)
		if err != nil {
			return command.Failure(err)
		}
	}

	if _, err := fmt.Fprint(stdio.Out, render(message)); err != nil {
		return command.Failure(err)
	}
	return nil
}

// readMessage reads the whole of stdio.In and returns it joined into one line.
func readMessage(stdio command.IO) (string, error) {
	var b strings.Builder
	sc := bufio.NewScanner(stdio.In)
	for sc.Scan() {
		b.WriteString(sc.Text())
	}
	if err := sc.Err(); err != nil {
		return "", err
	}
	return b.String(), nil
}

// render builds the full cowthink output: a thought bubble containing message
// (wrapped at bubbleWidth columns) above the cow ASCII art.
func render(message string) string {
	var b strings.Builder
	b.WriteString(border)
	b.WriteByte('\n')
	b.WriteString(wrap(message, bubbleWidth))
	b.WriteByte('\n')
	b.WriteString(border)
	b.WriteByte('\n')
	b.WriteString(cow)
	b.WriteByte('\n')
	return b.String()
}

// wrap breaks src into lines of at most column bytes, joined by newlines.
func wrap(src string, column int) string {
	if column <= 0 {
		return src
	}
	var buf []string
	for i := 0; i < len(src); i += column {
		if i+column < len(src) {
			buf = append(buf, src[i:i+column])
		} else {
			buf = append(buf, src[i:])
		}
	}
	return strings.Join(buf, "\n")
}
