package sl

import "testing"

// TestLongestRow verifies longestRow returns the width of the widest train row.
// It is the pure helper that drives how far the animation must scroll, so it can
// be asserted without a terminal. drawString and animate require an initialized
// termbox terminal (a real TTY) and so are not unit-testable here.
func TestLongestRow(t *testing.T) {
	got := longestRow()

	// Compute the expected width directly from the package art so the test
	// stays correct if the locomotive art changes.
	want := 0
	for _, row := range train {
		if n := len([]rune(row)); n > want {
			want = n
		}
	}

	if got != want {
		t.Errorf("longestRow() = %d, want %d", got, want)
	}
	if got <= 0 {
		t.Errorf("longestRow() = %d, want a positive width", got)
	}
	// Every row must fit inside the reported longest width.
	for i, row := range train {
		if n := len([]rune(row)); n > got {
			t.Errorf("row %d has width %d > longestRow() %d", i, n, got)
		}
	}
}
