package date

import "time"

// SetNow replaces the clock with a function returning fixed, restoring the
// previous clock when the returned function is called. It exists so the
// external date_test package can make output deterministic.
func SetNow(fixed time.Time) (restore func()) {
	prev := nowFn
	nowFn = func() time.Time { return fixed }
	return func() { nowFn = prev }
}

// FormatTime exposes the pure strftime formatter for testing.
func FormatTime(t time.Time, format string) string { return formatTime(t, format) }
