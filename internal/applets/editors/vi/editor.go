package vi

import "strings"

// mode is the editor's current input mode.
type mode int

const (
	modeNormal mode = iota
	modeInsert
	modeCommand // typing a ":" ex command
)

// editor is the in-memory editing state. It is deliberately decoupled from any
// terminal so the whole command model can be unit-tested by feeding it bytes.
type editor struct {
	lines    []string
	cx, cy   int // cursor column and row (0-based)
	mode     mode
	pending  byte   // first key of a two-key command (d, g, Z)
	cmdline  string // accumulates the ":" command
	filename string
	dirty    bool
	quit     bool   // set when the editor should stop
	save     bool   // set when the buffer should be written on quit
	message  string // status message (e.g. for :q with unsaved changes)
}

// newEditor builds an editor over the given file contents.
func newEditor(filename, content string) *editor {
	lines := []string{""}
	if content != "" {
		lines = strings.Split(strings.TrimSuffix(content, "\n"), "\n")
	}
	return &editor{lines: lines, filename: filename, mode: modeNormal}
}

// content returns the buffer as a string with a trailing newline.
func (e *editor) content() string {
	return strings.Join(e.lines, "\n") + "\n"
}

// feed processes one input byte according to the current mode.
func (e *editor) feed(b byte) {
	switch e.mode {
	case modeInsert:
		e.feedInsert(b)
	case modeCommand:
		e.feedCommand(b)
	default:
		e.feedNormal(b)
	}
}

// feedString feeds every byte of s in order (handy for tests and batch mode).
func (e *editor) feedString(s string) {
	for i := 0; i < len(s) && !e.quit; i++ {
		e.feed(s[i])
	}
}

const esc = 0x1b

// feedNormal handles a key in normal mode.
func (e *editor) feedNormal(b byte) {
	// Resolve a pending two-key command first.
	if e.pending != 0 {
		first := e.pending
		e.pending = 0
		switch {
		case first == 'd' && b == 'd':
			e.deleteLine()
		case first == 'g' && b == 'g':
			e.cy, e.cx = 0, 0
		case first == 'Z' && b == 'Z':
			e.save = true
			e.quit = true
		}
		return
	}

	switch b {
	case 'h':
		e.moveLeft()
	case 'l':
		e.moveRight()
	case 'j':
		e.moveDown()
	case 'k':
		e.moveUp()
	case '0':
		e.cx = 0
	case '$':
		e.cx = lastCol(e.lines[e.cy])
	case 'G':
		e.cy = len(e.lines) - 1
		e.clampX()
	case 'x':
		e.deleteRune()
	case 'i':
		e.mode = modeInsert
	case 'a':
		if len(e.lines[e.cy]) > 0 {
			e.cx++
		}
		e.mode = modeInsert
	case 'A':
		e.cx = len(e.lines[e.cy])
		e.mode = modeInsert
	case 'o':
		e.openBelow()
	case 'O':
		e.openAbove()
	case 'd', 'g', 'Z':
		e.pending = b
	case ':':
		e.mode = modeCommand
		e.cmdline = ""
	}
}

// feedInsert handles a key in insert mode.
func (e *editor) feedInsert(b byte) {
	switch b {
	case esc:
		e.mode = modeNormal
		if e.cx > 0 {
			e.cx--
		}
	case '\r', '\n':
		e.splitLine()
	case 0x7f, 0x08: // DEL / backspace
		e.backspace()
	default:
		e.insertRune(b)
	}
}

// feedCommand accumulates and then runs a ":" command on Enter.
func (e *editor) feedCommand(b byte) {
	switch b {
	case '\r', '\n':
		e.runExCommand(e.cmdline)
		e.mode = modeNormal
		e.cmdline = ""
	case esc:
		e.mode = modeNormal
		e.cmdline = ""
	case 0x7f, 0x08:
		if len(e.cmdline) > 0 {
			e.cmdline = e.cmdline[:len(e.cmdline)-1]
		}
	default:
		e.cmdline += string(b)
	}
}

// runExCommand interprets the small set of ":" commands.
func (e *editor) runExCommand(cmd string) {
	switch cmd {
	case "w":
		e.save = true
	case "wq", "x":
		e.save = true
		e.quit = true
	case "q":
		if e.dirty {
			e.message = "E37: No write since last change (add ! to override)"
			return
		}
		e.quit = true
	case "q!":
		e.quit = true
	}
}

// --- buffer operations ---

func (e *editor) moveLeft() {
	if e.cx > 0 {
		e.cx--
	}
}

func (e *editor) moveRight() {
	if e.cx < lastCol(e.lines[e.cy]) {
		e.cx++
	}
}

func (e *editor) moveDown() {
	if e.cy < len(e.lines)-1 {
		e.cy++
		e.clampX()
	}
}

func (e *editor) moveUp() {
	if e.cy > 0 {
		e.cy--
		e.clampX()
	}
}

func (e *editor) clampX() {
	if m := lastCol(e.lines[e.cy]); e.cx > m {
		e.cx = m
	}
}

func (e *editor) deleteRune() {
	line := e.lines[e.cy]
	if e.cx < len(line) {
		e.lines[e.cy] = line[:e.cx] + line[e.cx+1:]
		e.dirty = true
		e.clampX()
	}
}

func (e *editor) deleteLine() {
	if len(e.lines) == 1 {
		e.lines[0] = ""
	} else {
		e.lines = append(e.lines[:e.cy], e.lines[e.cy+1:]...)
		if e.cy >= len(e.lines) {
			e.cy = len(e.lines) - 1
		}
	}
	e.cx = 0
	e.dirty = true
}

func (e *editor) openBelow() {
	e.lines = insertAt(e.lines, e.cy+1, "")
	e.cy++
	e.cx = 0
	e.mode = modeInsert
	e.dirty = true
}

func (e *editor) openAbove() {
	e.lines = insertAt(e.lines, e.cy, "")
	e.cx = 0
	e.mode = modeInsert
	e.dirty = true
}

func (e *editor) insertRune(b byte) {
	line := e.lines[e.cy]
	if e.cx > len(line) {
		e.cx = len(line)
	}
	e.lines[e.cy] = line[:e.cx] + string(b) + line[e.cx:]
	e.cx++
	e.dirty = true
}

func (e *editor) splitLine() {
	line := e.lines[e.cy]
	if e.cx > len(line) {
		e.cx = len(line)
	}
	head, tail := line[:e.cx], line[e.cx:]
	e.lines[e.cy] = head
	e.lines = insertAt(e.lines, e.cy+1, tail)
	e.cy++
	e.cx = 0
	e.dirty = true
}

func (e *editor) backspace() {
	if e.cx > 0 {
		line := e.lines[e.cy]
		e.lines[e.cy] = line[:e.cx-1] + line[e.cx:]
		e.cx--
		e.dirty = true
		return
	}
	if e.cy > 0 {
		prev := e.lines[e.cy-1]
		e.cx = len(prev)
		e.lines[e.cy-1] = prev + e.lines[e.cy]
		e.lines = append(e.lines[:e.cy], e.lines[e.cy+1:]...)
		e.cy--
		e.dirty = true
	}
}

// lastCol is the rightmost column the cursor may rest on in normal mode.
func lastCol(line string) int {
	if len(line) == 0 {
		return 0
	}
	return len(line) - 1
}

// insertAt inserts v into s at index i.
func insertAt(s []string, i int, v string) []string {
	s = append(s, "")
	copy(s[i+1:], s[i:])
	s[i] = v
	return s
}
