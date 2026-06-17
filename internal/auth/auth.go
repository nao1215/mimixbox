// Package auth provides basic authentication: verifying a user's password the
// way a login-style command would.
//
// MimixBox does not support PAM. The shadow backend in shadow.go verifies the
// password against the crypt(3) hash in /etc/shadow, which needs no cgo and so
// keeps the default, fully static MimixBox build (CGO_ENABLED=0) intact.
//
// Callers use Authenticate and stay unaware of how the backend is implemented.
package auth

import "errors"

// ErrNoBackend is returned when no authentication backend was compiled in.
var ErrNoBackend = errors.New("auth: no authentication backend configured")

// backend is set by the shadow backend's init (shadow.go).
var backend func(user, password string) (bool, error)

// Authenticate reports whether password is correct for user. A false result
// with a nil error means the credentials simply did not match; a non-nil error
// means authentication could not be performed (for example /etc/shadow could
// not be read, which usually means the caller is not root).
func Authenticate(user, password string) (bool, error) {
	if backend == nil {
		return false, ErrNoBackend
	}
	return backend(user, password)
}
