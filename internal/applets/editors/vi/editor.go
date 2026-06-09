package vi

import "strings"

// mode is the editor's current input mode.
type mode int

const (
	modeNormal mode = iota
	modeInsert
	modeCommand // typing a ":" ex command
	modeSearch  // typing a "/" or "?" search pattern
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

	count            int        // pending numeric prefix (3j, 2x, ...)
	register         string     // yanked text (yy/dd) for p/P
	registerLinewise bool       // whether register holds whole lines
	undoStack        []snapshot // states to restore with u
	searchPat        string     // current "/" or "?" pattern being typed
	searchForward    bool       // direction of the in-progress search
	lastSearch       string     // last completed search pattern (for n/N)
	lastForward      bool       // direction of the last search
}

// snapshot is a saved editor state for undo.
type snapshot struct {
	lines  []string
	cx, cy int
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
	case modeSearch:
		e.feedSearch(b)
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
	case modeSearch:
		e.mode = modeNormal
		e.searchPat = ""
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
	// A leading digit (or a non-leading 0) builds the count prefix.
	if (b >= '1' && b <= '9') || (b == '0' && e.count > 0) {
		e.count = e.count*10 + int(b-'0')
		return
	}

	if e.pending != 0 {
		e.handlePending(b)
		return
	}

	// Operator/prefix keys keep the count for their second key.
	switch b {
	case 'd', 'g', 'Z', 'y':
		e.pending = b
		return
	}

	hadCount := e.count > 0
	cnt := e.takeCount()
	switch b {
	case 'h':
		e.repeat(cnt, e.moveLeft)
	case 'l':
		e.repeat(cnt, e.moveRight)
	case 'j':
		e.repeat(cnt, e.moveDown)
	case 'k':
		e.repeat(cnt, e.moveUp)
	case 'w':
		e.repeat(cnt, e.moveWordForward)
	case 'b':
		e.repeat(cnt, e.moveWordBackward)
	case 'e':
		e.repeat(cnt, e.moveWordEnd)
	case '0':
		e.cx = 0
	case '$':
		e.cx = lastCol(e.lines[e.cy])
	case 'G':
		// With a count, go to that 1-based line; otherwise the last line.
		if hadCount {
			e.cy = cnt - 1
		} else {
			e.cy = len(e.lines) - 1
		}
		e.clampY()
		e.clampX()
	case 'x':
		e.snapshot()
		e.repeat(cnt, e.deleteRune)
	case 'u':
		e.undo()
	case 'p':
		e.snapshot()
		e.repeat(cnt, e.pasteBelow)
	case 'P':
		e.snapshot()
		e.repeat(cnt, e.pasteAbove)
	case 'i':
		e.snapshot()
		e.mode = modeInsert
	case 'a':
		e.snapshot()
		if len(e.lines[e.cy]) > 0 {
			e.cx++
		}
		e.mode = modeInsert
	case 'A':
		e.snapshot()
		e.cx = len(e.lines[e.cy])
		e.mode = modeInsert
	case 'o':
		e.snapshot()
		e.openBelow()
	case 'O':
		e.snapshot()
		e.openAbove()
	case '/':
		e.mode = modeSearch
		e.searchForward = true
		e.searchPat = ""
	case '?':
		e.mode = modeSearch
		e.searchForward = false
		e.searchPat = ""
	case 'n':
		e.searchAgain(e.lastForward)
	case 'N':
		e.searchAgain(!e.lastForward)
	case ':':
		e.mode = modeCommand
		e.cmdline = ""
	}
}

// handlePending resolves the second key of an operator (d, g, Z, y).
func (e *editor) handlePending(b byte) {
	first := e.pending
	e.pending = 0
	cnt := e.takeCount()
	switch {
	case first == 'd' && b == 'd':
		e.snapshot()
		e.yankLines(cnt) // dd also fills the register, like vi
		e.repeat(cnt, e.deleteLine)
	case first == 'y' && b == 'y':
		e.yankLines(cnt)
	case first == 'g' && b == 'g':
		e.cy, e.cx = 0, 0
	case first == 'Z' && b == 'Z':
		e.save = true
		e.quit = true
	}
}

// takeCount returns the pending count (at least 1) and clears it.
func (e *editor) takeCount() int {
	c := e.count
	e.count = 0
	if c < 1 {
		return 1
	}
	return c
}

// repeat runs f n times.
func (e *editor) repeat(n int, f func()) {
	for i := 0; i < n; i++ {
		f()
	}
}

// clampY keeps the cursor row within the buffer.
func (e *editor) clampY() {
	if e.cy < 0 {
		e.cy = 0
	}
	if e.cy > len(e.lines)-1 {
		e.cy = len(e.lines) - 1
	}
}

// snapshot records the current buffer and cursor so u can restore them. It is
// taken before each change (and once when entering insert mode, so a whole
// insertion undoes as one step).
func (e *editor) snapshot() {
	cp := make([]string, len(e.lines))
	copy(cp, e.lines)
	e.undoStack = append(e.undoStack, snapshot{lines: cp, cx: e.cx, cy: e.cy})
}

// undo restores the most recent snapshot.
func (e *editor) undo() {
	if len(e.undoStack) == 0 {
		return
	}
	s := e.undoStack[len(e.undoStack)-1]
	e.undoStack = e.undoStack[:len(e.undoStack)-1]
	e.lines = s.lines
	e.cy, e.cx = s.cy, s.cx
	e.clampY()
	e.clampX()
	e.dirty = true
}

// yankLines copies n lines from the cursor into the register (linewise).
func (e *editor) yankLines(n int) {
	end := e.cy + n
	if end > len(e.lines) {
		end = len(e.lines)
	}
	var b strings.Builder
	for i := e.cy; i < end; i++ {
		b.WriteString(e.lines[i])
		b.WriteByte('\n')
	}
	e.register = b.String()
	e.registerLinewise = true
}

// pasteBelow inserts the register's lines below the current line (p).
func (e *editor) pasteBelow() {
	if e.register == "" || !e.registerLinewise {
		return
	}
	lines := strings.Split(strings.TrimSuffix(e.register, "\n"), "\n")
	e.lines = insertLines(e.lines, e.cy+1, lines)
	e.cy++
	e.cx = 0
	e.dirty = true
}

// pasteAbove inserts the register's lines above the current line (P).
func (e *editor) pasteAbove() {
	if e.register == "" || !e.registerLinewise {
		return
	}
	lines := strings.Split(strings.TrimSuffix(e.register, "\n"), "\n")
	e.lines = insertLines(e.lines, e.cy, lines)
	e.cx = 0
	e.dirty = true
}

// insertLines inserts vs into s at index i.
func insertLines(s []string, i int, vs []string) []string {
	out := make([]string, 0, len(s)+len(vs))
	out = append(out, s[:i]...)
	out = append(out, vs...)
	out = append(out, s[i:]...)
	return out
}

func isWordSpace(c byte) bool { return c == ' ' || c == '\t' }

// moveWordForward moves to the start of the next whitespace-delimited word,
// wrapping onto following lines.
func (e *editor) moveWordForward() {
	line := e.lines[e.cy]
	i := e.cx
	for i < len(line) && !isWordSpace(line[i]) {
		i++
	}
	for i < len(line) && isWordSpace(line[i]) {
		i++
	}
	if i < len(line) {
		e.cx = i
		return
	}
	if e.cy < len(e.lines)-1 {
		e.cy++
		l := e.lines[e.cy]
		j := 0
		for j < len(l) && isWordSpace(l[j]) {
			j++
		}
		e.cx = j
		e.clampX()
		return
	}
	e.cx = lastCol(line)
}

// moveWordBackward moves to the start of the previous whitespace-delimited word.
func (e *editor) moveWordBackward() {
	line := e.lines[e.cy]
	i := e.cx
	if i == 0 {
		if e.cy == 0 {
			return
		}
		e.cy--
		line = e.lines[e.cy]
		i = len(line)
	}
	i--
	for i > 0 && isWordSpace(line[i]) {
		i--
	}
	for i > 0 && !isWordSpace(line[i-1]) {
		i--
	}
	if i < 0 {
		i = 0
	}
	e.cx = i
}

// moveWordEnd moves to the end of the next whitespace-delimited word.
func (e *editor) moveWordEnd() {
	line := e.lines[e.cy]
	i := e.cx + 1
	for i < len(line) && isWordSpace(line[i]) {
		i++
	}
	for i < len(line)-1 && !isWordSpace(line[i+1]) {
		i++
	}
	if i < len(line) {
		e.cx = i
		e.clampX()
		return
	}
	if e.cy < len(e.lines)-1 {
		e.cy++
		l := e.lines[e.cy]
		j := 0
		for j < len(l) && isWordSpace(l[j]) {
			j++
		}
		for j < len(l)-1 && !isWordSpace(l[j+1]) {
			j++
		}
		e.cx = j
		e.clampX()
	}
}

// searchExecute moves the cursor to the next literal match of pat in the given
// direction, wrapping around the buffer.
func (e *editor) searchExecute(pat string, forward bool) {
	if pat == "" {
		return
	}
	n := len(e.lines)
	if forward {
		for off := 0; off < n; off++ {
			row := (e.cy + off) % n
			line := e.lines[row]
			from := 0
			if off == 0 {
				from = e.cx + 1
			}
			if from <= len(line) {
				if idx := strings.Index(line[from:], pat); idx >= 0 {
					e.cy, e.cx = row, from+idx
					e.clampX()
					return
				}
			}
		}
		return
	}
	for off := 0; off < n; off++ {
		row := ((e.cy-off)%n + n) % n
		line := e.lines[row]
		limit := len(line)
		if off == 0 {
			limit = e.cx
		}
		if limit >= 0 && limit <= len(line) {
			if idx := strings.LastIndex(line[:limit], pat); idx >= 0 {
				e.cy, e.cx = row, idx
				e.clampX()
				return
			}
		}
	}
}

// searchAgain repeats the last search in the given direction (n / N).
func (e *editor) searchAgain(forward bool) {
	if e.lastSearch != "" {
		e.searchExecute(e.lastSearch, forward)
	}
}

// feedSearch accumulates a "/" or "?" pattern and runs it on Enter.
func (e *editor) feedSearch(b byte) {
	switch b {
	case '\r', '\n':
		e.mode = modeNormal
		e.lastSearch = e.searchPat
		e.lastForward = e.searchForward
		e.searchExecute(e.searchPat, e.searchForward)
		e.searchPat = ""
	case 0x7f, 0x08:
		if len(e.searchPat) > 0 {
			e.searchPat = e.searchPat[:len(e.searchPat)-1]
		}
	default:
		e.searchPat += string(b)
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
