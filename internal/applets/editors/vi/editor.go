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

	// Terminal escape-sequence decoding state. Arrow keys and similar keys
	// arrive as multi-byte sequences (ESC [ A, ...); decoding them here keeps
	// the trailing byte from being mistaken for an editing command.
	escSt  escState
	csiBuf []byte // parameter/intermediate bytes of a CSI sequence
}

// escState tracks where we are in decoding a terminal escape sequence.
type escState int

const (
	escNone   escState = iota // not in a sequence
	escGotEsc                 // saw ESC; the next byte decides
	escCSI                    // saw ESC [ ; collecting until a final byte
	escSS3                    // saw ESC O ; the next byte is the key
)

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

const esc = 0x1b

// feed processes one input byte, first decoding terminal escape sequences
// (arrow keys, Home/End, Delete, ...) into editor actions so their trailing
// byte is never mistaken for an editing command, then dispatching ordinary keys
// to the current mode.
func (e *editor) feed(b byte) {
	switch e.escSt {
	case escGotEsc:
		e.escSt = escNone
		switch b {
		case '[':
			e.escSt = escCSI
			e.csiBuf = e.csiBuf[:0]
		case 'O':
			e.escSt = escSS3
		default:
			// A lone ESC (not the start of a sequence): apply it, then process
			// this byte normally.
			e.handleEsc()
			e.dispatch(b)
		}
		return
	case escCSI:
		// A CSI sequence ends at a final byte in 0x40..0x7e; earlier bytes are
		// parameters/intermediates.
		if b >= 0x40 && b <= 0x7e {
			e.escSt = escNone
			e.decodeCSI(append(e.csiBuf, b))
		} else {
			e.csiBuf = append(e.csiBuf, b)
		}
		return
	case escSS3:
		e.escSt = escNone
		e.decodeArrow(b)
		return
	}

	if b == esc {
		e.escSt = escGotEsc
		return
	}
	e.dispatch(b)
}

// dispatch routes an ordinary (non-escape) byte to the current mode.
func (e *editor) dispatch(b byte) {
	switch e.mode {
	case modeInsert:
		e.feedInsert(b)
	case modeCommand:
		e.feedCommand(b)
	default:
		e.feedNormal(b)
	}
}

// handleEsc applies a standalone Escape key: it leaves insert/command mode for
// normal mode (a no-op when already in normal mode).
func (e *editor) handleEsc() {
	switch e.mode {
	case modeInsert:
		e.mode = modeNormal
		if e.cx > 0 {
			e.cx--
		}
	case modeCommand:
		e.mode = modeNormal
		e.cmdline = ""
	}
}

// decodeCSI turns a CSI sequence (the bytes after ESC [) into a cursor motion
// or edit. Unknown sequences are ignored so they can never mutate the buffer in
// surprising ways.
func (e *editor) decodeCSI(seq []byte) {
	if len(seq) == 0 {
		return
	}
	final := seq[len(seq)-1]
	param := string(seq[:len(seq)-1])
	switch final {
	case 'A', 'B', 'C', 'D', 'H', 'F':
		e.decodeArrow(final)
	case '~':
		switch param {
		case "1", "7": // Home
			e.cx = 0
		case "4", "8": // End
			e.cx = lastCol(e.lines[e.cy])
		case "3": // Delete
			e.deleteRune()
		}
	}
}

// decodeArrow maps a final byte shared by CSI and SS3 cursor keys to a motion.
func (e *editor) decodeArrow(final byte) {
	switch final {
	case 'A':
		e.moveUp()
	case 'B':
		e.moveDown()
	case 'C':
		e.moveRight()
	case 'D':
		e.moveLeft()
	case 'H': // Home
		e.cx = 0
	case 'F': // End
		e.cx = lastCol(e.lines[e.cy])
	}
}

// flush resolves any pending escape state at end of input: a lone trailing ESC
// is applied, and an incomplete sequence is dropped.
func (e *editor) flush() {
	if e.escSt == escGotEsc {
		e.handleEsc()
	}
	e.escSt = escNone
}

// feedString feeds every byte of s in order (handy for tests and batch mode),
// then flushes any pending escape state.
func (e *editor) feedString(s string) {
	for i := 0; i < len(s) && !e.quit; i++ {
		e.feed(s[i])
	}
	e.flush()
}

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

// feedInsert handles a key in insert mode. Escape is decoded earlier (see feed
// and handleEsc), so it never reaches here.
func (e *editor) feedInsert(b byte) {
	switch b {
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
