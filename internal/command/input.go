package command

import (
	"errors"
	"fmt"
	"io"
	"os"
)

// MaxLineSize is the largest single line (token) the line-oriented applets
// accept when reading through a bufio.Scanner. The default scanner cap is only
// 64 KiB, which made applets like rev, fold, cut, grep, sed, and awk reject
// valid long lines with "token too long" while their GNU counterparts handled
// them. Pass this as the scanner's max so a line is bounded only by available
// memory, matching the host tools:
//
//	sc := bufio.NewScanner(r)
//	sc.Buffer(make([]byte, 0, 64*1024), command.MaxLineSize)
const MaxLineSize = 1 << 30 // 1 GiB

// Open resolves an operand to a reader the way the GNU file utilities do: the
// name "-" means standard input (taken from the injected IO so tests stay in
// memory), and any other name is opened as a file. The caller must Close the
// result; closing the stdin wrapper is a no-op.
func Open(stdio IO, name string) (io.ReadCloser, error) {
	if name == "-" {
		return io.NopCloser(stdio.In), nil
	}
	return os.Open(name) //nolint:gosec // operating on a user-named file is the whole point
}

// FileError formats a failed file operation GNU-style as "name: reason". An
// os.PathError repeats the operation and path ("open foo: no such file..."),
// so it is unwrapped to just the underlying reason, leaving the caller to add
// the command-name prefix (e.g. "cat: foo: no such file or directory").
func FileError(name string, err error) string {
	var pe *os.PathError
	if errors.As(err, &pe) {
		return fmt.Sprintf("%s: %v", name, pe.Err)
	}
	return fmt.Sprintf("%s: %v", name, err)
}
