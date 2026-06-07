// Package command defines the small framework that MimixBox applets are built
// on. Every applet is a Command: it receives its input and output as
// dependencies and its arguments as a slice, so it can be exercised entirely in
// memory by a unit test. The package also bridges Commands to the legacy applet
// table in internal/applets, so migrated and not-yet-migrated applets coexist.
package command

import (
	"context"
	"io"
)

// IO carries the streams a command reads from and writes to. Injecting these
// instead of touching os.Stdin/os.Stdout directly is what makes commands
// testable: a test passes bytes.Buffer values and inspects the result.
type IO struct {
	In  io.Reader
	Out io.Writer
	Err io.Writer
}

// Command is a single MimixBox applet such as cat or wc.
type Command interface {
	// Name is the command name as a user types it (e.g. "cat").
	Name() string
	// Synopsis is the one-line description shown in the applet list.
	Synopsis() string
	// Run executes the command. args are the arguments after the command
	// name (os.Args[1:]). A nil return means success; see ExitError to
	// control the process exit code.
	Run(ctx context.Context, io IO, args []string) error
}
