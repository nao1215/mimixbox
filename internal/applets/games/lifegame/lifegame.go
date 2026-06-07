// Package lifegame implements the lifegame applet: Conway's Game of Life.
//
// The simulation itself lives in the pure Board type so it can be unit-tested
// without a terminal. Run animates the board with termbox; when a terminal
// cannot be initialized (no TTY, as under tests or CI) it exits gracefully
// instead of crashing.
package lifegame

import (
	"context"
	"math/rand"
	"time"

	"github.com/nao1215/mimixbox/internal/command"
	termbox "github.com/nsf/termbox-go"
)

// Command is the lifegame applet.
type Command struct{}

// New returns a lifegame command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "lifegame" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Life game (Conway's Game of Life)" }

const (
	backGround = termbox.ColorBlack
	alive      = termbox.ColorWhite
	interval   = 100 * time.Millisecond
)

// Board is a Conway's Game of Life grid. It is independent of termbox so the
// simulation rules can be exercised by unit tests. cells is indexed
// [row][column]; a true value means the cell is alive.
type Board struct {
	width  int
	height int
	cells  [][]bool
}

// NewBoard returns an empty (all-dead) Board of the given size. Non-positive
// dimensions are clamped to zero.
func NewBoard(width, height int) *Board {
	if width < 0 {
		width = 0
	}
	if height < 0 {
		height = 0
	}
	cells := make([][]bool, height)
	for i := range cells {
		cells[i] = make([]bool, width)
	}
	return &Board{width: width, height: height, cells: cells}
}

// Width returns the board width in columns.
func (b *Board) Width() int { return b.width }

// Height returns the board height in rows.
func (b *Board) Height() int { return b.height }

// Set marks the cell at (row, column) alive or dead. Out-of-range coordinates
// are ignored.
func (b *Board) Set(row, column int, value bool) {
	if !b.inBounds(row, column) {
		return
	}
	b.cells[row][column] = value
}

// Alive reports whether the cell at (row, column) is alive. Cells outside the
// board are considered dead, which gives the board fixed (non-wrapping) edges.
func (b *Board) Alive(row, column int) bool {
	if !b.inBounds(row, column) {
		return false
	}
	return b.cells[row][column]
}

// inBounds reports whether (row, column) is inside the board.
func (b *Board) inBounds(row, column int) bool {
	return row >= 0 && column >= 0 && row < b.height && column < b.width
}

// neighbors counts the live cells in the eight cells surrounding (row, column).
func (b *Board) neighbors(row, column int) int {
	sum := 0
	for i := -1; i <= 1; i++ {
		for j := -1; j <= 1; j++ {
			if i == 0 && j == 0 {
				continue
			}
			if b.Alive(row+i, column+j) {
				sum++
			}
		}
	}
	return sum
}

// Next returns a new Board advanced one generation under Conway's rules:
//   - a dead cell with exactly 3 live neighbors becomes alive (birth);
//   - a live cell with 2 or 3 live neighbors stays alive (survival);
//   - any other live cell dies (under- or over-population).
func (b *Board) Next() *Board {
	next := NewBoard(b.width, b.height)
	for i := 0; i < b.height; i++ {
		for j := 0; j < b.width; j++ {
			n := b.neighbors(i, j)
			if b.cells[i][j] {
				next.cells[i][j] = n == 2 || n == 3
			} else {
				next.cells[i][j] = n == 3
			}
		}
	}
	return next
}

// Randomize fills the board with a random pattern; roughly one cell in twenty
// starts alive.
func (b *Board) Randomize(r *rand.Rand) {
	for i := 0; i < b.height; i++ {
		for j := 0; j < b.width; j++ {
			b.cells[i][j] = r.Intn(20) == 0
		}
	}
}

// Run executes lifegame.
func (c *Command) Run(ctx context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]", stdio.Err)

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	return c.animate(ctx)
}

// animate drives the simulation on the terminal using termbox. When a terminal
// cannot be initialized (no TTY, as under tests/CI), it returns nil so the
// command degrades gracefully instead of crashing.
func (c *Command) animate(ctx context.Context) error {
	if err := termbox.Init(); err != nil {
		// No real terminal (tests/CI): nothing to animate. Exit gracefully.
		return nil //nolint:nilerr // a missing terminal is not a failure for lifegame
	}
	defer termbox.Close()

	width, height := termbox.Size()
	if width <= 0 || height <= 0 {
		return nil
	}

	board := NewBoard(width, height)
	board.Randomize(rand.New(rand.NewSource(time.Now().UnixNano())))

	queue := pollEvent()
	for {
		select {
		case <-ctx.Done():
			return nil
		case ev := <-queue:
			if isGameEnd(ev.Key) {
				return nil
			}
		default:
			if err := render(board); err != nil {
				return nil
			}
			board = board.Next()
			time.Sleep(interval)
		}
	}
}

// pollEvent forwards termbox events on a channel so the main loop can poll for
// quit keys without blocking the animation.
func pollEvent() chan termbox.Event {
	q := make(chan termbox.Event)
	go func() {
		for {
			q <- termbox.PollEvent()
		}
	}()
	return q
}

// isGameEnd reports whether key is one of the quit keys (Esc, Ctrl-C, Ctrl-D).
func isGameEnd(key termbox.Key) bool {
	return key == termbox.KeyEsc || key == termbox.KeyCtrlD || key == termbox.KeyCtrlC
}

// render draws the board to the terminal. It returns an error if termbox fails
// to clear or flush, so the caller can stop the loop.
func render(b *Board) error {
	if err := termbox.Clear(backGround, backGround); err != nil {
		return err
	}
	for i := 0; i < b.height; i++ {
		for j := 0; j < b.width; j++ {
			if b.cells[i][j] {
				termbox.SetCell(j, i, '█', alive, backGround)
			}
		}
	}
	return termbox.Flush()
}
