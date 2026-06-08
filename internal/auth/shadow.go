//go:build !pam

// Package auth's default backend: verify a password against the crypt(3) hash
// stored in /etc/shadow. It needs no cgo, so it works in the default static
// build and on any Linux system without PAM.
package auth

import (
	"bufio"
	"io"
	"os"
	"strings"

	"github.com/GehirnInc/crypt"
	_ "github.com/GehirnInc/crypt/apr1_crypt"   // register $apr1$
	_ "github.com/GehirnInc/crypt/md5_crypt"    // register $1$
	_ "github.com/GehirnInc/crypt/sha256_crypt" // register $5$
	_ "github.com/GehirnInc/crypt/sha512_crypt" // register $6$
)

// shadowPath is the file the shadow backend reads; tests point it elsewhere.
var shadowPath = "/etc/shadow"

func init() { backend = shadowAuthenticate }

// shadowAuthenticate verifies password for user against shadowPath.
func shadowAuthenticate(user, password string) (bool, error) {
	f, err := os.Open(shadowPath) //nolint:gosec // /etc/shadow is the file we must read
	if err != nil {
		return false, err
	}
	defer func() { _ = f.Close() }()

	hash, found, err := lookupHash(f, user)
	if err != nil {
		return false, err
	}
	if !found {
		return false, nil
	}
	return verifyHash(hash, password), nil
}

// lookupHash returns the stored password hash for user from a shadow stream.
func lookupHash(r io.Reader, user string) (hash string, found bool, err error) {
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		fields := strings.SplitN(sc.Text(), ":", 3)
		if len(fields) >= 2 && fields[0] == user {
			return fields[1], true, nil
		}
	}
	return "", false, sc.Err()
}

// verifyHash reports whether password matches the shadow hash field. Empty,
// locked ("!"/"*") or otherwise non-crypt fields never authenticate.
func verifyHash(hash, password string) bool {
	if hash == "" || strings.HasPrefix(hash, "!") || strings.HasPrefix(hash, "*") {
		return false
	}
	if !crypt.IsHashSupported(hash) {
		return false
	}
	return crypt.NewFromHash(hash).Verify(hash, []byte(password)) == nil
}
