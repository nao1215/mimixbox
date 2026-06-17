// Package tr implements the tr applet: translate, squeeze, or delete characters
// read from standard input and write the result to standard output, following
// the common GNU tr semantics.
package tr

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"unicode/utf8"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the tr applet.
type Command struct{}

// New returns a tr command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "tr" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Translate or delete characters" }

type options struct {
	delete       bool
	squeeze      bool
	complement   bool
	truncateSet1 bool
}

// Run executes tr.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]... SET1 [SET2]", stdio.Err).WithHelp(command.Help{
		Description: "Translate, squeeze, or delete characters from standard input, writing the result to standard output. " +
			"SET1 and SET2 may use ranges (a-z), classes ([:upper:]), and C-style escapes.",
		Examples: []command.Example{
			{Command: "tr a-z A-Z", Explain: "Translate lowercase letters to uppercase."},
			{Command: "tr -d '0-9'", Explain: "Delete every digit from the input."},
			{Command: "tr -s ' '", Explain: "Squeeze repeated spaces into a single space."},
		},
		ExitStatus: "0  the input was translated successfully.\n1  an operand was missing or invalid, or input could not be read.",
	})
	del := fs.BoolP("delete", "d", false, "delete characters in SET1, do not translate")
	squeeze := fs.BoolP("squeeze-repeats", "s", false,
		"replace each sequence of a repeated character that is listed in the last SET with a single occurrence of that character")
	complement := fs.BoolP("complement", "c", false, "use the complement of SET1")
	// GNU tr also spells --complement as -C.
	fs.BoolP("Complement", "C", false, "use the complement of SET1 (same as -c)")
	truncate := fs.BoolP("truncate-set1", "t", false, "first truncate SET1 to length of SET2")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	complementC, _ := fs.GetBool("Complement")
	opts := options{
		delete:       *del,
		squeeze:      *squeeze,
		complement:   *complement || complementC,
		truncateSet1: *truncate,
	}

	operands := fs.Args()
	if usageErr := validate(stdio, opts, operands); usageErr != nil {
		return usageErr
	}

	set1, err := expandSet(operands[0])
	if err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "tr: %v\n", err)
		return command.SilentFailure()
	}
	var set2 []rune
	if len(operands) > 1 {
		set2, err = expandSet(operands[1])
		if err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "tr: %v\n", err)
			return command.SilentFailure()
		}
	}

	// With --truncate-set1 (-t), SET1 is truncated to the length of SET2 before
	// translating, so any SET1 characters beyond SET2's length are passed through
	// unchanged. GNU applies this only when translating (SET2 present).
	if opts.truncateSet1 && len(set2) > 0 && len(set1) > len(set2) {
		set1 = set1[:len(set2)]
	}

	if err := streamTransform(stdio.In, stdio.Out, set1, set2, opts); err != nil {
		return command.Failuref("%v", err)
	}
	return nil
}

// validate checks the operands the way GNU tr does and, on a usage problem,
// prints a GNU-style message to stderr and returns a silent failure.
func validate(stdio command.IO, opts options, operands []string) error {
	if len(operands) == 0 {
		_, _ = fmt.Fprintln(stdio.Err, "tr: missing operand")
		_, _ = fmt.Fprintln(stdio.Err, "Try 'tr --help' for more information.")
		return command.SilentFailure()
	}
	// In translate mode (no -d, no -s) SET2 is mandatory.
	if !opts.delete && !opts.squeeze && len(operands) < 2 {
		_, _ = fmt.Fprintf(stdio.Err, "tr: missing operand after '%s'\n", operands[0])
		_, _ = fmt.Fprintln(stdio.Err, "Try 'tr --help' for more information.")
		return command.SilentFailure()
	}
	if len(operands) > 2 {
		_, _ = fmt.Fprintf(stdio.Err, "tr: extra operand '%s'\n", operands[2])
		_, _ = fmt.Fprintln(stdio.Err, "Try 'tr --help' for more information.")
		return command.SilentFailure()
	}
	return nil
}

// cell is one decoded input symbol. A valid UTF-8 sequence becomes a normal
// rune (raw=false); an invalid byte is preserved verbatim as a raw byte
// (raw=true, r in 0..255). Keeping the raw flag lets tr behave like GNU tr on
// binary input — matching SET entries by byte value and writing untranslated
// bytes back unchanged instead of corrupting them into U+FFFD (issue #953).
type cell struct {
	r   rune
	raw bool
}

// decodeCells splits raw input into cells, preserving invalid UTF-8 bytes.
func decodeCells(data []byte) []cell {
	cells := make([]cell, 0, len(data))
	for i := 0; i < len(data); {
		r, size := utf8.DecodeRune(data[i:])
		if r == utf8.RuneError && size == 1 {
			cells = append(cells, cell{r: rune(data[i]), raw: true})
			i++
			continue
		}
		cells = append(cells, cell{r: r, raw: false})
		i += size
	}
	return cells
}

// appendCell writes a cell to out: a raw byte as itself, a normal rune as UTF-8.
func appendCell(out []byte, c cell) []byte {
	if c.raw {
		return append(out, byte(c.r))
	}
	return utf8.AppendRune(out, c.r)
}

// streamTransform reads from r, applies the tr operation, and writes the result
// to w, processing the input in chunks so the whole input is never buffered
// (issue #952). Multi-byte UTF-8 sequences that straddle a chunk boundary are
// carried to the next read, and the squeeze run state is preserved across
// chunks, so the output is identical to processing the whole input at once.
func streamTransform(r io.Reader, w io.Writer, set1, set2 []rune, opts options) error {
	p := newProcessor(set1, set2, opts)
	bw := bufio.NewWriter(w)
	buf := make([]byte, 64*1024)
	out := make([]byte, 0, 64*1024)
	var carry []byte
	for {
		n, rerr := r.Read(buf)
		atEOF := rerr != nil

		chunk := buf[:n]
		if len(carry) > 0 {
			chunk = append(carry, chunk...)
		}

		proc := len(chunk)
		if !atEOF {
			// Hold back an incomplete trailing UTF-8 sequence; at EOF it is
			// flushed and decoded as raw bytes, matching whole-buffer decoding.
			proc = completeRuneLen(chunk)
		}
		out = p.process(out[:0], decodeCells(chunk[:proc]))
		if len(out) > 0 {
			if _, werr := bw.Write(out); werr != nil {
				return werr
			}
		}
		if !atEOF {
			carry = append([]byte(nil), chunk[proc:]...)
		}

		if atEOF {
			if rerr == io.EOF {
				break
			}
			return rerr
		}
	}
	return bw.Flush()
}

// completeRuneLen returns the length of the longest prefix of b that does not
// end in the middle of a multi-byte UTF-8 sequence. An invalid byte counts as a
// complete (width-1) symbol, matching utf8.FullRune.
func completeRuneLen(b []byte) int {
	for i := 1; i <= utf8.UTFMax && i <= len(b); i++ {
		start := len(b) - i
		if utf8.RuneStart(b[start]) {
			if utf8.FullRune(b[start:]) {
				return len(b)
			}
			return start
		}
	}
	return len(b)
}

// processor applies the tr operation to a stream of cells. The only state it
// carries between chunks is the squeeze run (prev rune); a buffering version
// kept this implicitly while walking the whole input.
type processor struct {
	opts   options
	set2   []rune
	inSet1 func(rune) bool
	inSet2 func(rune) bool

	translate  bool
	transMap   map[rune]rune
	transComp  map[rune]struct{}
	transLast  rune
	complement bool

	prev     rune
	havePrev bool
}

// newProcessor precomputes the membership predicates and translation table for
// the requested operation.
func newProcessor(set1, set2 []rune, opts options) *processor {
	p := &processor{
		opts:       opts,
		set2:       set2,
		inSet1:     membership(set1, opts.complement),
		complement: opts.complement,
	}
	switch {
	case opts.delete:
		if opts.squeeze && len(set2) > 0 {
			p.inSet2 = membership(set2, false)
		}
	case opts.squeeze && len(set2) == 0:
		// squeeze-only collapses runs of SET1 directly via inSet1.
	default:
		if len(set2) > 0 {
			p.translate = true
			if opts.complement {
				p.transComp = make(map[rune]struct{}, len(set1))
				for _, r := range set1 {
					p.transComp[r] = struct{}{}
				}
				p.transLast = set2[len(set2)-1]
			} else {
				p.transMap = make(map[rune]rune, len(set1))
				for i, r := range set1 {
					if i < len(set2) {
						p.transMap[r] = set2[i]
					} else {
						p.transMap[r] = set2[len(set2)-1]
					}
				}
			}
		}
		if opts.squeeze {
			p.inSet2 = membership(set2, false)
		}
	}
	return p
}

// process applies the operation to cells, appending output to out.
func (p *processor) process(out []byte, cells []cell) []byte {
	for _, c := range cells {
		switch {
		case p.opts.delete:
			if p.inSet1(c.r) {
				continue
			}
			if p.opts.squeeze && len(p.set2) > 0 {
				out = p.squeezeEmit(out, c, p.inSet2)
			} else {
				out = appendCell(out, c)
			}
		case p.opts.squeeze && len(p.set2) == 0:
			out = p.squeezeEmit(out, c, p.inSet1)
		default:
			c = p.translateCell(c)
			if p.opts.squeeze {
				out = p.squeezeEmit(out, c, p.inSet2)
			} else {
				out = appendCell(out, c)
			}
		}
	}
	return out
}

// translateCell maps one cell according to SET1 -> SET2.
func (p *processor) translateCell(c cell) cell {
	if !p.translate {
		return c
	}
	if p.complement {
		if _, ok := p.transComp[c.r]; ok {
			return c
		}
		return c.mapped(p.transLast)
	}
	if to, ok := p.transMap[c.r]; ok {
		return c.mapped(to)
	}
	return c
}

// squeezeEmit appends c unless it repeats the previous selected character, in
// which case it is collapsed away. The run state persists across chunks.
func (p *processor) squeezeEmit(out []byte, c cell, selected func(rune) bool) []byte {
	if p.havePrev && c.r == p.prev && selected(c.r) {
		return out
	}
	out = appendCell(out, c)
	p.prev = c.r
	p.havePrev = true
	return out
}

// mapped returns c translated to rune to, keeping the raw-byte representation
// when the result still fits in a single byte (so a translated binary byte is
// written as a byte, not as multi-byte UTF-8).
func (c cell) mapped(to rune) cell {
	return cell{r: to, raw: c.raw && to <= 0xFF}
}

// membership returns a predicate reporting whether a rune is selected by set.
// When complement is true the predicate is negated.
func membership(set []rune, complement bool) func(rune) bool {
	m := make(map[rune]struct{}, len(set))
	for _, r := range set {
		m[r] = struct{}{}
	}
	return func(r rune) bool {
		_, ok := m[r]
		if complement {
			return !ok
		}
		return ok
	}
}

// expandSet expands a tr SET specification into its sequence of runes,
// resolving escapes (\n, \t, \\, octal \nnn), ranges (a-z), and character
// classes ([:upper:], [:lower:], [:digit:], [:space:], [:alpha:], [:alnum:]).
func expandSet(spec string) ([]rune, error) {
	src := []rune(spec)
	var out []rune

	for i := 0; i < len(src); {
		// Character class: [:name:]
		if src[i] == '[' && i+1 < len(src) && src[i+1] == ':' {
			end := indexClassEnd(src, i)
			if end >= 0 {
				name := string(src[i+2 : end])
				class, ok := classRunes(name)
				if !ok {
					return nil, fmt.Errorf("unknown character class %q", name)
				}
				out = append(out, class...)
				i = end + 2 // skip past ":]"
				continue
			}
		}

		// Parse one character (possibly an escape).
		r, ni, err := nextRune(src, i)
		if err != nil {
			return nil, err
		}

		// Range: a-z
		if ni < len(src) && src[ni] == '-' && ni+1 < len(src) {
			hi, nj, err := nextRune(src, ni+1)
			if err != nil {
				return nil, err
			}
			if hi < r {
				return nil, fmt.Errorf("range-endpoints of '%c-%c' are in reverse collating sequence order", r, hi)
			}
			for c := r; c <= hi; c++ {
				out = append(out, c)
			}
			i = nj
			continue
		}

		out = append(out, r)
		i = ni
	}
	return out, nil
}

// indexClassEnd returns the index of the ':' that closes a [: ... :] class
// starting at start, or -1 if there is no proper close.
func indexClassEnd(src []rune, start int) int {
	for j := start + 2; j+1 < len(src); j++ {
		if src[j] == ':' && src[j+1] == ']' {
			return j
		}
	}
	return -1
}

// nextRune decodes the character at position i, handling C and octal escapes,
// and returns the rune plus the index just past it.
func nextRune(src []rune, i int) (rune, int, error) {
	if src[i] != '\\' {
		return src[i], i + 1, nil
	}
	if i+1 >= len(src) {
		// A trailing backslash stands for itself.
		return '\\', i + 1, nil
	}
	c := src[i+1]
	switch c {
	case 'n':
		return '\n', i + 2, nil
	case 't':
		return '\t', i + 2, nil
	case 'r':
		return '\r', i + 2, nil
	case 'f':
		return '\f', i + 2, nil
	case 'v':
		return '\v', i + 2, nil
	case 'b':
		return '\b', i + 2, nil
	case 'a':
		return '\a', i + 2, nil
	case '\\':
		return '\\', i + 2, nil
	}
	if c >= '0' && c <= '7' {
		// Octal escape: up to three octal digits.
		val := rune(0)
		j := i + 1
		for k := 0; k < 3 && j < len(src) && src[j] >= '0' && src[j] <= '7'; k++ {
			val = val*8 + (src[j] - '0')
			j++
		}
		return val, j, nil
	}
	// Unknown escape: the backslash is dropped, the next char is literal.
	return c, i + 2, nil
}

// classRunes returns the runes belonging to a POSIX character class.
func classRunes(name string) ([]rune, bool) {
	switch name {
	case "upper":
		return rangeRunes('A', 'Z'), true
	case "lower":
		return rangeRunes('a', 'z'), true
	case "digit":
		return rangeRunes('0', '9'), true
	case "alpha":
		return append(rangeRunes('A', 'Z'), rangeRunes('a', 'z')...), true
	case "alnum":
		r := rangeRunes('0', '9')
		r = append(r, rangeRunes('A', 'Z')...)
		r = append(r, rangeRunes('a', 'z')...)
		return r, true
	case "space":
		return []rune{'\t', '\n', '\v', '\f', '\r', ' '}, true
	case "blank":
		return []rune{'\t', ' '}, true
	default:
		return nil, false
	}
}

func rangeRunes(lo, hi rune) []rune {
	out := make([]rune, 0, hi-lo+1)
	for c := lo; c <= hi; c++ {
		out = append(out, c)
	}
	return out
}
