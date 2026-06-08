// Package nyancat implements the nyancat applet: animate the Nyan Cat trailing
// a rainbow across the terminal. The animation needs a real terminal; without
// one (tests/CI) it exits gracefully.
package nyancat

import (
	"context"
	"strings"
	"time"

	"github.com/nao1215/mimixbox/internal/command"
	termbox "github.com/nsf/termbox-go"
)

// Command is the nyancat applet.
type Command struct{}

// New returns a nyancat command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "nyancat" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Animate the rainbow-trailing Nyan Cat" }

// cat holds the ASCII-art rows of Nyan Cat. Keeping it as package data lets a
// test assert the art is present without initializing a terminal.
var cat = []string{
	"  ,---/V\\ ",
	" ~|__(o.o)",
	"   UU  UU ",
}

// Frame returns one animation frame: a rainbow trail of the given length drawn
// to the left of the cat, as a multi-line string. A negative length is treated
// as zero.
func Frame(trail int) string {
	if trail < 0 {
		trail = 0
	}
	rainbow := strings.Repeat("=", trail)
	var b strings.Builder
	for i, row := range cat {
		if i > 0 {
			b.WriteByte('\n')
		}
		// The trail streams from the cat's middle row.
		if i == 1 {
			b.WriteString(rainbow)
		} else {
			b.WriteString(strings.Repeat(" ", trail))
		}
		b.WriteString(row)
	}
	return b.String()
}

// Run executes nyancat.
func (c *Command) Run(ctx context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]", stdio.Err)

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	return c.animate(ctx)
}

// animate streams the rainbow trail across the terminal using termbox. Without
// a terminal it returns nil so the command degrades gracefully.
func (c *Command) animate(ctx context.Context) error {
	if err := termbox.Init(); err != nil {
		return nil //nolint:nilerr // a missing terminal is not a failure
	}
	defer termbox.Close()

	width, height := termbox.Size()
	if width <= 0 || height <= 0 {
		return nil
	}
	top := (height - len(cat)) / 2
	if top < 0 {
		top = 0
	}

	for trail := 0; ; trail = (trail + 1) % width {
		select {
		case <-ctx.Done():
			return nil
		default:
		}
		if err := termbox.Clear(termbox.ColorDefault, termbox.ColorDefault); err != nil {
			return nil
		}
		for i, line := range strings.Split(Frame(trail), "\n") {
			drawString(0, top+i, line)
		}
		if err := termbox.Flush(); err != nil {
			return nil
		}
		time.Sleep(time.Second / 15)
	}
}

// drawString draws s starting at column x, row y, skipping cells off-screen.
func drawString(x, y int, s string) {
	for _, r := range s {
		if x >= 0 {
			termbox.SetCell(x, y, r, termbox.ColorDefault, termbox.ColorDefault)
		}
		x++
	}
}
