// Package sl implements the sl applet: it animates a steam locomotive across
// the terminal to cure the bad habit of mistyping "ls". The animation needs a
// real terminal; when one is not available (for example under tests or CI) the
// command exits gracefully without doing anything.
package sl

import (
	"context"
	"strings"
	"time"

	"github.com/nao1215/mimixbox/internal/command"
	termbox "github.com/nsf/termbox-go"
)

// Command is the sl applet.
type Command struct{}

// New returns an sl command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "sl" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Cure your bad habit of mistyping" }

// train holds the ASCII-art rows of the steam locomotive. Keeping it as package
// data lets a test assert the art is present without initializing a terminal.
var train = []string{
	"      ====        ________                ___________ ",
	"  _D _|  |_______/        \\__I_I_____===__|_________| ",
	"   |(_)---  |   H\\________/ |   |        =|___ ___|      _________________         ",
	"   /     |  |   H  |  |     |   |         ||_| |_||     _|                \\_____A  ",
	"  |      |  |   H  |__--------------------| [___] |   =|                        |  ",
	"  | ________|___H__/__|_____/[ ][ ]\\_______|       |   -|      ʕ ◔ ϖ ◔ ʔ      |  ",
	"  |/ |   |-----------I_____I [ ][ ] [ ]  D   |=======|____|_______________________|_ ",
	"__/ =| o |=-O=====O=====O=====O \\ ____Y___________|__|__________________________|_ ",
	" |/-=|___|=    ||    ||    ||    |_____/~\\___/          |_D__D__D_|  |_D__D__D_|   ",
	"  \\_/      \\__/  \\__/  \\__/  \\__/      \\_/               \\_/   \\_/    \\_/   \\_/    ",
}

// TrainFrame returns the steam locomotive ASCII art shifted right by offset
// columns. It is a pure helper so the animation can be exercised (and the art
// asserted) without a terminal. A negative offset is treated as zero.
func TrainFrame(offset int) string {
	if offset < 0 {
		offset = 0
	}
	pad := strings.Repeat(" ", offset)
	var b strings.Builder
	for i, row := range train {
		if i > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(pad)
		b.WriteString(row)
	}
	return b.String()
}

// Run executes sl.
func (c *Command) Run(ctx context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]", stdio.Err)

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	return c.animate(ctx)
}

// animate drives the steam locomotive across the terminal using termbox. When a
// terminal cannot be initialized (no TTY, as under tests/CI), it returns nil so
// the command degrades gracefully instead of crashing.
func (c *Command) animate(ctx context.Context) error {
	if err := termbox.Init(); err != nil {
		// No real terminal (tests/CI): nothing to animate. Exit gracefully.
		return nil //nolint:nilerr // a missing terminal is not a failure for sl
	}
	defer termbox.Close()

	width, height := termbox.Size()
	if width <= 0 || height <= 0 {
		return nil
	}

	// The train enters from the right edge and travels off the left edge.
	for x := width; x > -longestRow(); x-- {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
		top := (height - len(train)) / 2
		if top < 0 {
			top = 0
		}
		for i, row := range train {
			drawString(x, top+i, row)
		}
		if err := termbox.Flush(); err != nil {
			return nil
		}
		time.Sleep(time.Second / 30)
	}
	return nil
}

// longestRow returns the width of the widest train row.
func longestRow() int {
	max := 0
	for _, row := range train {
		if n := len([]rune(row)); n > max {
			max = n
		}
	}
	return max
}

// drawString draws s starting at column x, row y. Cells outside the terminal
// are skipped so the train can scroll in and out of view.
func drawString(x, y int, s string) {
	for _, r := range s {
		if x >= 0 {
			termbox.SetCell(x, y, r, termbox.ColorDefault, termbox.ColorDefault)
		}
		x++
	}
}
