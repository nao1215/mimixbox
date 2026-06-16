package vi

import (
	"bytes"
	"testing"
)

// TestWordMotionsAcrossLines exercises the line-wrapping branches of the word
// motions, which the single-line cases in editor_test.go leave uncovered.
func TestWordMotionsAcrossLines(t *testing.T) {
	t.Parallel()

	// w from the last word of a line wraps to the first word of the next line.
	e := drive("ab cd\nef gh", "ww") // ab -> cd (col 3) -> wraps to ef on line 1
	if e.cy != 1 || e.cx != 0 {
		t.Errorf("ww across line -> (%d,%d), want (1,0)", e.cy, e.cx)
	}

	// w on the final word of the final line clamps at the last column.
	e = drive("solo", "w")
	if e.cy != 0 || e.cx != lastCol("solo") {
		t.Errorf("w at end of buffer -> (%d,%d), want (0,%d)", e.cy, e.cx, lastCol("solo"))
	}

	// e wraps onto the next line and lands on that word's end.
	e = drive("ab\ncde", "ee") // end of ab -> end of cde
	if e.cy != 1 || e.cx != lastCol("cde") {
		t.Errorf("ee across line -> (%d,%d), want (1,%d)", e.cy, e.cx, lastCol("cde"))
	}

	// b at column 0 wraps to the previous line.
	e = drive("foo\nbar", "jb") // line 1 col 0, then b wraps to line 0
	if e.cy != 0 {
		t.Errorf("b from start of line -> cy=%d, want 0", e.cy)
	}

	// b at the very start of the buffer is a no-op.
	e = drive("foo", "b")
	if e.cy != 0 || e.cx != 0 {
		t.Errorf("b at buffer start -> (%d,%d), want (0,0)", e.cy, e.cx)
	}
}

// TestExCommandX covers the ":x" alias (write+quit) and the bare ":q" on a
// clean buffer (quit without the dirty warning).
func TestExCommandXAndCleanQuit(t *testing.T) {
	t.Parallel()

	e := drive("data", ":x\r")
	if !e.save || !e.quit {
		t.Errorf(":x save=%v quit=%v, want both true", e.save, e.quit)
	}

	e = drive("data", ":q\r") // clean buffer: q quits with no message
	if !e.quit {
		t.Error(":q on a clean buffer should quit")
	}
	if e.message != "" {
		t.Errorf(":q on a clean buffer should not warn, message=%q", e.message)
	}
}

// TestExCommandUnknownIsNoop verifies an unrecognized ex command neither saves
// nor quits.
func TestExCommandUnknownIsNoop(t *testing.T) {
	t.Parallel()
	e := drive("data", ":zzz\r")
	if e.save || e.quit {
		t.Errorf("unknown ex command changed state: save=%v quit=%v", e.save, e.quit)
	}
	if e.mode != modeNormal {
		t.Errorf("Enter should leave command mode, mode=%s", modeName(e.mode))
	}
}

// TestBackspaceAtBufferStartIsNoop covers the early-return branch of backspace
// when the cursor is at row 0, column 0.
func TestBackspaceAtBufferStartIsNoop(t *testing.T) {
	t.Parallel()
	e := drive("abc", "i\x7f") // insert at (0,0), backspace
	if e.lines[0] != "abc" {
		t.Errorf("backspace at start mutated buffer: %q", e.lines[0])
	}
}

// TestDeleteLineLastLineMovesCursorUp covers the branch in deleteLine where the
// cursor row exceeds the new line count after deleting the last line.
func TestDeleteLineLastLineMovesCursorUp(t *testing.T) {
	t.Parallel()
	e := drive("l1\nl2\nl3", "Gdd") // go to last line, delete it
	if e.content() != "l1\nl2\n" {
		t.Errorf("Gdd -> %q, want %q", e.content(), "l1\nl2\n")
	}
	if e.cy != 1 {
		t.Errorf("after deleting last line cy=%d, want 1", e.cy)
	}
}

// TestSearchBackspaceEditsPattern drives the backspace branch of feedSearch.
func TestSearchBackspaceEditsPattern(t *testing.T) {
	t.Parallel()
	// Type "/betX", backspace removes X, leaving "bet" which still matches "beta".
	e := drive("alpha\nbeta", "/betX\x7f\r")
	if e.cy != 1 {
		t.Errorf("search after backspace -> cy=%d, want 1", e.cy)
	}
}

// TestSearchEmptyPatternIsNoop covers the empty-pattern early return of
// searchExecute (pressing Enter on an empty "/" prompt).
func TestSearchEmptyPatternIsNoop(t *testing.T) {
	t.Parallel()
	e := drive("alpha\nbeta", "/\r")
	if e.cy != 0 || e.cx != 0 {
		t.Errorf("empty search moved cursor to (%d,%d), want (0,0)", e.cy, e.cx)
	}
}

// TestSearchAgainWithoutPriorSearch covers searchAgain's guard when no search
// has run yet (n / N before any /).
func TestSearchAgainWithoutPriorSearch(t *testing.T) {
	t.Parallel()
	e := drive("alpha\nbeta", "n")
	if e.cy != 0 || e.cx != 0 {
		t.Errorf("n with no prior search moved cursor to (%d,%d)", e.cy, e.cx)
	}
}

// TestNCommandRepeatsSearchForward exercises the n/N path that reuses
// lastForward through searchAgain after a real search.
func TestSearchNAndCapitalN(t *testing.T) {
	t.Parallel()
	e := drive("foo\nbar\nfoo\nbaz\nfoo", "/foo\r") // lands on line 2 (first match after cursor)
	if e.cy != 2 {
		t.Fatalf("/foo -> cy=%d, want 2", e.cy)
	}
	e.feedString("n") // next match -> line 4
	if e.cy != 4 {
		t.Errorf("n -> cy=%d, want 4", e.cy)
	}
	e.feedString("N") // previous match -> back to line 2
	if e.cy != 2 {
		t.Errorf("N -> cy=%d, want 2", e.cy)
	}
}

// TestPasteWithEmptyRegisterIsNoop covers the empty-register guards of
// pasteBelow and pasteAbove (p / P before any yank).
func TestPasteWithEmptyRegisterIsNoop(t *testing.T) {
	t.Parallel()
	if e := drive("a\nb", "p"); e.content() != "a\nb\n" {
		t.Errorf("p with empty register changed buffer: %q", e.content())
	}
	if e := drive("a\nb", "P"); e.content() != "a\nb\n" {
		t.Errorf("P with empty register changed buffer: %q", e.content())
	}
}

// TestUndoWithEmptyStackIsNoop covers undo's guard when nothing has been
// recorded.
func TestUndoWithEmptyStackIsNoop(t *testing.T) {
	t.Parallel()
	e := drive("abc", "u")
	if e.content() != "abc\n" || e.dirty {
		t.Errorf("u with empty undo stack changed state: content=%q dirty=%v", e.content(), e.dirty)
	}
}

// TestYankPasteAbove exercises pasteAbove with a multi-line register.
func TestYankMultiPasteAbove(t *testing.T) {
	t.Parallel()
	e := drive("l1\nl2\nl3", "2yyjP") // yank l1,l2; move down; paste above line 2
	want := "l1\nl1\nl2\nl2\nl3\n"
	if e.content() != want {
		t.Errorf("2yy j P -> %q, want %q", e.content(), want)
	}
}

// TestGotoLineWithCount covers the counted-G branch (jump to a specific line).
func TestGotoLineWithCount(t *testing.T) {
	t.Parallel()
	e := drive("a\nb\nc\nd\ne", "3G")
	if e.cy != 2 {
		t.Errorf("3G -> cy=%d, want 2 (1-based line 3)", e.cy)
	}
	// A count past the end clamps to the last line.
	e = drive("a\nb\nc", "9G")
	if e.cy != 2 {
		t.Errorf("9G -> cy=%d, want clamped to 2", e.cy)
	}
}

// TestModeNameAllBranches covers modeName for every mode constant.
func TestModeNameAllBranches(t *testing.T) {
	t.Parallel()
	cases := map[mode]string{
		modeNormal:  "NORMAL",
		modeInsert:  "INSERT",
		modeCommand: "COMMAND",
		modeSearch:  "SEARCH",
	}
	for m, want := range cases {
		if got := modeName(m); got != want {
			t.Errorf("modeName(%d) = %q, want %q", m, got, want)
		}
	}
}

// TestRedrawSearchStatus covers the search-prompt branches of redraw's status
// line, both forward ("/") and backward ("?").
func TestRedrawSearchStatus(t *testing.T) {
	t.Parallel()

	render := func(e *editor) string {
		var b bytes.Buffer
		redraw(&b, e)
		return b.String()
	}

	// Forward search prompt shows "/pat".
	if got := render(drive("alpha", "/al")); !bytes.Contains([]byte(got), []byte("/al")) {
		t.Errorf("redraw should show forward search prompt /al: %q", got)
	}

	// Backward search prompt shows "?pat".
	if got := render(drive("alpha", "?al")); !bytes.Contains([]byte(got), []byte("?al")) {
		t.Errorf("redraw should show backward search prompt ?al: %q", got)
	}
}
