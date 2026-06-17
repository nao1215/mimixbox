// Package textproc holds the pure text-processing logic shared by the text
// applets (cat, tac, nl, head, tail, wc). Everything here works on io.Reader /
// io.Writer or plain values and never touches the process, so each behavior is
// covered by an ordinary in-memory unit test and the applet packages stay thin.
package textproc

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
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
			// Word boundaries follow GNU wc: only a printable, non-space
			// character starts a word. White space ends the current word, while
			// control bytes and invalid UTF-8 bytes are transparent — they
			// neither start nor end a word — so "a\x01b" is one word and a lone
			// NUL/control byte is zero words (issue #953). An invalid byte is
			// reported by ReadRune as RuneError with size 1.
			invalidByte := ru == unicode.ReplacementChar && size == 1
			switch {
			case unicode.IsSpace(ru):
				inWord = false
			case !invalidByte && unicode.IsPrint(ru) && !inWord:
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
	// NumberRegexp numbers only lines matching Pattern (nl -b pBRE).
	NumberRegexp
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
	// Pattern is consulted when Style is NumberRegexp (nl -b pBRE).
	Pattern *regexp.Regexp

	// Sections enables nl's header/body/footer behavior. When false the input is
	// numbered as a single body using Style. When true the input is divided into
	// header/body/footer sections by the delimiter lines (\:\:\: header, \:\:
	// body, \: footer); each section is numbered with HeaderStyle / Style /
	// FooterStyle respectively, the delimiter lines are emitted as blank lines,
	// and the line number resets to Start at the start of each logical page (the
	// header, or a body that is not preceded by a header).
	Sections     bool
	HeaderStyle  NumberStyle
	FooterStyle  NumberStyle
	HeaderRegexp *regexp.Regexp
	FooterRegexp *regexp.Regexp
	// JoinBlankLines counts this many consecutive blank lines as a single line
	// for numbering (nl -l). Zero or one disables the grouping.
	JoinBlankLines int
}

// section identifies which logical part of the input a line belongs to.
type section int

const (
	sectionBody section = iota
	sectionHeader
	sectionFooter
)

// WriteTo copies r to w, numbering lines according to the Numberer's settings.
// It reads r one line at a time rather than slurping it all into memory, so it
// streams arbitrarily large input (a pipe or a multi-gigabyte file) in constant
// space.
func (n Numberer) WriteTo(w io.Writer, r io.Reader) error {
	br := bufio.NewReader(r)
	num := n.Start
	cur := sectionBody
	blankRun := 0 // consecutive blank lines seen so far in the current run
	for {
		chunk, err := br.ReadString('\n')
		if chunk != "" {
			body, nl := chunk, ""
			if strings.HasSuffix(chunk, "\n") {
				body, nl = chunk[:len(chunk)-1], "\n"
			}

			// A section delimiter line switches sections, resets the line number
			// to Start (every delimiter begins a fresh section, matching GNU nl
			// without -p), and is itself emitted as a blank line (never numbered).
			// Delimiters are only honored in Sections mode.
			if n.Sections {
				if sec, ok := delimiterSection(body); ok {
					num = n.Start
					cur = sec
					blankRun = 0
					if _, werr := io.WriteString(w, nl); werr != nil {
						return werr
					}
					if err == io.EOF {
						return nil
					}
					if err != nil {
						return err
					}
					continue
				}
			}

			numbered, werr := n.emitLine(w, body, nl, num, cur, &blankRun)
			if werr != nil {
				return werr
			}
			if numbered {
				num += n.Increment
			}
		}
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
	}
}

// delimiterSection reports the section a delimiter line opens. GNU nl uses the
// escape sequences "\:\:\:" (header), "\:\:" (body) and "\:" (footer), where the
// backslash-colon pair appears literally in the input.
func delimiterSection(body string) (section, bool) {
	switch body {
	case `\:\:\:`:
		return sectionHeader, true
	case `\:\:`:
		return sectionBody, true
	case `\:`:
		return sectionFooter, true
	default:
		return sectionBody, false
	}
}

// emitLine writes one numbered (or blank-padded) line and reports whether it
// consumed a line number. sec selects the section's numbering style, and
// blankRun tracks the length of the current run of blank lines so that
// JoinBlankLines can collapse N of them into a single numbered line.
func (n Numberer) emitLine(w io.Writer, body, nl string, num int, sec section, blankRun *int) (bool, error) {
	if body == "" {
		*blankRun++
	} else {
		*blankRun = 0
	}

	if n.numbered(body, sec, *blankRun) {
		if body == "" {
			// A numbered blank line ends the current run so the next blank line
			// starts a fresh group of JoinBlankLines.
			*blankRun = 0
		}
		_, err := io.WriteString(w, n.format(num)+n.Separator+body+nl)
		return true, err
	}
	prefix := ""
	if n.PadBlank {
		prefix = strings.Repeat(" ", n.Width+len(n.Separator))
	}
	_, err := io.WriteString(w, prefix+body+nl)
	return false, err
}

// styleFor returns the numbering style and regexp pattern that govern the given
// section.
func (n Numberer) styleFor(sec section) (NumberStyle, *regexp.Regexp) {
	switch sec {
	case sectionHeader:
		return n.HeaderStyle, n.HeaderRegexp
	case sectionFooter:
		return n.FooterStyle, n.FooterRegexp
	default:
		return n.Style, n.Pattern
	}
}

// numbered reports whether the given line is numbered. blankRun is the count of
// consecutive blank lines ending at this line (1 for the first blank); with
// JoinBlankLines set to N a blank line is numbered only on every Nth blank.
func (n Numberer) numbered(body string, sec section, blankRun int) bool {
	style, pattern := n.styleFor(sec)
	switch style {
	case NumberAll:
		if body == "" && n.JoinBlankLines > 1 {
			return blankRun%n.JoinBlankLines == 0
		}
		return true
	case NumberNonBlank:
		return body != ""
	case NumberRegexp:
		return pattern != nil && pattern.MatchString(body)
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

// HeadLines writes the first n lines of r to w, preserving line endings.
func HeadLines(w io.Writer, r io.Reader, n int) error {
	return HeadRecords(w, r, n, '\n')
}

// HeadRecords writes the first n records of r to w, preserving each record's
// trailing delimiter. With delim '\n' this is the line-oriented head; with
// delim '\0' (the -z/--zero-terminated mode) records are NUL-delimited and any
// embedded newlines are kept verbatim.
func HeadRecords(w io.Writer, r io.Reader, n int, delim byte) error {
	if n <= 0 {
		return nil
	}
	br := bufio.NewReader(r)
	for i := 0; i < n; i++ {
		s, err := br.ReadString(delim)
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
	return TailRecords(w, r, n, '\n')
}

// TailRecords writes the last n records of r to w, preserving each record's
// trailing delimiter. With delim '\n' this is the line-oriented tail; with
// delim '\0' (the -z/--zero-terminated mode) records are NUL-delimited and any
// embedded newlines are kept verbatim.
func TailRecords(w io.Writer, r io.Reader, n int, delim byte) error {
	if n <= 0 {
		return nil
	}
	br := bufio.NewReader(r)
	ring := make([]string, 0, n)
	for {
		s, err := br.ReadString(delim)
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
