// Package auth provides cross-platform basic authentication: verifying a user's
// password the way a login-style command would.
//
// The backend is selected at build time so the default, fully static MimixBox
// build (CGO_ENABLED=0) never needs PAM:
//
//   - default build (no build tag): the shadow backend in shadow.go verifies
//     the password against the crypt(3) hash in /etc/shadow.
//   - `-tags pam`: a PAM backend (pam.go) takes over, for systems that want
//     Pluggable Authentication Modules. Because PAM needs cgo and the libpam
//     development files, it is gated behind the build tag so its absence can
//     never break the default build.
//
// Callers use Authenticate and stay unaware of which backend is compiled in.
package auth

import "errors"

// ErrNoBackend is returned when no authentication backend was compiled in.
var ErrNoBackend = errors.New("auth: no authentication backend configured")

// backend is set by the compiled-in backend's init (shadow.go or pam.go).
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
