package command

import (
	"context"
	"errors"
	"fmt"
)

// Execute runs c with the given IO and arguments and returns the process exit
// code. It is the single place that turns a command's error into an exit code
// and a "name: message" line, so the behavior is identical in production and
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
			_, _ = fmt.Fprintf(io.Err, "%s: %s\n", c.Name(), exit.Err)
		}
		return exit.Code
	}

	_, _ = fmt.Fprintf(io.Err, "%s: %s\n", c.Name(), err)
	return ExitFailure
}
