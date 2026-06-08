package vi

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

// drive builds an editor over content and feeds it the keystroke script.
func drive(content, keys string) *editor {
	e := newEditor("test.txt", content)
	e.feedString(keys)
	return e
}

func TestNewEditor(t *testing.T) {
	t.Parallel()
	e := newEditor("f", "")
	if len(e.lines) != 1 || e.lines[0] != "" {
		t.Errorf("empty file should yield one empty line, got %v", e.lines)
	}
	e = newEditor("f", "a\nb\n")
	if len(e.lines) != 2 || e.lines[0] != "a" || e.lines[1] != "b" {
		t.Errorf("lines = %v, want [a b]", e.lines)
	}
}

func TestContent(t *testing.T) {
	t.Parallel()
	e := newEditor("f", "x\ny")
	if e.content() != "x\ny\n" {
		t.Errorf("content = %q", e.content())
	}
}

func TestInsertText(t *testing.T) {
	t.Parallel()
	// i, type "hi", ESC.
	e := drive("", "ihi\x1b")
	if e.lines[0] != "hi" {
		t.Errorf("lines[0] = %q, want hi", e.lines[0])
	}
	if e.mode != modeNormal {
		t.Errorf("ESC should return to normal mode")
	}
	if !e.dirty {
		t.Error("buffer should be dirty after insert")
	}
}

func TestAppend(t *testing.T) {
	t.Parallel()
	// On "ab", 'a' inserts after the cursor (col 0 -> after 'a').
	e := drive("ab", "aX\x1b")
	if e.lines[0] != "aXb" {
		t.Errorf("lines[0] = %q, want aXb", e.lines[0])
	}
}

func TestAppendEndOfLine(t *testing.T) {
	t.Parallel()
	e := drive("ab", "A!\x1b")
	if e.lines[0] != "ab!" {
		t.Errorf("lines[0] = %q, want ab!", e.lines[0])
	}
}

func TestDeleteRune(t *testing.T) {
	t.Parallel()
	e := drive("abc", "x")
	if e.lines[0] != "bc" {
		t.Errorf("lines[0] = %q, want bc", e.lines[0])
	}
}

func TestDeleteLine(t *testing.T) {
	t.Parallel()
	e := drive("one\ntwo\nthree", "jdd") // move to line 2, delete it
	if len(e.lines) != 2 || e.lines[0] != "one" || e.lines[1] != "three" {
		t.Errorf("lines = %v, want [one three]", e.lines)
	}
}

func TestDeleteOnlyLine(t *testing.T) {
	t.Parallel()
	e := drive("solo", "dd")
	if len(e.lines) != 1 || e.lines[0] != "" {
		t.Errorf("lines = %v, want one empty line", e.lines)
	}
}

func TestOpenBelow(t *testing.T) {
	t.Parallel()
	e := drive("top", "oNEW\x1b")
	if len(e.lines) != 2 || e.lines[1] != "NEW" {
		t.Errorf("lines = %v, want [top NEW]", e.lines)
	}
}

func TestOpenAbove(t *testing.T) {
	t.Parallel()
	e := drive("bottom", "ONEW\x1b")
	if len(e.lines) != 2 || e.lines[0] != "NEW" || e.lines[1] != "bottom" {
		t.Errorf("lines = %v, want [NEW bottom]", e.lines)
	}
}

func TestMotions(t *testing.T) {
	t.Parallel()
	e := newEditor("f", "abcde\nfghij")
	e.feedString("ll") // cx -> 2
	if e.cx != 2 {
		t.Errorf("cx = %d, want 2", e.cx)
	}
	e.feedString("$")
	if e.cx != 4 {
		t.Errorf("$ -> cx = %d, want 4", e.cx)
	}
	e.feedString("0")
	if e.cx != 0 {
		t.Errorf("0 -> cx = %d, want 0", e.cx)
	}
	e.feedString("j")
	if e.cy != 1 {
		t.Errorf("j -> cy = %d, want 1", e.cy)
	}
	e.feedString("G")
	if e.cy != 1 {
		t.Errorf("G -> cy = %d, want last line 1", e.cy)
	}
	e.feedString("gg")
	if e.cy != 0 {
		t.Errorf("gg -> cy = %d, want 0", e.cy)
	}
}

func TestSplitAndBackspace(t *testing.T) {
	t.Parallel()
	// Insert at start, press Enter to split "abc" -> "" and "abc".
	e := drive("abc", "i\r\x1b")
	if len(e.lines) != 2 || e.lines[0] != "" || e.lines[1] != "abc" {
		t.Errorf("after split lines = %v", e.lines)
	}

	// Backspace joining: on line 2 col 0, insert-mode backspace joins lines.
	e2 := newEditor("f", "ab\ncd")
	e2.feedString("j")  // line 2
	e2.feedString("i")  // insert at col 0
	e2.feedString("\x7f") // backspace -> join
	if len(e2.lines) != 1 || e2.lines[0] != "abcd" {
		t.Errorf("after join lines = %v, want [abcd]", e2.lines)
	}
}

func TestExWriteQuit(t *testing.T) {
	t.Parallel()
	e := drive("data", ":wq\r")
	if !e.save || !e.quit {
		t.Errorf("save=%v quit=%v, want both true", e.save, e.quit)
	}
}

func TestExWriteThenQuit(t *testing.T) {
	t.Parallel()
	e := drive("x", "iY\x1b:w\r")
	if !e.save {
		t.Error(":w should set save")
	}
	if e.quit {
		t.Error(":w alone should not quit")
	}
}

func TestExQuitDirtyBlocked(t *testing.T) {
	t.Parallel()
	e := drive("x", "iY\x1b:q\r")
	if e.quit {
		t.Error(":q with unsaved changes should not quit")
	}
	if e.message == "" {
		t.Error("expected a warning message")
	}
}

func TestExQuitForce(t *testing.T) {
	t.Parallel()
	e := drive("x", "iY\x1b:q!\r")
	if !e.quit {
		t.Error(":q! should quit")
	}
	if e.save {
		t.Error(":q! should not save")
	}
}

func TestZZ(t *testing.T) {
	t.Parallel()
	e := drive("x", "ZZ")
	if !e.save || !e.quit {
		t.Errorf("ZZ save=%v quit=%v, want both true", e.save, e.quit)
	}
}

func TestCommandModeEscapeAndBackspace(t *testing.T) {
	t.Parallel()
	// ":" then type "wX", backspace removes X, ESC cancels -> no save.
	e := drive("x", ":wX\x7f\x1b")
	if e.save || e.quit {
		t.Errorf("ESC should cancel the command, save=%v quit=%v", e.save, e.quit)
	}
	if e.mode != modeNormal {
		t.Error("ESC should return to normal mode")
	}
}

func TestRedraw(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		keys string
		want string
	}{
		{"normal status", "", "NORMAL"},
		{"insert status", "i", "INSERT"},
		{"command line", ":w", ":w"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			e := drive("alpha\nbeta", tt.keys)
			var b bytes.Buffer
			redraw(&b, e)
			out := b.String()
			if !strings.Contains(out, "alpha") || !strings.Contains(out, "beta") {
				t.Errorf("redraw missing buffer text: %q", out)
			}
			if !strings.Contains(out, tt.want) {
				t.Errorf("redraw = %q, want it to contain %q", out, tt.want)
			}
			if !strings.Contains(out, "\x1b[2J") {
				t.Errorf("redraw should clear the screen: %q", out)
			}
		})
	}
}

func TestRedrawDirtyAndMessage(t *testing.T) {
	t.Parallel()
	// Dirty marker after an edit.
	e := drive("x", "iY\x1b")
	var b bytes.Buffer
	redraw(&b, e)
	if !strings.Contains(b.String(), "[+]") {
		t.Errorf("redraw should show dirty marker: %q", b.String())
	}

	// A status message (e.g. from :q on a dirty buffer) is shown verbatim.
	e2 := drive("x", "iY\x1b:q\r")
	var b2 bytes.Buffer
	redraw(&b2, e2)
	if !strings.Contains(b2.String(), "No write since last change") {
		t.Errorf("redraw should show the message: %q", b2.String())
	}
}

func TestIsTerminalNonFile(t *testing.T) {
	t.Parallel()
	if isTerminal(strings.NewReader("x")) {
		t.Error("a strings.Reader is not a terminal")
	}
}

func TestIsTerminalPipeIsNotTTY(t *testing.T) {
	t.Parallel()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = r.Close(); _ = w.Close() }()
	if isTerminal(r) {
		t.Error("an os.Pipe read end is not a terminal")
	}
}

// TestRunInteractiveFallsBackOnPipe verifies that runInteractive degrades to
// batch processing when its input is an *os.File that is not a real terminal
// (the TCGETS ioctl fails on a pipe).
func TestRunInteractiveFallsBackOnPipe(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	if _, werr := w.WriteString("x:wq\r"); werr != nil {
		t.Fatal(werr)
	}
	_ = w.Close()
	defer func() { _ = r.Close() }()

	e := newEditor("f", "abc")
	out := &bytes.Buffer{}
	runInteractive(command.IO{In: r, Out: out, Err: &bytes.Buffer{}}, e)

	if e.lines[0] != "bc" {
		t.Errorf("batch fallback should have applied 'x': lines[0] = %q", e.lines[0])
	}
	if !e.save || !e.quit {
		t.Errorf("batch fallback should have run :wq (save=%v quit=%v)", e.save, e.quit)
	}
}
