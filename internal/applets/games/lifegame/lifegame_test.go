package lifegame_test

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/nao1215/mimixbox/internal/applets/games/lifegame"
	"github.com/nao1215/mimixbox/internal/command"
)

func TestNew(t *testing.T) {
	t.Parallel()
	if lifegame.New() == nil {
		t.Fatal("New() returned nil")
	}
}

func TestName(t *testing.T) {
	t.Parallel()
	if got := lifegame.New().Name(); got != "lifegame" {
		t.Errorf("Name() = %q, want %q", got, "lifegame")
	}
}

func TestSynopsis(t *testing.T) {
	t.Parallel()
	want := "Life game (Conway's Game of Life)"
	if got := lifegame.New().Synopsis(); got != want {
		t.Errorf("Synopsis() = %q, want %q", got, want)
	}
}

// alivePoints returns the set of live coordinates on the board as "row,col"
// strings so two generations can be compared by content.
func alivePoints(b *lifegame.Board) map[[2]int]bool {
	set := map[[2]int]bool{}
	for i := 0; i < b.Height(); i++ {
		for j := 0; j < b.Width(); j++ {
			if b.Alive(i, j) {
				set[[2]int{i, j}] = true
			}
		}
	}
	return set
}

func setCells(b *lifegame.Board, points [][2]int) {
	for _, p := range points {
		b.Set(p[0], p[1], true)
	}
}

func equalPoints(got map[[2]int]bool, want [][2]int) bool {
	if len(got) != len(want) {
		return false
	}
	for _, p := range want {
		if !got[p] {
			return false
		}
	}
	return true
}

// TestBlinkerOscillates checks the classic period-2 oscillator: a horizontal
// row of three flips to a vertical column of three and back again.
func TestBlinkerOscillates(t *testing.T) {
	t.Parallel()

	// Horizontal blinker centered in a 5x5 board.
	b := lifegame.NewBoard(5, 5)
	setCells(b, [][2]int{{2, 1}, {2, 2}, {2, 3}})

	vertical := [][2]int{{1, 2}, {2, 2}, {3, 2}}
	horizontal := [][2]int{{2, 1}, {2, 2}, {2, 3}}

	gen1 := b.Next()
	if !equalPoints(alivePoints(gen1), vertical) {
		t.Fatalf("after 1 step, alive = %v, want vertical %v", alivePoints(gen1), vertical)
	}

	gen2 := gen1.Next()
	if !equalPoints(alivePoints(gen2), horizontal) {
		t.Fatalf("after 2 steps, alive = %v, want horizontal %v", alivePoints(gen2), horizontal)
	}
}

// TestBlockStable checks the 2x2 still life: it must not change.
func TestBlockStable(t *testing.T) {
	t.Parallel()

	b := lifegame.NewBoard(4, 4)
	block := [][2]int{{1, 1}, {1, 2}, {2, 1}, {2, 2}}
	setCells(b, block)

	next := b.Next()
	if !equalPoints(alivePoints(next), block) {
		t.Errorf("block changed: alive = %v, want %v", alivePoints(next), block)
	}
}

// TestLoneCellDies checks under-population: a single live cell with no
// neighbors dies in the next generation.
func TestLoneCellDies(t *testing.T) {
	t.Parallel()

	b := lifegame.NewBoard(3, 3)
	b.Set(1, 1, true)

	next := b.Next()
	if len(alivePoints(next)) != 0 {
		t.Errorf("lone cell survived: alive = %v", alivePoints(next))
	}
}

// TestRunNonTTY verifies that Run returns promptly without error when there is
// no terminal (as under tests/CI), and does not hang.
func TestRunNonTTY(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: errBuf}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- lifegame.New().Run(ctx, io, nil)
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Run error = %v", err)
		}
	case <-ctx.Done():
		t.Fatal("Run did not return; it hung")
	}
}
