// Package dd implements the dd applet: convert and copy a file. Unlike the
// other applets, dd takes its arguments as key=value operands (if=, of=, bs=,
// conv=, ...) rather than getopt-style flags, so it parses them by hand.
package dd

import (
	"context"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// zeroReader is an io.Reader that yields an endless stream of zero bytes. It is
// used to seek on a non-seekable writer (stdout) by streaming the zero padding
// instead of allocating it all up front.
type zeroReader struct{}

func (zeroReader) Read(p []byte) (int, error) {
	for i := range p {
		p[i] = 0
	}
	return len(p), nil
}

// Command is the dd applet.
type Command struct{}

// New returns a dd command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "dd" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Convert and copy a file" }

// statusMode controls how much of the transfer summary dd writes to stderr.
type statusMode int

const (
	statusDefault  statusMode = iota // full summary
	statusNone                       // suppress everything
	statusNoxfer                     // suppress the final "N bytes" transfer line
	statusProgress                   // like default (progress reporting is best-effort)
)

// options holds the parsed dd operands.
type options struct {
	inputFile  string // if=, empty means stdin
	outputFile string // of=, empty means stdout
	ibs        int64  // input block size
	obs        int64  // output block size
	count      int64  // copy only this many input blocks; -1 means all
	skip       int64  // skip this many ibs-sized blocks of input
	seek       int64  // skip this many obs-sized blocks of output
	lcase      bool   // conv=lcase
	ucase      bool   // conv=ucase
	notrunc    bool   // conv=notrunc
	sync       bool   // conv=sync: pad each input block to ibs with NUL
	status     statusMode
}

// defaultOptions returns options seeded with dd's defaults.
func defaultOptions() options {
	return options{
		ibs:    512,
		obs:    512,
		count:  -1,
		skip:   0,
		seek:   0,
		status: statusDefault,
	}
}

// result records the block counts and byte total of a copy, used both to build
// the summary and to let tests assert on the work performed.
type result struct {
	inFull   int64 // complete input blocks read
	inPart   int64 // partial input blocks read
	outFull  int64 // complete output blocks written
	outPart  int64 // partial output blocks written
	bytesOut int64 // total bytes written
}

// Run executes dd.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	if command.HandleHelpVersionWith(stdio, c.Name(), "[OPERAND]...", command.Help{
		Description: "Copy a file, converting and formatting according to the operands. Operands use " +
			"the GNU dd OPERAND=VALUE syntax: if=FILE, of=FILE, bs=BYTES, count=N, skip=N, seek=N, " +
			"and conv=CONVS. With no if=/of= it copies standard input to standard output.",
		Examples: []command.Example{
			{Command: "dd if=/dev/zero of=out bs=1M count=10", Explain: "Write 10 MiB of zeros to 'out'."},
			{Command: "dd if=disk.img bs=512 skip=1", Explain: "Copy disk.img to stdout, skipping the first 512-byte block."},
		},
		ExitStatus: "0  the copy completed successfully.\n1  an operand was invalid or an I/O error occurred.",
	}, args) {
		return nil
	}
	opts, err := parseArgs(args)
	if err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "dd: %v\n", err)
		return command.SilentFailure()
	}

	in, closeIn, err := openInput(stdio, opts)
	if err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "dd: %s\n", command.FileError(opts.inputFile, err))
		return command.SilentFailure()
	}
	defer closeIn()

	out, closeOut, err := openOutput(stdio, opts)
	if err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "dd: %s\n", command.FileError(opts.outputFile, err))
		return command.SilentFailure()
	}
	defer closeOut()

	res, err := copyDD(in, out, opts)
	if err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "dd: %v\n", err)
		return command.Failure(err)
	}

	writeSummary(stdio.Err, res, opts)
	return nil
}

// openInput resolves the if= operand to a reader. Skipping is applied here.
func openInput(stdio command.IO, opts options) (io.Reader, func(), error) {
	var r io.Reader
	closeFn := func() {}
	if opts.inputFile == "" {
		r = stdio.In
	} else {
		f, err := os.Open(opts.inputFile) //nolint:gosec // operating on a user-named file is the point
		if err != nil {
			return nil, closeFn, err
		}
		r = f
		closeFn = func() { _ = f.Close() }
	}
	if opts.skip > 0 {
		if _, err := io.CopyN(io.Discard, r, opts.skip*opts.ibs); err != nil && err != io.EOF {
			closeFn()
			return nil, func() {}, err
		}
	}
	return r, closeFn, nil
}

// openOutput resolves the of= operand to a writer. Seeking is applied here.
func openOutput(stdio command.IO, opts options) (io.Writer, func(), error) {
	if opts.outputFile == "" {
		w := stdio.Out
		if opts.seek > 0 {
			if _, err := io.CopyN(w, zeroReader{}, opts.seek*opts.obs); err != nil {
				return nil, func() {}, err
			}
		}
		return w, func() {}, nil
	}

	flag := os.O_WRONLY | os.O_CREATE
	if !opts.notrunc {
		flag |= os.O_TRUNC
	}
	f, err := os.OpenFile(opts.outputFile, flag, 0o644) //nolint:gosec // user-named file
	if err != nil {
		return nil, func() {}, err
	}
	if opts.seek > 0 {
		if _, err := f.Seek(opts.seek*opts.obs, io.SeekStart); err != nil {
			_ = f.Close()
			return nil, func() {}, err
		}
	}
	return f, func() { _ = f.Close() }, nil
}

// copyDD performs the core copy from r to w according to opts. It is pure with
// respect to its reader and writer, so a test can drive it with bytes.Buffer.
func copyDD(r io.Reader, w io.Writer, opts options) (result, error) {
	var res result
	buf := make([]byte, opts.ibs)

	for opts.count < 0 || res.inFull+res.inPart < opts.count {
		n, err := io.ReadFull(r, buf)
		if n == 0 {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				break
			}
			if err != nil {
				return res, err
			}
		}

		block := buf[:n]
		if int64(n) == opts.ibs {
			res.inFull++
		} else if n > 0 {
			res.inPart++
			if opts.sync {
				padded := make([]byte, opts.ibs)
				copy(padded, block)
				block = padded
			}
		}

		block = applyConv(block, opts)

		if werr := writeBlock(w, block, opts, &res); werr != nil {
			return res, werr
		}

		if err == io.EOF || err == io.ErrUnexpectedEOF {
			break
		}
		if err != nil && err != io.ErrUnexpectedEOF {
			return res, err
		}
	}
	return res, nil
}

// writeBlock writes a single converted block, recording the output counts.
func writeBlock(w io.Writer, block []byte, opts options, res *result) error {
	if len(block) == 0 {
		return nil
	}
	n, err := w.Write(block)
	res.bytesOut += int64(n)
	if n == len(block) && int64(len(block)) == opts.obs {
		res.outFull++
	} else if n > 0 {
		res.outPart++
	}
	return err
}

// applyConv applies the byte-level conv transforms (lcase/ucase) to a block and
// returns the transformed bytes. It is pure and unit-testable.
func applyConv(block []byte, opts options) []byte {
	if !opts.lcase && !opts.ucase {
		return block
	}
	out := make([]byte, len(block))
	for i, b := range block {
		switch {
		case opts.ucase && b >= 'a' && b <= 'z':
			out[i] = b - ('a' - 'A')
		case opts.lcase && b >= 'A' && b <= 'Z':
			out[i] = b + ('a' - 'A')
		default:
			out[i] = b
		}
	}
	return out
}

// writeSummary prints the standard dd records-in/records-out summary to w
// unless status=none suppresses it.
func writeSummary(w io.Writer, res result, opts options) {
	if opts.status == statusNone {
		return
	}
	_, _ = fmt.Fprintf(w, "%d+%d records in\n", res.inFull, res.inPart)
	_, _ = fmt.Fprintf(w, "%d+%d records out\n", res.outFull, res.outPart)
	if opts.status != statusNoxfer {
		_, _ = fmt.Fprintf(w, "%d bytes copied\n", res.bytesOut)
	}
}

// parseArgs parses dd's key=value operands into options.
func parseArgs(args []string) (options, error) {
	opts := defaultOptions()
	bsSet := false

	for _, arg := range args {
		key, value, ok := strings.Cut(arg, "=")
		if !ok {
			return opts, fmt.Errorf("unrecognized operand %q", arg)
		}
		switch key {
		case "if":
			opts.inputFile = value
		case "of":
			opts.outputFile = value
		case "bs":
			n, err := parseSize(value)
			if err != nil {
				return opts, fmt.Errorf("invalid bs=%q: %w", value, err)
			}
			opts.ibs = n
			opts.obs = n
			bsSet = true
		case "ibs":
			n, err := parseSize(value)
			if err != nil {
				return opts, fmt.Errorf("invalid ibs=%q: %w", value, err)
			}
			if !bsSet {
				opts.ibs = n
			}
		case "obs":
			n, err := parseSize(value)
			if err != nil {
				return opts, fmt.Errorf("invalid obs=%q: %w", value, err)
			}
			if !bsSet {
				opts.obs = n
			}
		case "count":
			n, err := parseSize(value)
			if err != nil {
				return opts, fmt.Errorf("invalid count=%q: %w", value, err)
			}
			opts.count = n
		case "skip":
			n, err := parseSize(value)
			if err != nil {
				return opts, fmt.Errorf("invalid skip=%q: %w", value, err)
			}
			opts.skip = n
		case "seek":
			n, err := parseSize(value)
			if err != nil {
				return opts, fmt.Errorf("invalid seek=%q: %w", value, err)
			}
			opts.seek = n
		case "conv":
			if err := applyConvFlags(&opts, value); err != nil {
				return opts, err
			}
		case "status":
			if err := applyStatus(&opts, value); err != nil {
				return opts, err
			}
		default:
			return opts, fmt.Errorf("unrecognized operand %q", arg)
		}
	}

	if opts.ibs <= 0 {
		return opts, fmt.Errorf("invalid input block size %d", opts.ibs)
	}
	if opts.obs <= 0 {
		return opts, fmt.Errorf("invalid output block size %d", opts.obs)
	}
	return opts, nil
}

// applyConvFlags parses a comma-separated conv= list into opts.
func applyConvFlags(opts *options, value string) error {
	for _, c := range strings.Split(value, ",") {
		switch c {
		case "lcase":
			opts.lcase = true
		case "ucase":
			opts.ucase = true
		case "notrunc":
			opts.notrunc = true
		case "sync":
			opts.sync = true
		case "":
			// ignore empty entries
		default:
			return fmt.Errorf("unknown conversion %q", c)
		}
	}
	if opts.lcase && opts.ucase {
		return fmt.Errorf("cannot combine conv=lcase and conv=ucase")
	}
	return nil
}

// applyStatus parses the status= operand.
func applyStatus(opts *options, value string) error {
	switch value {
	case "none":
		opts.status = statusNone
	case "noxfer":
		opts.status = statusNoxfer
	case "progress":
		opts.status = statusProgress
	default:
		return fmt.Errorf("unknown status %q", value)
	}
	return nil
}

// ParseSize is the exported entry point for parseSize, used by tests in the
// external dd_test package.
func ParseSize(s string) (int64, error) { return parseSize(s) }

// parseSize parses a dd block-size string such as "512", "1k", or "2M" into a
// byte count. Recognized suffixes: c=1, w=2, b=512, k/K=1024, M=1024*1024,
// G=1024*1024*1024. It is pure and unit-testable.
func parseSize(s string) (int64, error) {
	if s == "" {
		return 0, fmt.Errorf("empty size")
	}

	var mult int64 = 1
	switch s[len(s)-1] {
	case 'c':
		mult, s = 1, s[:len(s)-1]
	case 'w':
		mult, s = 2, s[:len(s)-1]
	case 'b':
		mult, s = 512, s[:len(s)-1]
	case 'k', 'K':
		mult, s = 1024, s[:len(s)-1]
	case 'M':
		mult, s = 1024*1024, s[:len(s)-1]
	case 'G':
		mult, s = 1024*1024*1024, s[:len(s)-1]
	}

	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("not a number")
	}
	if n < 0 {
		return 0, fmt.Errorf("negative size")
	}
	if mult > 1 && n > math.MaxInt64/mult {
		return 0, fmt.Errorf("size too large (overflow)")
	}
	return n * mult, nil
}
