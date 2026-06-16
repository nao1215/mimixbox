package lifegame_test

import (
	"bytes"
	"context"
	"math/rand"
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

// TestNewBoardClampsNegativeDimensions checks that negative width/height are
// clamped to zero rather than allocating a negative-length slice.
func TestNewBoardClampsNegativeDimensions(t *testing.T) {
	t.Parallel()
	b := lifegame.NewBoard(-5, -3)
	if b.Width() != 0 || b.Height() != 0 {
		t.Errorf("dimensions = %dx%d, want 0x0", b.Width(), b.Height())
	}
}

// TestSetOutOfBoundsIgnored verifies that Set silently ignores out-of-range
// coordinates and never panics.
func TestSetOutOfBoundsIgnored(t *testing.T) {
	t.Parallel()
	b := lifegame.NewBoard(3, 3)
	b.Set(-1, 0, true)
	b.Set(0, -1, true)
	b.Set(3, 0, true)
	b.Set(0, 3, true)
	if len(alivePoints(b)) != 0 {
		t.Errorf("out-of-bounds Set changed board: %v", alivePoints(b))
	}
}

// TestAliveOutOfBoundsIsDead checks the non-wrapping edge behavior: cells
// outside the board read as dead.
func TestAliveOutOfBoundsIsDead(t *testing.T) {
	t.Parallel()
	b := lifegame.NewBoard(2, 2)
	if b.Alive(-1, 0) || b.Alive(0, -1) || b.Alive(2, 0) || b.Alive(0, 2) {
		t.Error("out-of-bounds cell reported alive, want dead")
	}
}

// TestRandomizeIsDeterministicForSeed checks Randomize: the same seed produces
// the same pattern, and a different seed produces a different one, confirming
// it actually reads from the supplied source.
func TestRandomizeIsDeterministicForSeed(t *testing.T) {
	t.Parallel()
	a := lifegame.NewBoard(20, 20)
	a.Randomize(rand.New(rand.NewSource(42)))

	b := lifegame.NewBoard(20, 20)
	b.Randomize(rand.New(rand.NewSource(42)))

	if !equalPoints(alivePoints(a), pointsSlice(alivePoints(b))) {
		t.Error("same seed produced different boards")
	}

	c := lifegame.NewBoard(20, 20)
	c.Randomize(rand.New(rand.NewSource(7)))
	if equalPoints(alivePoints(a), pointsSlice(alivePoints(c))) {
		t.Error("different seeds produced identical boards (unlikely)")
	}
}

// pointsSlice converts an alive set into the [][2]int form equalPoints expects.
func pointsSlice(set map[[2]int]bool) [][2]int {
	out := make([][2]int, 0, len(set))
	for p := range set {
		out = append(out, p)
	}
	return out
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
