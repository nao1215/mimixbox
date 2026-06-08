package sed

import (
	"fmt"
	"io"
	"regexp"
	"strings"
)

// addrKind classifies the three address forms sed understands.
type addrKind int

const (
	addrNone addrKind = iota
	addrLine
	addrLast
	addrRegex
)

// address is one address on a command (a line number, $ or /regexp/).
type address struct {
	kind addrKind
	line int
	re   *regexp.Regexp
}

// cmd is one parsed sed command with its optional one- or two-address selector.
type cmd struct {
	a1, a2 *address
	active bool // range state for two-address commands

	name byte // 's', 'p', 'd', 'q'

	// substitute parameters (name == 's')
	re         *regexp.Regexp
	repl       string
	global     bool
	printSub   bool
	occurrence int
}

// parse turns a script string into a list of commands. When extended is false,
// patterns are treated as POSIX basic regular expressions (BRE) and translated
// to the extended syntax Go's regexp engine expects.
func parse(script string, extended bool) ([]cmd, error) {
	p := &parser{s: script, extended: extended}
	var cmds []cmd
	for {
		p.skipSeparators()
		if p.eof() {
			return cmds, nil
		}
		c, err := p.command()
		if err != nil {
			return nil, err
		}
		cmds = append(cmds, c)
	}
}

// parser walks the script text one rune at a time.
type parser struct {
	s        string
	i        int
	extended bool
}

// compileRE compiles expr as a regexp, translating BRE to ERE first unless the
// parser is in extended mode.
func (p *parser) compileRE(expr string) (*regexp.Regexp, error) {
	if !p.extended {
		expr = bre2ere(expr)
	}
	return regexp.Compile(expr)
}

// bre2ere converts a POSIX basic regular expression to the equivalent extended
// syntax Go's regexp engine uses: backslash-escaped \( \) \{ \} \+ \? \| become
// the operators, while bare ( ) { } + ? | are literal in BRE and so are
// escaped.
func bre2ere(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		ch := s[i]
		if ch == '\\' && i+1 < len(s) {
			next := s[i+1]
			switch next {
			case '(', ')', '{', '}', '+', '?', '|':
				b.WriteByte(next) // promote to operator
			default:
				b.WriteByte('\\')
				b.WriteByte(next)
			}
			i++
			continue
		}
		switch ch {
		case '(', ')', '{', '}', '+', '?', '|':
			b.WriteByte('\\') // demote bare operator to literal
			b.WriteByte(ch)
		default:
			b.WriteByte(ch)
		}
	}
	return b.String()
}

func (p *parser) eof() bool { return p.i >= len(p.s) }

// skipSeparators consumes whitespace, newlines and ';' between commands.
func (p *parser) skipSeparators() {
	for !p.eof() {
		switch p.s[p.i] {
		case ' ', '\t', '\n', ';':
			p.i++
		default:
			return
		}
	}
}

// command parses one addressed command.
func (p *parser) command() (cmd, error) {
	var c cmd
	a1 := p.address()
	c.a1 = a1
	if a1 != nil && !p.eof() && p.s[p.i] == ',' {
		p.i++
		c.a2 = p.address()
		if c.a2 == nil {
			return c, fmt.Errorf("expected address after ','")
		}
	}
	p.skipSpaces()
	if p.eof() {
		return c, fmt.Errorf("missing command")
	}

	c.name = p.s[p.i]
	p.i++
	switch c.name {
	case 's':
		return p.substitute(c)
	case 'p', 'd', 'q':
		return c, nil
	default:
		return c, fmt.Errorf("unknown command: %q", string(c.name))
	}
}

// address parses an optional address at the cursor, or returns nil.
func (p *parser) address() *address {
	p.skipSpaces()
	if p.eof() {
		return nil
	}
	switch ch := p.s[p.i]; {
	case ch == '$':
		p.i++
		return &address{kind: addrLast}
	case ch >= '0' && ch <= '9':
		start := p.i
		for !p.eof() && p.s[p.i] >= '0' && p.s[p.i] <= '9' {
			p.i++
		}
		n, _ := strconvAtoi(p.s[start:p.i])
		return &address{kind: addrLine, line: n}
	case ch == '/':
		p.i++
		expr := p.until('/')
		re, err := p.compileRE(expr)
		if err != nil {
			return &address{kind: addrRegex, re: regexp.MustCompile(regexp.QuoteMeta(expr))}
		}
		return &address{kind: addrRegex, re: re}
	default:
		return nil
	}
}

// substitute parses the body of an s command: s<delim>re<delim>repl<delim>flags.
func (p *parser) substitute(c cmd) (cmd, error) {
	if p.eof() {
		return c, fmt.Errorf("unterminated 's' command")
	}
	delim := p.s[p.i]
	p.i++
	pattern := p.until(delim)
	repl := p.until(delim)
	flags := p.flags()

	caseInsensitive := strings.ContainsRune(flags, 'i') || strings.ContainsRune(flags, 'I')
	expr := pattern
	if !p.extended {
		expr = bre2ere(expr)
	}
	if caseInsensitive {
		expr = "(?i)" + expr
	}
	re, err := regexp.Compile(expr)
	if err != nil {
		return c, fmt.Errorf("invalid regex in 's' command: %v", err)
	}
	c.re = re
	c.repl = translateRepl(repl)
	c.global = strings.ContainsRune(flags, 'g')
	c.printSub = strings.ContainsRune(flags, 'p')
	for _, r := range flags {
		if r >= '0' && r <= '9' {
			c.occurrence = int(r - '0')
		}
	}
	return c, nil
}

// until consumes runes up to (and consuming) the next unescaped delim, returning
// the text in between with the escape of delim resolved.
func (p *parser) until(delim byte) string {
	var b strings.Builder
	for !p.eof() {
		ch := p.s[p.i]
		if ch == '\\' && p.i+1 < len(p.s) {
			next := p.s[p.i+1]
			if next == delim {
				b.WriteByte(delim)
				p.i += 2
				continue
			}
			b.WriteByte(ch)
			b.WriteByte(next)
			p.i += 2
			continue
		}
		if ch == delim {
			p.i++
			return b.String()
		}
		b.WriteByte(ch)
		p.i++
	}
	return b.String()
}

// flags reads the trailing flag characters of an s command.
func (p *parser) flags() string {
	start := p.i
	for !p.eof() {
		ch := p.s[p.i]
		if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') {
			p.i++
			continue
		}
		break
	}
	return p.s[start:p.i]
}

func (p *parser) skipSpaces() {
	for !p.eof() && (p.s[p.i] == ' ' || p.s[p.i] == '\t') {
		p.i++
	}
}

// translateRepl converts a sed replacement (& for whole match, \1..\9 for
// groups) into the Go regexp template syntax ($0, $1, ...), escaping literal $.
func translateRepl(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '&':
			b.WriteString("${0}")
		case '$':
			b.WriteString("$$")
		case '\\':
			if i+1 < len(s) {
				next := s[i+1]
				switch {
				case next >= '0' && next <= '9':
					b.WriteString("${" + string(next) + "}")
				case next == '&':
					b.WriteByte('&')
				case next == '\\':
					b.WriteByte('\\')
				case next == 'n':
					b.WriteByte('\n')
				case next == 't':
					b.WriteByte('\t')
				default:
					b.WriteByte(next)
				}
				i++
				continue
			}
			b.WriteByte('\\')
		default:
			b.WriteByte(s[i])
		}
	}
	return b.String()
}

// editor applies a parsed program to a slice of lines.
type editor struct {
	program []cmd
	quiet   bool
	out     io.Writer
}

// run processes every line through the program, honoring auto-print.
func (e *editor) run(lines []string) {
	last := len(lines)
	for idx := range lines {
		lineNo := idx + 1
		pattern := lines[idx]
		deleted := false
		quit := false

		for ci := range e.program {
			c := &e.program[ci]
			if !e.selected(c, lineNo, last, pattern) {
				continue
			}
			switch c.name {
			case 's':
				pattern = e.applySub(c, pattern)
			case 'p':
				_, _ = fmt.Fprintln(e.out, pattern)
			case 'd':
				deleted = true
			case 'q':
				quit = true
			}
			if deleted || quit {
				break
			}
		}

		if !deleted && !e.quiet {
			_, _ = fmt.Fprintln(e.out, pattern)
		}
		if quit {
			return
		}
	}
}

// applySub runs the substitute command on a line, honoring g, the Nth-occurrence
// selector and the p flag.
func (e *editor) applySub(c *cmd, line string) string {
	var result string
	switch {
	case c.occurrence > 0:
		result = replaceNth(c.re, line, c.repl, c.occurrence, c.global)
	case c.global:
		result = c.re.ReplaceAllString(line, c.repl)
	default:
		result = replaceNth(c.re, line, c.repl, 1, false)
	}
	if c.printSub && result != line {
		_, _ = fmt.Fprintln(e.out, result)
	}
	return result
}

// replaceNth replaces the nth match (and, when global, every match after it).
func replaceNth(re *regexp.Regexp, line, repl string, n int, global bool) string {
	locs := re.FindAllStringSubmatchIndex(line, -1)
	if len(locs) < n {
		return line
	}
	var b strings.Builder
	prev := 0
	for i, loc := range locs {
		if i+1 < n {
			continue
		}
		if i+1 > n && !global {
			break
		}
		b.WriteString(line[prev:loc[0]])
		b.Write(re.ExpandString(nil, repl, line, loc))
		prev = loc[1]
		if !global {
			break
		}
	}
	b.WriteString(line[prev:])
	return b.String()
}

// selected reports whether a command's address selects the current line,
// maintaining range state for two-address commands.
func (e *editor) selected(c *cmd, lineNo, last int, pattern string) bool {
	if c.a1 == nil {
		return true
	}
	if c.a2 == nil {
		return matchAddr(c.a1, lineNo, last, pattern)
	}
	// Two-address range.
	if !c.active {
		if matchAddr(c.a1, lineNo, last, pattern) {
			c.active = true
			// A numeric end address on or before the current line ends the
			// range immediately (single line).
			if c.a2.kind == addrLine && c.a2.line <= lineNo {
				c.active = false
			}
			return true
		}
		return false
	}
	// Already in range: this line is included; check whether it ends here.
	if matchAddr(c.a2, lineNo, last, pattern) {
		c.active = false
	}
	return true
}

// matchAddr reports whether a single address matches the current line.
func matchAddr(a *address, lineNo, last int, pattern string) bool {
	switch a.kind {
	case addrLine:
		return a.line == lineNo
	case addrLast:
		return lineNo == last
	case addrRegex:
		return a.re.MatchString(pattern)
	default:
		return false
	}
}
