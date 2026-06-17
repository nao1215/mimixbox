package auth

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// sha512crypt hash of "secret" with salt "abcdefgh".
const secretHash = "$6$abcdefgh$ltjgWl6579NluT/Vi1nwEvcil.G5Nbc4NiXZaNGStk8PSwGfQv72N2CKPPrVACtLtip/cZ/1GM/O6IND4WQhG."

func withShadow(t *testing.T, content string) {
	t.Helper()
	p := filepath.Join(t.TempDir(), "shadow")
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	orig := shadowPath
	shadowPath = p
	t.Cleanup(func() { shadowPath = orig })
}

func TestAuthenticateSuccess(t *testing.T) {
	withShadow(t, "alice:"+secretHash+":19000:0:99999:7:::\n")
	ok, err := Authenticate("alice", "secret")
	if err != nil {
		t.Fatalf("Authenticate error = %v", err)
	}
	if !ok {
		t.Error("expected authentication to succeed")
	}
}

func TestAuthenticateWrongPassword(t *testing.T) {
	withShadow(t, "alice:"+secretHash+":19000:0:99999:7:::\n")
	ok, err := Authenticate("alice", "wrong")
	if err != nil {
		t.Fatalf("Authenticate error = %v", err)
	}
	if ok {
		t.Error("expected authentication to fail for a wrong password")
	}
}

func TestAuthenticateUnknownUser(t *testing.T) {
	withShadow(t, "alice:"+secretHash+":19000:0:99999:7:::\n")
	ok, err := Authenticate("bob", "secret")
	if err != nil {
		t.Fatalf("Authenticate error = %v", err)
	}
	if ok {
		t.Error("unknown user must not authenticate")
	}
}

func TestLockedAccount(t *testing.T) {
	withShadow(t, "alice:!:19000:0:99999:7:::\nbob:*:19000:0:99999:7:::\n")
	for _, u := range []string{"alice", "bob"} {
		ok, err := Authenticate(u, "secret")
		if err != nil {
			t.Fatalf("Authenticate(%s) error = %v", u, err)
		}
		if ok {
			t.Errorf("locked account %s must not authenticate", u)
		}
	}
}

func TestEmptyHashField(t *testing.T) {
	withShadow(t, "alice::19000:0:99999:7:::\n")
	ok, _ := Authenticate("alice", "")
	if ok {
		t.Error("empty hash field must not authenticate")
	}
}

func TestUnsupportedHash(t *testing.T) {
	withShadow(t, "alice:plaintextnotcrypt:19000::::\n")
	ok, err := Authenticate("alice", "plaintextnotcrypt")
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	if ok {
		t.Error("a non-crypt hash field must not authenticate")
	}
}

func TestMissingShadowFile(t *testing.T) {
	orig := shadowPath
	shadowPath = "/no/such/shadow/file"
	t.Cleanup(func() { shadowPath = orig })

	if _, err := Authenticate("alice", "secret"); err == nil {
		t.Fatal("expected an error when the shadow file cannot be read")
	}
}

func TestLookupHash(t *testing.T) {
	r := strings.NewReader("root:$6$x$h:1::\nalice:$1$y$z:2::\n")
	hash, found, err := lookupHash(r, "alice")
	if err != nil || !found || hash != "$1$y$z" {
		t.Errorf("lookupHash = %q, %v, %v", hash, found, err)
	}
}

func TestVerifyHashSchemes(t *testing.T) {
	// md5crypt of "secret" with salt "abcdefgh".
	md5 := "$1$abcdefgh$cHJi5PXp/ki/ktXzqlk6I1"
	if !verifyHash(md5, "secret") {
		t.Error("md5crypt of secret should verify")
	}
	if verifyHash(md5, "nope") {
		t.Error("wrong password should not verify")
	}
}
