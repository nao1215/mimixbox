// Package tr implements the tr applet: translate, squeeze, or delete characters
// read from standard input and write the result to standard output, following
// the common GNU tr semantics.
package tr

import (
	"context"
	"fmt"
	"io"
	"strings"

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
	delete     bool
	squeeze    bool
	complement bool
}

// Run executes tr.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]... SET1 [SET2]", stdio.Err)
	del := fs.BoolP("delete", "d", false, "delete characters in SET1, do not translate")
	squeeze := fs.BoolP("squeeze-repeats", "s", false,
		"replace each sequence of a repeated character that is listed in the last SET with a single occurrence of that character")
	complement := fs.BoolP("complement", "c", false, "use the complement of SET1")
	// GNU tr also spells --complement as -C.
	fs.BoolP("Complement", "C", false, "use the complement of SET1 (same as -c)")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	complementC, _ := fs.GetBool("Complement")
	opts := options{
		delete:     *del,
		squeeze:    *squeeze,
		complement: *complement || complementC,
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

	data, readErr := io.ReadAll(stdio.In)
	if readErr != nil {
		return command.Failuref("%v", readErr)
	}

	result := transform(string(data), set1, set2, opts)
	if _, err := io.WriteString(stdio.Out, result); err != nil {
		return command.Failure(err)
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

// transform applies the requested operation (delete, squeeze, translate, or a
// combination) to s and returns the result.
func transform(s string, set1, set2 []rune, opts options) string {
	in := []rune(s)

	// Resolve which runes SET1 selects, honoring -c/-C complement.
	inSet1 := membership(set1, opts.complement)

	var b strings.Builder
	b.Grow(len(in))

	switch {
	case opts.delete:
		// Delete chars in SET1, then optionally squeeze repeats from SET2.
		var del strings.Builder
		for _, r := range in {
			if !inSet1(r) {
				del.WriteRune(r)
			}
		}
		if opts.squeeze && len(set2) > 0 {
			inSet2 := membership(set2, false)
			writeSqueezed(&b, []rune(del.String()), inSet2)
		} else {
			b.WriteString(del.String())
		}
	case opts.squeeze && len(set2) == 0:
		// Squeeze only: collapse runs of characters in SET1.
		writeSqueezed(&b, in, inSet1)
	default:
		// Translate SET1 -> SET2, then optionally squeeze repeats from SET2.
		mapped := translateRunes(in, set1, set2, opts.complement)
		if opts.squeeze {
			inSet2 := membership(set2, false)
			writeSqueezed(&b, mapped, inSet2)
		} else {
			b.WriteString(string(mapped))
		}
	}
	return b.String()
}

// translateRunes maps each rune of in according to SET1 -> SET2. With the
// complement flag, every rune not in SET1 maps to the last rune of SET2 (GNU
// behavior). Otherwise the i-th rune of SET1 maps to the i-th rune of SET2; when
// SET2 is shorter, its last rune is repeated to pad it.
func translateRunes(in, set1, set2 []rune, complement bool) []rune {
	if len(set2) == 0 {
		return in
	}
	out := make([]rune, 0, len(in))
	if complement {
		inSet1 := make(map[rune]struct{}, len(set1))
		for _, r := range set1 {
			inSet1[r] = struct{}{}
		}
		last := set2[len(set2)-1]
		for _, r := range in {
			if _, ok := inSet1[r]; ok {
				out = append(out, r)
			} else {
				out = append(out, last)
			}
		}
		return out
	}

	m := make(map[rune]rune, len(set1))
	for i, r := range set1 {
		if i < len(set2) {
			m[r] = set2[i]
		} else {
			m[r] = set2[len(set2)-1]
		}
	}
	for _, r := range in {
		if to, ok := m[r]; ok {
			out = append(out, to)
		} else {
			out = append(out, r)
		}
	}
	return out
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

// writeSqueezed writes runes to b, collapsing each run of repeated characters
// for which selected reports true into a single occurrence.
func writeSqueezed(b *strings.Builder, in []rune, selected func(rune) bool) {
	var prev rune
	havePrev := false
	for _, r := range in {
		if havePrev && r == prev && selected(r) {
			continue
		}
		b.WriteRune(r)
		prev = r
		havePrev = true
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
