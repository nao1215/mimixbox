//go:build pam

// Package auth's PAM backend slot. Building with `-tags pam` selects this file
// instead of shadow.go. It is intentionally a stub: a real PAM backend needs
// cgo and the libpam development files, which the default static MimixBox build
// deliberately avoids. Dropping a cgo implementation here (conversation
// callback calling pam_start/pam_authenticate/pam_end) is the documented way to
// add PAM support without touching any caller, since everything still goes
// through auth.Authenticate.
package auth

import "errors"

// errPAMNotBuilt is returned until a real cgo PAM implementation replaces this
// stub. It keeps the package compiling under `-tags pam` so the build-time
// switch is real and callers can rely on the seam.
var errPAMNotBuilt = errors.New("auth: PAM backend not built (provide a cgo libpam implementation in pam.go)")

func init() {
	backend = func(_, _ string) (bool, error) { return false, errPAMNotBuilt }
}
