// Package cowsay implements the cowsay applet: print a message inside a speech
// bubble above a cow drawn in ASCII art. The message comes from the operands,
// or from standard input when no operands are given.
package cowsay

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

// cow is the ASCII art drawn below the speech bubble.
const cow = `   \ 
    \   ^__^
     \  (oo)\_______
        (__)\       )\/\
            ||----w |
            ||     ||`

// Command is the cowsay applet.
type Command struct{}

// New returns a cowsay command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "cowsay" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Print message with cow's ASCII art" }

// Run executes cowsay.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]... [MESSAGE]", stdio.Err)
	proceed, err := fs.Parse(stdio, args)
	if !proceed {
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

// readMessage reads the whole of stdio.In and returns it with any trailing
// newline removed, mirroring how a piped message is presented.
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

// render builds the full cowsay output: a speech bubble containing message
// (wrapped at bubbleWidth columns) above the cow ASCII art. The result ends
// with a trailing newline.
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

// wrap breaks src into lines of at most column bytes, joined by newlines. It
// reproduces the behavior of the original implementation's WrapString.
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
