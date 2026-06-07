package command

import (
	"context"
	"errors"
	"fmt"
	"os"
)

// Execute runs c with the given IO and arguments and returns the process exit
// code. It is the single place that turns a command's error into an exit code
// and a "name: message" line, so the behaviour is identical in production and
// in tests. args are the arguments after the command name (os.Args[1:]).
func Execute(ctx context.Context, c Command, io IO, args []string) int {
	err := c.Run(ctx, io, args)
	if err == nil {
		return ExitSuccess
	}

	// A silent failure already wrote its message (e.g. a flag parse error).
	var s silent
	if errors.As(err, &s) {
		return s.code
	}

	var exit *ExitError
	if errors.As(err, &exit) {
		if exit.Err != nil {
			fmt.Fprintf(io.Err, "%s: %s\n", c.Name(), exit.Err)
		}
		return exit.Code
	}

	fmt.Fprintf(io.Err, "%s: %s\n", c.Name(), err)
	return ExitFailure
}

// Adapt bridges a Command to the legacy applet entry-point signature used by
// internal/applets. It wires the command to the process streams and reports the
// exit code; the error is always nil because Execute has already printed any
// message, so the caller need only exit with the returned code.
func Adapt(c Command) func() (int, error) {
	return func() (int, error) {
		stdio := IO{In: os.Stdin, Out: os.Stdout, Err: os.Stderr}
		return Execute(context.Background(), c, stdio, os.Args[1:]), nil
	}
}
