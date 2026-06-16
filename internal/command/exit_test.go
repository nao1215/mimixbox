package command_test

import (
	"context"
	"errors"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func TestExitErrorErrorWithWrappedErr(t *testing.T) {
	t.Parallel()
	e := &command.ExitError{Code: 7, Err: errors.New("disk full")}
	if got := e.Error(); got != "disk full" {
		t.Errorf("Error() = %q, want %q", got, "disk full")
	}
}

func TestExitErrorErrorWithoutWrappedErr(t *testing.T) {
	t.Parallel()
	e := &command.ExitError{Code: 5}
	if got := e.Error(); got != "exit status 5" {
		t.Errorf("Error() = %q, want %q", got, "exit status 5")
	}
}

func TestExitErrorUnwrap(t *testing.T) {
	t.Parallel()
	sentinel := errors.New("root cause")
	e := &command.ExitError{Code: command.ExitFailure, Err: sentinel}
	if !errors.Is(e, sentinel) {
		t.Errorf("errors.Is should find the wrapped sentinel via Unwrap")
	}
	if errors.Unwrap(e) != sentinel {
		t.Errorf("Unwrap() = %v, want %v", errors.Unwrap(e), sentinel)
	}
}

func TestExitErrorUnwrapNil(t *testing.T) {
	t.Parallel()
	e := &command.ExitError{Code: command.ExitFailure}
	if errors.Unwrap(e) != nil {
		t.Errorf("Unwrap() = %v, want nil", errors.Unwrap(e))
	}
}

func TestFailure(t *testing.T) {
	t.Parallel()
	cause := errors.New("boom")
	e := command.Failure(cause)
	if e.Code != command.ExitFailure {
		t.Errorf("Code = %d, want %d", e.Code, command.ExitFailure)
	}
	if e.Err != cause {
		t.Errorf("Err = %v, want %v", e.Err, cause)
	}
	if got := e.Error(); got != "boom" {
		t.Errorf("Error() = %q, want %q", got, "boom")
	}
	if !errors.Is(e, cause) {
		t.Errorf("Failure result should unwrap to its cause")
	}
}

func TestFailuref(t *testing.T) {
	t.Parallel()
	e := command.Failuref("cannot open %s: code %d", "file.txt", 13)
	if e.Code != command.ExitFailure {
		t.Errorf("Code = %d, want %d", e.Code, command.ExitFailure)
	}
	if got := e.Error(); got != "cannot open file.txt: code 13" {
		t.Errorf("Error() = %q, want %q", got, "cannot open file.txt: code 13")
	}
}

func TestFailurefWrapsErrorVerb(t *testing.T) {
	t.Parallel()
	cause := errors.New("permission denied")
	e := command.Failuref("open: %w", cause)
	if !errors.Is(e, cause) {
		t.Errorf("Failuref with %%w should let errors.Is find the cause")
	}
}

func TestSilentFailureErrorIsEmpty(t *testing.T) {
	t.Parallel()
	// SilentFailure carries an exit code but no message, because the underlying
	// machinery already wrote to stderr; its Error() must therefore be empty.
	err := command.SilentFailure()
	if err == nil {
		t.Fatal("SilentFailure() returned nil")
	}
	if got := err.Error(); got != "" {
		t.Errorf("SilentFailure().Error() = %q, want empty string", got)
	}
}

func TestSilentFailureMapsToFailureCode(t *testing.T) {
	t.Parallel()
	// Execute is the single place that maps a silent failure to its exit code
	// while printing nothing extra to stderr.
	io, _, errBuf := newIO()
	code := command.Execute(context.Background(), stub{
		name: "demo",
		run: func(context.Context, command.IO, []string) error {
			return command.SilentFailure()
		},
	}, io, nil)
	if code != command.ExitFailure {
		t.Errorf("exit code = %d, want %d", code, command.ExitFailure)
	}
	if errBuf.Len() != 0 {
		t.Errorf("stderr = %q, want empty (message already written)", errBuf.String())
	}
}
