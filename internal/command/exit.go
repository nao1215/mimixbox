package command

import "fmt"

// Exit codes shared by every command. Commands should return an *ExitError (or
// a plain error, which maps to ExitFailure) rather than calling os.Exit, so the
// runner stays in control and tests can assert on the code.
const (
	ExitSuccess = 0
	ExitFailure = 1
)

// ExitError lets a command choose the process exit code while still returning a
// normal error value. Code is the status to exit with; Err, when set, is the
// message printed by the runner (prefixed with the command name).
type ExitError struct {
	Code int
	Err  error
}

// Error implements the error interface.
func (e *ExitError) Error() string {
	if e.Err == nil {
		return fmt.Sprintf("exit status %d", e.Code)
	}
	return e.Err.Error()
}

// Unwrap exposes the wrapped error to errors.Is / errors.As.
func (e *ExitError) Unwrap() error { return e.Err }

// Failure wraps err with ExitFailure (status 1).
func Failure(err error) *ExitError {
	return &ExitError{Code: ExitFailure, Err: err}
}

// Failuref is Failure with a formatted message.
func Failuref(format string, a ...any) *ExitError {
	return &ExitError{Code: ExitFailure, Err: fmt.Errorf(format, a...)}
}

// silent is an error that carries an exit code but no message, used when the
// underlying machinery (for example pflag) has already written to stderr.
type silent struct{ code int }

func (s silent) Error() string { return "" }

// SilentFailure reports a usage failure whose message was already printed.
func SilentFailure() error { return silent{code: ExitFailure} }
