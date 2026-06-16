// Package cmatrix implements the cmatrix applet: show the falling-glyph "digital
// rain" effect. The animation needs a real terminal; without one (tests/CI) it
// exits gracefully.
package cmatrix

import (
	"context"
	"math/rand"
	"strings"
	"time"

	"github.com/nao1215/mimixbox/internal/command"
	termbox "github.com/nsf/termbox-go"
)

// Command is the cmatrix applet.
type Command struct{}

// New returns a cmatrix command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "cmatrix" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Show the falling-glyph digital rain effect" }

// glyphs is the alphabet the rain is drawn from.
const glyphs = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789@#$%&"

// trailLen is how many glyphs trail behind each drop head.
const trailLen = 6

// advance moves each drop head down one row, wrapping back to the top once a
// head has fallen trailLen rows past the bottom. next supplies the restart row.
func advance(heads []int, height int, next func() int) []int {
	out := make([]int, len(heads))
	for i, h := range heads {
		if h > height+trailLen {
			out[i] = next()
		} else {
			out[i] = h + 1
		}
	}
	return out
}

// RenderFrame draws one frame: for every column, the trailLen glyphs ending at
// the head row are lit. glyph supplies the rune for a given column and row so
// the renderer stays deterministic under test.
func RenderFrame(width, height int, heads []int, glyph func(col, row int) rune) string {
	grid := make([][]rune, height)
	for r := range grid {
		grid[r] = []rune(strings.Repeat(" ", width))
	}
	for col := 0; col < width && col < len(heads); col++ {
		head := heads[col]
		for t := 0; t < trailLen; t++ {
			row := head - t
			if row >= 0 && row < height {
				grid[row][col] = glyph(col, row)
			}
		}
	}
	var b strings.Builder
	for r := 0; r < height; r++ {
		b.WriteString(strings.TrimRight(string(grid[r]), " "))
		b.WriteByte('\n')
	}
	return b.String()
}

// Run executes cmatrix.
func (c *Command) Run(ctx context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]", stdio.Err).WithHelp(command.Help{
		Description: "Show the falling-glyph \"digital rain\" animation in the terminal. The effect needs a real " +
			"terminal; without one (for example under tests or CI) it exits gracefully. Press Ctrl-C to stop.",
		Examples: []command.Example{
			{Command: "cmatrix", Explain: "Run the digital rain animation until interrupted."},
		},
		ExitStatus: "0  the animation exited cleanly (or no terminal was available).",
	})

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	return c.animate(ctx)
}

// animate drives the rain with termbox. Without a terminal it returns nil so the
// command degrades gracefully.
func (c *Command) animate(ctx context.Context) error {
	if err := termbox.Init(); err != nil {
		return nil //nolint:nilerr // a missing terminal is not a failure
	}
	defer termbox.Close()

	width, height := termbox.Size()
	if width <= 0 || height <= 0 {
		return nil
	}

	heads := make([]int, width)
	for i := range heads {
		heads[i] = rand.Intn(height) //nolint:gosec // rain need not be cryptographically random
	}
	restart := func() int { return -rand.Intn(height) } //nolint:gosec // see above

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}
		if err := termbox.Clear(termbox.ColorDefault, termbox.ColorDefault); err != nil {
			return nil
		}
		for col := 0; col < width; col++ {
			head := heads[col]
			for t := 0; t < trailLen; t++ {
				row := head - t
				if row >= 0 && row < height {
					termbox.SetCell(col, row, randomGlyph(), termbox.ColorGreen, termbox.ColorDefault)
				}
			}
		}
		if err := termbox.Flush(); err != nil {
			return nil
		}
		heads = advance(heads, height, restart)
		time.Sleep(time.Second / 20)
	}
}

// randomGlyph returns a random rune from the rain alphabet.
func randomGlyph() rune {
	return rune(glyphs[rand.Intn(len(glyphs))]) //nolint:gosec // decorative randomness
}
