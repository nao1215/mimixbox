// Package textproc holds the pure text-processing logic shared by the text
// applets (cat, tac, nl, head, tail, wc). Everything here works on io.Reader /
// io.Writer or plain values and never touches the process, so each behaviour is
// covered by an ordinary in-memory unit test and the applet packages stay thin.
package textproc

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"unicode"
)

// Count is the tally that wc reports for a single input. The zero value is a
// valid empty count, and Counts combine with Add so a "total" line is just the
// running sum of the per-file counts.
type Count struct {
	Lines        int // number of newline characters
	Words        int // whitespace-separated words
	Runes        int // characters (wc -m)
	Bytes        int // bytes (wc -c)
	MaxLineWidth int // length in runes of the longest line (wc -L)
}

// Add returns the element-wise sum of two Counts, taking the maximum for
// MaxLineWidth so totals stay meaningful.
func (c Count) Add(o Count) Count {
	c.Lines += o.Lines
	c.Words += o.Words
	c.Runes += o.Runes
	c.Bytes += o.Bytes
	if o.MaxLineWidth > c.MaxLineWidth {
		c.MaxLineWidth = o.MaxLineWidth
	}
	return c
}

// CountReader reads r to completion and reports its Count.
func CountReader(r io.Reader) (Count, error) {
	var c Count
	br := bufio.NewReader(r)
	inWord := false
	lineWidth := 0
	for {
		ru, size, err := br.ReadRune()
		if size > 0 {
			c.Bytes += size
			c.Runes++
			if ru == '\n' {
				c.Lines++
				if lineWidth > c.MaxLineWidth {
					c.MaxLineWidth = lineWidth
				}
				lineWidth = 0
			} else {
				lineWidth++
			}
			if unicode.IsSpace(ru) {
				inWord = false
			} else if !inWord {
				inWord = true
				c.Words++
			}
		}
		if err != nil {
			if err == io.EOF {
				if lineWidth > c.MaxLineWidth {
					c.MaxLineWidth = lineWidth
				}
				return c, nil
			}
			return c, err
		}
	}
}

// Reverse returns text with its records in reverse order, the way tac does.
// Records are delimited by sep, which is treated as a trailing separator: the
// input "a\nb\nc" reverses to "cb\na\n" because only the first two records own a
// newline. This matches GNU tac (without the --before flag).
func Reverse(text, sep string) string {
	if text == "" || sep == "" {
		return text
	}
	var records []string
	for text != "" {
		i := strings.Index(text, sep)
		if i < 0 {
			records = append(records, text)
			break
		}
		records = append(records, text[:i+len(sep)])
		text = text[i+len(sep):]
	}
	var b strings.Builder
	for i := len(records) - 1; i >= 0; i-- {
		b.WriteString(records[i])
	}
	return b.String()
}

// NumberStyle selects which lines a Numberer numbers.
type NumberStyle int

const (
	// NumberNone numbers no lines (nl -bn).
	NumberNone NumberStyle = iota
	// NumberAll numbers every line (cat -n, nl -ba).
	NumberAll
	// NumberNonBlank numbers only non-empty lines (cat -b, nl -bt).
	NumberNonBlank
)

// NumberJustify selects how the line number is padded inside its field.
type NumberJustify int

const (
	// JustifyRight right-justifies with spaces (nl -n rn).
	JustifyRight NumberJustify = iota
	// JustifyRightZero right-justifies with leading zeros (nl -n rz).
	JustifyRightZero
	// JustifyLeft left-justifies with spaces (nl -n ln).
	JustifyLeft
)

// Numberer writes line-numbered text, covering both cat (-n/-b) and nl. The
// zero value is not useful; build one with the fields set. Cat uses
// PadBlank=false (an unnumbered line is emitted verbatim); nl uses
// PadBlank=true (an unnumbered line is left-padded so columns line up).
type Numberer struct {
	Style     NumberStyle
	Start     int
	Increment int
	Width     int
	Separator string
	Justify   NumberJustify
	PadBlank  bool
}

// WriteTo copies r to w, numbering lines according to the Numberer's settings.
func (n Numberer) WriteTo(w io.Writer, r io.Reader) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	num := n.Start
	for _, line := range splitKeepNewline(string(data)) {
		body, nl := line.body, line.newline
		if n.numbered(body) {
			if _, err := io.WriteString(w, n.format(num)+n.Separator+body+nl); err != nil {
				return err
			}
			num += n.Increment
			continue
		}
		prefix := ""
		if n.PadBlank {
			prefix = strings.Repeat(" ", n.Width+len(n.Separator))
		}
		if _, err := io.WriteString(w, prefix+body+nl); err != nil {
			return err
		}
	}
	return nil
}

func (n Numberer) numbered(body string) bool {
	switch n.Style {
	case NumberAll:
		return true
	case NumberNonBlank:
		return body != ""
	default:
		return false
	}
}

func (n Numberer) format(num int) string {
	switch n.Justify {
	case JustifyRightZero:
		return fmt.Sprintf("%0*d", n.Width, num)
	case JustifyLeft:
		return fmt.Sprintf("%-*d", n.Width, num)
	default:
		return fmt.Sprintf("%*d", n.Width, num)
	}
}

type line struct {
	body    string
	newline string
}

// splitKeepNewline splits s into lines, recording for each whether it ended
// with a newline so the original line endings can be reproduced exactly.
func splitKeepNewline(s string) []line {
	var lines []line
	for s != "" {
		i := strings.IndexByte(s, '\n')
		if i < 0 {
			lines = append(lines, line{body: s})
			break
		}
		lines = append(lines, line{body: s[:i], newline: "\n"})
		s = s[i+1:]
	}
	return lines
}

// HeadLines writes the first n lines of r to w, preserving line endings.
func HeadLines(w io.Writer, r io.Reader, n int) error {
	if n <= 0 {
		return nil
	}
	br := bufio.NewReader(r)
	for i := 0; i < n; i++ {
		s, err := br.ReadString('\n')
		if s != "" {
			if _, werr := io.WriteString(w, s); werr != nil {
				return werr
			}
		}
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
	}
	return nil
}

// TailLines writes the last n lines of r to w, preserving line endings.
func TailLines(w io.Writer, r io.Reader, n int) error {
	if n <= 0 {
		return nil
	}
	br := bufio.NewReader(r)
	ring := make([]string, 0, n)
	for {
		s, err := br.ReadString('\n')
		if s != "" {
			if len(ring) < n {
				ring = append(ring, s)
			} else {
				copy(ring, ring[1:])
				ring[n-1] = s
			}
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
	}
	for _, s := range ring {
		if _, werr := io.WriteString(w, s); werr != nil {
			return werr
		}
	}
	return nil
}

// HeadBytes writes the first n bytes of r to w.
func HeadBytes(w io.Writer, r io.Reader, n int) error {
	if n <= 0 {
		return nil
	}
	_, err := io.CopyN(w, r, int64(n))
	if err == io.EOF {
		return nil
	}
	return err
}

// TailBytes writes the last n bytes of r to w.
func TailBytes(w io.Writer, r io.Reader, n int) error {
	if n <= 0 {
		return nil
	}
	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	if n > len(data) {
		n = len(data)
	}
	_, err = w.Write(data[len(data)-n:])
	return err
}
