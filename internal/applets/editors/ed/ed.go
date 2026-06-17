// Package ed implements the ed applet: a minimal line editor suited to scripted
// editing. It supports line addressing and the core verbs (a, i, c, d, p, n, w,
// q, =, s) that cover non-interactive edits.
package ed

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the ed applet.
type Command struct{}

// New returns an ed command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "ed" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "A line-oriented text editor" }

// editor holds the buffer and editing state.
type editor struct {
	lines    []string // line N is lines[N-1]
	dot      int      // current line (1-based; 0 when empty)
	file     string
	in       *bufio.Scanner
	out      io.Writer
	errw     io.Writer
	modified bool
}

// Run executes ed.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[FILE]", stdio.Err).WithHelp(command.Help{
		Description: "Edit text line by line, reading commands from standard input. With FILE, load it " +
			"first and print its size. Commands: a/i/c (add text, ended by a line with just '.'), " +
			"d (delete), p/n (print), w (write), = (line number), s/re/repl/ (substitute), q (quit).",
		Examples: []command.Example{
			{Command: "printf '1,$p\\nq\\n' | ed file.txt", Explain: "Print the whole file."},
			{Command: "printf '2a\\nnew line\\n.\\nw\\nq\\n' | ed file.txt", Explain: "Insert a line after line 2 and save."},
		},
		ExitStatus: "0  the file was edited and written without error.\n1  an error occurred.",
	})
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	sc := bufio.NewScanner(stdio.In)
	sc.Buffer(make([]byte, 0, 64*1024), command.MaxLineSize)

	e := &editor{dot: 0, in: sc, out: stdio.Out, errw: stdio.Err}
	if rest := fs.Args(); len(rest) > 0 {
		e.file = rest[0]
		if err := e.load(); err != nil {
			_, _ = fmt.Fprintln(stdio.Err, "?")
		}
	}

	for sc.Scan() {
		if quit := e.command(sc.Text()); quit {
			break
		}
	}
	return nil
}

// load reads the file into the buffer and prints its byte count.
func (e *editor) load() error {
	data, err := os.ReadFile(e.file) //nolint:gosec // user-named file
	if err != nil {
		return err
	}
	text := string(data)
	if text == "" {
		e.lines = nil
	} else {
		e.lines = strings.Split(strings.TrimSuffix(text, "\n"), "\n")
	}
	e.dot = len(e.lines)
	_, _ = fmt.Fprintln(e.out, byteCount(e.lines))
	return nil
}

// byteCount returns the size of the buffer as ed counts it (each line plus a
// trailing newline).
func byteCount(lines []string) int {
	n := 0
	for _, l := range lines {
		n += len(l) + 1
	}
	return n
}

// command executes one command line and reports whether to quit.
func (e *editor) command(line string) (quit bool) {
	a1, a2, naddr, rest := e.parseAddresses(line)
	if rest == "" {
		// A bare address moves to and prints that line; an empty line advances.
		target := a2
		if naddr == 0 {
			target = e.dot + 1
		}
		if !e.valid(target) {
			e.fail()
			return false
		}
		e.dot = target
		_, _ = fmt.Fprintln(e.out, e.lines[e.dot-1])
		return false
	}

	cmd := rest[0]
	arg := strings.TrimSpace(rest[1:])
	switch cmd {
	case 'q':
		return true
	case 'Q':
		return true
	case 'a':
		e.insert(e.defaultLine(naddr, a2), false)
	case 'i':
		e.insert(e.defaultLine(naddr, a2), true)
	case 'c':
		e.change(naddr, a1, a2)
	case 'd':
		e.delete(naddr, a1, a2)
	case 'p':
		e.print(naddr, a1, a2, false)
	case 'n':
		e.print(naddr, a1, a2, true)
	case '=':
		n := len(e.lines)
		if naddr > 0 {
			n = a2
		}
		_, _ = fmt.Fprintln(e.out, n)
	case 'w':
		e.write(arg)
	case 's':
		e.substitute(naddr, a1, a2, rest[1:])
	default:
		e.fail()
	}
	return false
}

// defaultLine returns the address to use for a/i, defaulting to the current line.
func (e *editor) defaultLine(naddr, a2 int) int {
	if naddr == 0 {
		return e.dot
	}
	return a2
}

func (e *editor) fail() { _, _ = fmt.Fprintln(e.errw, "?") }

func (e *editor) valid(n int) bool { return n >= 1 && n <= len(e.lines) }

// insert reads input lines (until a "." line) and inserts them. When before is
// true they go before the address (i); otherwise after it (a).
func (e *editor) insert(at int, before bool) {
	pos := at // 'a' appends after 'at'
	if before {
		pos = at - 1 // 'i' inserts before 'at'
	}
	if pos < 0 {
		pos = 0
	}
	if pos > len(e.lines) {
		pos = len(e.lines)
	}
	added := e.readInput()
	e.lines = append(e.lines[:pos], append(added, e.lines[pos:]...)...)
	if len(added) > 0 {
		e.dot = pos + len(added)
		e.modified = true
	}
}

// readInput collects buffer lines from the command stream until a line
// containing only ".".
func (e *editor) readInput() []string {
	var added []string
	for e.in.Scan() {
		t := e.in.Text()
		if t == "." {
			break
		}
		added = append(added, t)
	}
	return added
}

// change deletes the range then inserts new text in its place.
func (e *editor) change(naddr, a1, a2 int) {
	lo, hi := e.rangeOrDot(naddr, a1, a2)
	if !e.valid(lo) || !e.valid(hi) {
		e.fail()
		return
	}
	e.lines = append(e.lines[:lo-1], e.lines[hi:]...)
	added := e.readInput()
	e.lines = append(e.lines[:lo-1], append(added, e.lines[lo-1:]...)...)
	e.dot = lo - 1 + len(added)
	e.modified = true
}

// delete removes the addressed range.
func (e *editor) delete(naddr, a1, a2 int) {
	lo, hi := e.rangeOrDot(naddr, a1, a2)
	if !e.valid(lo) || !e.valid(hi) {
		e.fail()
		return
	}
	e.lines = append(e.lines[:lo-1], e.lines[hi:]...)
	e.dot = lo
	if e.dot > len(e.lines) {
		e.dot = len(e.lines)
	}
	e.modified = true
}

// print writes the addressed range, optionally with line numbers.
func (e *editor) print(naddr, a1, a2 int, numbered bool) {
	lo, hi := e.rangeOrDot(naddr, a1, a2)
	if !e.valid(lo) || !e.valid(hi) {
		e.fail()
		return
	}
	for i := lo; i <= hi; i++ {
		if numbered {
			_, _ = fmt.Fprintf(e.out, "%d\t%s\n", i, e.lines[i-1])
		} else {
			_, _ = fmt.Fprintln(e.out, e.lines[i-1])
		}
	}
	e.dot = hi
}

// write saves the buffer to arg (or the current file) and prints its byte count.
func (e *editor) write(arg string) {
	name := e.file
	if arg != "" {
		name = arg
	}
	if name == "" {
		e.fail()
		return
	}
	var b strings.Builder
	for _, l := range e.lines {
		b.WriteString(l)
		b.WriteByte('\n')
	}
	if err := os.WriteFile(name, []byte(b.String()), 0o644); err != nil { //nolint:gosec // user-named file
		e.fail()
		return
	}
	e.modified = false
	_, _ = fmt.Fprintln(e.out, byteCount(e.lines))
}

// substitute applies s/re/repl/[g] to the addressed range (default current line).
func (e *editor) substitute(naddr, a1, a2 int, spec string) {
	lo, hi := e.rangeOrDot(naddr, a1, a2)
	if !e.valid(lo) || !e.valid(hi) || len(spec) < 1 {
		e.fail()
		return
	}
	delim := spec[0]
	parts := strings.Split(spec[1:], string(delim))
	if len(parts) < 3 {
		e.fail()
		return
	}
	re, err := regexp.Compile(parts[0])
	if err != nil {
		e.fail()
		return
	}
	repl := parts[1]
	global := strings.Contains(parts[2], "g")
	for i := lo; i <= hi; i++ {
		if global {
			e.lines[i-1] = re.ReplaceAllString(e.lines[i-1], repl)
		} else {
			e.lines[i-1] = replaceFirst(re, e.lines[i-1], repl)
		}
	}
	e.dot = hi
	e.modified = true
}

// replaceFirst replaces only the first match of re in s.
func replaceFirst(re *regexp.Regexp, s, repl string) string {
	loc := re.FindStringIndex(s)
	if loc == nil {
		return s
	}
	return s[:loc[0]] + re.ReplaceAllString(s[loc[0]:loc[1]], repl) + s[loc[1]:]
}

// rangeOrDot resolves the effective range, defaulting to the current line.
func (e *editor) rangeOrDot(naddr, a1, a2 int) (int, int) {
	switch naddr {
	case 0:
		return e.dot, e.dot
	case 1:
		return a2, a2
	default:
		return a1, a2
	}
}

// parseAddresses parses the leading address(es) of a command line and returns
// the first and second address, how many were given, and the remaining command
// text.
func (e *editor) parseAddresses(s string) (a1, a2, naddr int, rest string) {
	i := 0
	first, ni, ok1 := e.oneAddr(s, i)
	if !ok1 {
		return 0, 0, 0, s
	}
	i = ni
	a1, a2 = first, first
	naddr = 1
	if i < len(s) && (s[i] == ',' || s[i] == ';') {
		sep := s[i]
		if sep == ';' {
			e.dot = first
		}
		i++
		second, nj, ok2 := e.oneAddr(s, i)
		if ok2 {
			a2 = second
			i = nj
		} else {
			a2 = len(e.lines)
		}
		naddr = 2
	}
	return a1, a2, naddr, s[i:]
}

// oneAddr parses a single address at i, returning its value, the next index, and
// whether an address was present.
func (e *editor) oneAddr(s string, i int) (val, next int, ok bool) {
	if i >= len(s) {
		return 0, i, false
	}
	switch s[i] {
	case '.':
		return e.dot, i + 1, true
	case '$':
		return len(e.lines), i + 1, true
	case ',':
		// A leading comma means 1,$ for the whole-file range; handled by caller.
		return 1, i, true
	case '+', '-':
		j := i + 1
		for j < len(s) && s[j] >= '0' && s[j] <= '9' {
			j++
		}
		n := 1
		if j > i+1 {
			n, _ = strconv.Atoi(s[i+1 : j])
		}
		if s[i] == '-' {
			return e.dot - n, j, true
		}
		return e.dot + n, j, true
	default:
		if s[i] >= '0' && s[i] <= '9' {
			j := i
			for j < len(s) && s[j] >= '0' && s[j] <= '9' {
				j++
			}
			n, _ := strconv.Atoi(s[i:j])
			return n, j, true
		}
	}
	return 0, i, false
}
