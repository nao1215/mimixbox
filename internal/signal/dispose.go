package signal

import (
	"os"
	"os/signal"
)

// IgnoreDuring sets the given signals to SIG_IGN for the duration of fn, then
// restores their prior process-wide disposition before returning. It is the
// single abstraction for the temporary signal-disposition change used by
// child-process-launching applets such as env (--ignore-signal) and nohup
// (which ignores SIGHUP): a launched child inherits the SIG_IGN disposition,
// while the parent process is left exactly as it was found.
//
// The restore runs via defer, so the prior disposition is reinstated on every
// exit path — normal return, an error from fn, and a panic from fn. When
// signals is empty, fn is run without touching any disposition. The error
// returned by fn (if any) is propagated unchanged.
func IgnoreDuring(signals []os.Signal, fn func() error) error {
	if len(signals) > 0 {
		signal.Ignore(signals...)
		defer signal.Reset(signals...)
	}
	return fn()
}
