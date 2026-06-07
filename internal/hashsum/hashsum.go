// Package hashsum implements the shared logic behind the message-digest applets
// md5sum, sha1sum, sha256sum and sha512sum. The four commands are identical
// except for the hash they compute, so each delegates to Run here, passing the
// constructor for its hash.Hash. The output matches GNU coreutils: each line is
// "<hexdigest>  <filename>" (two spaces), with "-" used for standard input, and
// the -c/--check mode verifies a digest list, printing "<file>: OK" or
// "<file>: FAILED".
package hashsum

import (
	"bufio"
	"context"
	"fmt"
	"hash"
	"io"
	"os"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is a single message-digest applet. It is constructed by New with the
// command name, its one-line synopsis and the constructor of the hash it uses.
type Command struct {
	name     string
	synopsis string
	newHash  func() hash.Hash
}

// New returns a Command for name (e.g. "md5sum") whose digest is produced by
// newHash (e.g. md5.New). synopsis is the description shown in the applet list.
func New(name, synopsis string, newHash func() hash.Hash) *Command {
	return &Command{name: name, synopsis: synopsis, newHash: newHash}
}

// Name returns the command name.
func (c *Command) Name() string { return c.name }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return c.synopsis }

// Run executes the command. With operands it prints the digest of each FILE;
// with none it digests standard input as "-". The -c/--check flag switches to
// verifying the digest list(s) named by the operands.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]... [FILE]...", stdio.Err)
	check := fs.BoolP("check", "c", false, "read checksums from the FILEs and check them")
	// -b/--binary and -t/--text are accepted for GNU compatibility. The output
	// of a Go hash is identical for both modes, so they have no observable
	// effect here, but rejecting them would break scripts that pass them.
	_ = fs.BoolP("binary", "b", false, "read in binary mode")
	_ = fs.BoolP("text", "t", false, "read in text mode (default)")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	if *check {
		return c.checkMode(stdio, fs.Args())
	}
	return c.digestMode(stdio, fs.Args())
}

// digestMode prints "<hexdigest>  <name>" for each operand. With no operands it
// digests standard input, naming it "-". A missing file or a directory is
// reported on stderr GNU-style and turns into a silent failure that only sets
// the exit code.
func (c *Command) digestMode(stdio command.IO, names []string) error {
	if len(names) == 0 {
		names = []string{"-"}
	}
	h := c.newHash()
	var firstErr error
	for _, name := range names {
		if name != "-" {
			info, statErr := os.Stat(name)
			if statErr != nil {
				if os.IsNotExist(statErr) {
					_, _ = fmt.Fprintf(stdio.Err, "%s: %s: No such file or directory\n", c.name, name)
				} else {
					_, _ = fmt.Fprintf(stdio.Err, "%s: %s\n", c.name, command.FileError(name, statErr))
				}
				firstErr = keep(firstErr)
				continue
			}
			if info.IsDir() {
				_, _ = fmt.Fprintf(stdio.Err, "%s: %s: It is directory\n", c.name, name)
				firstErr = keep(firstErr)
				continue
			}
		}
		r, openErr := command.Open(stdio, name)
		if openErr != nil {
			_, _ = fmt.Fprintf(stdio.Err, "%s: %s\n", c.name, command.FileError(name, openErr))
			firstErr = keep(firstErr)
			continue
		}
		sum, sumErr := digest(h, r)
		_ = r.Close()
		if sumErr != nil {
			_, _ = fmt.Fprintf(stdio.Err, "%s: %s\n", c.name, command.FileError(name, sumErr))
			firstErr = keep(firstErr)
			continue
		}
		_, _ = fmt.Fprintf(stdio.Out, "%s  %s\n", sum, name)
	}
	return firstErr
}

// checkMode reads digest lists from the operands (or standard input when there
// are none) and verifies each entry, printing "<file>: OK" or "<file>: FAILED".
// A line that cannot be parsed or a file that cannot be read is reported on
// stderr and turns into a silent failure.
func (c *Command) checkMode(stdio command.IO, names []string) error {
	if len(names) == 0 {
		names = []string{"-"}
	}
	h := c.newHash()
	var firstErr error
	for _, name := range names {
		r, openErr := command.Open(stdio, name)
		if openErr != nil {
			_, _ = fmt.Fprintf(stdio.Err, "%s: %s\n", c.name, command.FileError(name, openErr))
			firstErr = keep(firstErr)
			continue
		}
		if err := c.checkList(stdio, h, r); err != nil {
			firstErr = keep(firstErr)
		}
		_ = r.Close()
	}
	return firstErr
}

// checkList verifies every "<hexdigest>  <file>" line read from r.
func (c *Command) checkList(stdio command.IO, h hash.Hash, r io.Reader) error {
	var firstErr error
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := sc.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}
		want, file, ok := parseLine(line)
		if !ok {
			_, _ = fmt.Fprintf(stdio.Err, "%s: improperly formatted checksum line\n", c.name)
			firstErr = keep(firstErr)
			continue
		}
		f, err := command.Open(stdio, file)
		if err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "%s: %s\n", c.name, command.FileError(file, err))
			firstErr = keep(firstErr)
			continue
		}
		got, sumErr := digest(h, f)
		_ = f.Close()
		if sumErr != nil {
			_, _ = fmt.Fprintf(stdio.Err, "%s: %s\n", c.name, command.FileError(file, sumErr))
			firstErr = keep(firstErr)
			continue
		}
		if got == want {
			_, _ = fmt.Fprintf(stdio.Out, "%s: OK\n", file)
		} else {
			_, _ = fmt.Fprintf(stdio.Out, "%s: FAILED\n", file)
			firstErr = keep(firstErr)
		}
	}
	if err := sc.Err(); err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "%s: %v\n", c.name, err)
		firstErr = keep(firstErr)
	}
	return firstErr
}

// parseLine splits a GNU checksum line "<hexdigest>  <file>" into its digest and
// filename. The two-space separator is the canonical form; a single run of
// blanks is also tolerated.
func parseLine(line string) (digest, file string, ok bool) {
	if d, f, found := strings.Cut(line, "  "); found {
		return d, strings.TrimPrefix(f, " "), true
	}
	if fields := strings.Fields(line); len(fields) == 2 {
		return fields[0], fields[1], true
	}
	return "", "", false
}

// digest returns the lowercase hex digest of everything read from r, resetting
// h afterwards so it can be reused for the next operand.
func digest(h hash.Hash, r io.Reader) (string, error) {
	h.Reset()
	if _, err := io.Copy(h, r); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

// keep returns existing if it is already set, otherwise a fresh silent failure,
// so the first error wins while later operands are still processed.
func keep(existing error) error {
	if existing != nil {
		return existing
	}
	return command.SilentFailure()
}
