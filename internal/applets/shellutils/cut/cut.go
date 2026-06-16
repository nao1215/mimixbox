// Package cut implements the cut applet: remove sections from each line of
// files (or standard input) and print the result on the standard output, with
// the common GNU options for selecting bytes, characters, or fields.
package cut

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the cut applet.
type Command struct{}

// New returns a cut command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "cut" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Remove sections from each line of files" }

// mode is the selection unit chosen by exactly one of -b, -c or -f.
type mode int

const (
	modeNone mode = iota
	modeBytes
	modeChars
	modeFields
)

// options holds the resolved, validated cut configuration.
type options struct {
	mode            mode
	ranges          []rng
	delimiter       string
	outputDelimiter string
	onlyDelimited   bool
	complement      bool
	lineDelim       byte
}

// Run executes cut.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "OPTION... [FILE]...", stdio.Err).WithHelp(command.Help{
		Description: "Print selected parts of each line from each FILE to standard output. " +
			"Exactly one of -b (bytes), -c (characters), or -f (fields) selects the unit; " +
			"with no FILE, or when FILE is -, read standard input.",
		Examples: []command.Example{
			{Command: "cut -d: -f1 /etc/passwd", Explain: "Print the first colon-separated field of each line."},
			{Command: "cut -c1-5 file.txt", Explain: "Print the first five characters of each line."},
			{Command: "cut -f2,4 -d, data.csv", Explain: "Print the second and fourth comma-separated fields."},
		},
		ExitStatus: "0  success.\n1  an error occurred (e.g. an input file could not be read).",
	})
	bytesList := fs.StringP("bytes", "b", "", "select only these bytes")
	charsList := fs.StringP("characters", "c", "", "select only these characters")
	fieldsList := fs.StringP("fields", "f", "", "select only these fields")
	delimiter := fs.StringP("delimiter", "d", "\t", "use DELIM instead of TAB for field delimiter")
	onlyDelimited := fs.BoolP("only-delimited", "s", false, "do not print lines not containing delimiters")
	outputDelimiter := fs.String("output-delimiter", "", "use STRING as the output delimiter")
	complement := fs.Bool("complement", false, "complement the set of selected bytes, characters or fields")
	zeroTerminated := fs.BoolP("zero-terminated", "z", false, "line delimiter is NUL, not newline")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	opts, err := buildOptions(stdio, *bytesList, *charsList, *fieldsList,
		*delimiter, *outputDelimiter, *onlyDelimited, fs.Changed("output-delimiter"),
		*complement, *zeroTerminated)
	if err != nil {
		return err
	}

	return run(stdio, opts, fs.Args())
}

// buildOptions validates the mutually exclusive selection flags and parses the
// range list, returning the resolved options. Errors are written to stderr and
// reported as silent failures so the runner only sets the exit code.
func buildOptions(stdio command.IO, bytesList, charsList, fieldsList,
	delimiter, outputDelimiter string, onlyDelimited, outDelimSet, complement, zeroTerminated bool) (options, error) {
	m, list, err := selectMode(stdio, bytesList, charsList, fieldsList)
	if err != nil {
		return options{}, err
	}

	ranges, perr := parseRanges(list)
	if perr != nil {
		_, _ = fmt.Fprintf(stdio.Err, "cut: %v\n", perr)
		return options{}, command.SilentFailure()
	}

	outDelim := outputDelimiter
	if !outDelimSet {
		// Default output delimiter: the input delimiter for fields, empty
		// otherwise (bytes/characters are simply concatenated).
		if m == modeFields {
			outDelim = delimiter
		} else {
			outDelim = ""
		}
	}

	lineDelim := byte('\n')
	if zeroTerminated {
		lineDelim = 0
	}

	return options{
		mode:            m,
		ranges:          ranges,
		delimiter:       delimiter,
		outputDelimiter: outDelim,
		onlyDelimited:   onlyDelimited,
		complement:      complement,
		lineDelim:       lineDelim,
	}, nil
}

// selectMode enforces that exactly one of -b/-c/-f is given and returns the
// chosen mode together with its range list string.
func selectMode(stdio command.IO, bytesList, charsList, fieldsList string) (mode, string, error) {
	count := 0
	var m mode
	var list string
	if bytesList != "" {
		count++
		m, list = modeBytes, bytesList
	}
	if charsList != "" {
		count++
		m, list = modeChars, charsList
	}
	if fieldsList != "" {
		count++
		m, list = modeFields, fieldsList
	}

	switch {
	case count == 0:
		_, _ = fmt.Fprintln(stdio.Err, "cut: you must specify a list of bytes, characters, or fields")
		return modeNone, "", command.SilentFailure()
	case count > 1:
		_, _ = fmt.Fprintln(stdio.Err, "cut: only one type of list may be specified")
		return modeNone, "", command.SilentFailure()
	default:
		return m, list, nil
	}
}

// run reads every operand (defaulting to standard input) and cuts each line.
func run(stdio command.IO, opts options, files []string) error {
	if len(files) == 0 {
		files = []string{"-"}
	}
	var firstErr error
	for _, name := range files {
		r, err := command.Open(stdio, name)
		if err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "cut: %s\n", command.FileError(name, err))
			firstErr = keep(firstErr)
			continue
		}
		cerr := cutReader(stdio.Out, r, opts)
		_ = r.Close()
		if cerr != nil {
			_, _ = fmt.Fprintf(stdio.Err, "cut: %s\n", command.FileError(name, cerr))
			firstErr = keep(firstErr)
		}
	}
	return firstErr
}

// cutReader processes r line by line, writing the cut output to w. Lines are
// split on opts.lineDelim ('\n' by default, NUL with -z) and the same delimiter
// terminates each emitted line.
func cutReader(w io.Writer, r io.Reader, opts options) error {
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	sc.Split(splitOn(opts.lineDelim))
	bw := bufio.NewWriter(w)
	for sc.Scan() {
		out, emit := cutLine(sc.Text(), opts)
		if !emit {
			continue
		}
		if _, err := bw.WriteString(out); err != nil {
			return err
		}
		if err := bw.WriteByte(opts.lineDelim); err != nil {
			return err
		}
	}
	if err := sc.Err(); err != nil {
		return err
	}
	return bw.Flush()
}

// splitOn returns a bufio.SplitFunc that splits input on the byte delim,
// stripping the delimiter from each returned token. A trailing chunk without a
// final delimiter is still returned as the last token.
func splitOn(delim byte) bufio.SplitFunc {
	return func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}
		if i := bytes.IndexByte(data, delim); i >= 0 {
			return i + 1, data[:i], nil
		}
		if atEOF {
			return len(data), data, nil
		}
		return 0, nil, nil
	}
}

// rng is a 1-based inclusive range from a cut list. A zero hi means the range
// is open-ended ("N-").
type rng struct {
	lo int
	hi int // 0 means "to end of line"
}

// open reports whether the range extends to the end of the line.
func (r rng) open() bool { return r.hi == 0 }

// parseRanges parses a comma-separated cut LIST such as "1,3-5,7-" into a
// sorted, merged slice of ranges. Each item is one of N, N-, N-M or -M, all
// 1-based and inclusive.
func parseRanges(list string) ([]rng, error) {
	var ranges []rng
	for _, item := range strings.Split(list, ",") {
		item = strings.TrimSpace(item)
		if item == "" {
			return nil, fmt.Errorf("invalid byte, character or field list")
		}
		r, err := parseRange(item)
		if err != nil {
			return nil, err
		}
		ranges = append(ranges, r)
	}
	return mergeRanges(ranges), nil
}

// parseRange parses a single list item.
func parseRange(item string) (rng, error) {
	dash := strings.IndexByte(item, '-')
	if dash < 0 {
		n, err := parsePos(item)
		if err != nil {
			return rng{}, err
		}
		return rng{lo: n, hi: n}, nil
	}

	loStr := item[:dash]
	hiStr := item[dash+1:]

	var lo, hi int
	var err error
	if loStr == "" {
		lo = 1 // "-M" means from the first
	} else if lo, err = parsePos(loStr); err != nil {
		return rng{}, err
	}

	if hiStr == "" {
		hi = 0 // "N-" means to the end
	} else if hi, err = parsePos(hiStr); err != nil {
		return rng{}, err
	}

	if hi != 0 && lo > hi {
		return rng{}, fmt.Errorf("invalid decreasing range")
	}
	return rng{lo: lo, hi: hi}, nil
}

// parsePos parses a positive (>= 1) integer position from a list.
func parsePos(s string) (int, error) {
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("invalid byte, character or field list")
	}
	if n < 1 {
		return 0, fmt.Errorf("byte, character or field positions are numbered from 1")
	}
	return n, nil
}

// mergeRanges sorts ranges by their start and merges overlapping or adjacent
// ones so the selection is emitted in ascending order without duplication.
func mergeRanges(ranges []rng) []rng {
	if len(ranges) == 0 {
		return ranges
	}
	sort.Slice(ranges, func(i, j int) bool {
		if ranges[i].lo != ranges[j].lo {
			return ranges[i].lo < ranges[j].lo
		}
		// Open-ended ranges sort last among equal starts.
		return !ranges[i].open() && ranges[j].open()
	})

	merged := []rng{ranges[0]}
	for _, r := range ranges[1:] {
		last := &merged[len(merged)-1]
		if last.open() {
			// Nothing can extend past an open-ended range.
			continue
		}
		if r.lo <= last.hi+1 {
			if r.open() {
				last.hi = 0
			} else if r.hi > last.hi {
				last.hi = r.hi
			}
			continue
		}
		merged = append(merged, r)
	}
	return merged
}

// selected reports whether the 1-based position pos is covered by any range.
func selected(ranges []rng, pos int) bool {
	for _, r := range ranges {
		if pos < r.lo {
			return false // ranges are sorted; no later range can match a smaller pos
		}
		if r.open() || pos <= r.hi {
			return true
		}
	}
	return false
}

// keep reports whether the 1-based position pos should be emitted, honouring
// --complement (which inverts the range selection).
func keepPos(opts options, pos int) bool {
	in := selected(opts.ranges, pos)
	if opts.complement {
		return !in
	}
	return in
}

// cutLine applies the selection to a single input line (without its trailing
// newline). The second result reports whether the line should be emitted at
// all (false only for -s lines that contain no delimiter).
func cutLine(line string, opts options) (string, bool) {
	switch opts.mode {
	case modeFields:
		return cutFields(line, opts)
	case modeChars:
		return cutRunes(line, opts), true
	default: // modeBytes
		return cutBytes(line, opts), true
	}
}

// cutBytes selects bytes from line. Contiguous kept positions are emitted as a
// run; the output delimiter separates non-adjacent runs. --complement inverts
// the kept set.
func cutBytes(line string, opts options) string {
	var b strings.Builder
	prevKept := false
	for i := 0; i < len(line); i++ {
		if !keepPos(opts, i+1) {
			prevKept = false
			continue
		}
		if !prevKept && b.Len() > 0 {
			b.WriteString(opts.outputDelimiter)
		}
		b.WriteByte(line[i])
		prevKept = true
	}
	return b.String()
}

// cutRunes selects characters (runes) from line, mirroring cutBytes but over
// runes so multibyte characters stay intact.
func cutRunes(line string, opts options) string {
	runes := []rune(line)
	var b strings.Builder
	prevKept := false
	for i := range runes {
		if !keepPos(opts, i+1) {
			prevKept = false
			continue
		}
		if !prevKept && b.Len() > 0 {
			b.WriteString(opts.outputDelimiter)
		}
		b.WriteRune(runes[i])
		prevKept = true
	}
	return b.String()
}

// cutFields selects fields from line split on the delimiter. Lines with no
// delimiter are passed through unchanged unless -s suppresses them.
// --complement inverts the selected fields.
func cutFields(line string, opts options) (string, bool) {
	if !strings.Contains(line, opts.delimiter) {
		if opts.onlyDelimited {
			return "", false
		}
		return line, true
	}
	fields := strings.Split(line, opts.delimiter)
	var selectedFields []string
	for i := range fields {
		if keepPos(opts, i+1) {
			selectedFields = append(selectedFields, fields[i])
		}
	}
	return strings.Join(selectedFields, opts.outputDelimiter), true
}

func keep(existing error) error {
	if existing != nil {
		return existing
	}
	return command.SilentFailure()
}
