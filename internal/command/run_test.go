package command_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func TestExecuteExitErrorWithNilErrPrintsNothing(t *testing.T) {
	t.Parallel()
	// An ExitError with a custom code but no message must report the code while
	// printing nothing: the "exit.Err != nil" guard is false here.
	io, _, errBuf := newIO()
	code := command.Execute(context.Background(), stub{
		name: "demo",
		run: func(context.Context, command.IO, []string) error {
			return &command.ExitError{Code: 42}
		},
	}, io, nil)
	if code != 42 {
		t.Errorf("exit code = %d, want 42", code)
	}
	if errBuf.Len() != 0 {
		t.Errorf("stderr = %q, want empty when ExitError has no message", errBuf.String())
	}
}

func TestExecuteWrappedExitErrorIsUnwrapped(t *testing.T) {
	t.Parallel()
	// Execute uses errors.As, so an ExitError wrapped by fmt.Errorf still drives
	// the exit code and the wrapped message is printed with the command name.
	io, _, errBuf := newIO()
	code := command.Execute(context.Background(), stub{
		name: "demo",
		run: func(context.Context, command.IO, []string) error {
			return fmt.Errorf("context: %w", &command.ExitError{Code: 9, Err: errors.New("inner")})
		},
	}, io, nil)
	if code != 9 {
		t.Errorf("exit code = %d, want 9", code)
	}
	if errBuf.String() != "demo: inner\n" {
		t.Errorf("stderr = %q, want %q", errBuf.String(), "demo: inner\n")
	}
}

func TestExecuteWrappedSilentFailureIsUnwrapped(t *testing.T) {
	t.Parallel()
	// A silent failure wrapped by fmt.Errorf is still detected via errors.As and
	// maps to its code without printing anything extra.
	io, _, errBuf := newIO()
	code := command.Execute(context.Background(), stub{
		name: "demo",
		run: func(context.Context, command.IO, []string) error {
			return fmt.Errorf("wrap: %w", command.SilentFailure())
		},
	}, io, nil)
	if code != command.ExitFailure {
		t.Errorf("exit code = %d, want %d", code, command.ExitFailure)
	}
	if errBuf.Len() != 0 {
		t.Errorf("stderr = %q, want empty for silent failure", errBuf.String())
	}
}
