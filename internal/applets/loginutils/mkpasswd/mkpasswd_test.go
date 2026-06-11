package mkpasswd

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/GehirnInc/crypt"

	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, stdin string, args ...string) (string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(stdin), Out: out, Err: &bytes.Buffer{}}
	err := New().Run(context.Background(), io, args)
	return strings.TrimSpace(out.String()), err
}

func TestSHA512MatchesReference(t *testing.T) {
	// Identical to `openssl passwd -6 -salt abcdefgh secret`.
	const want = "$6$abcdefgh$ltjgWl6579NluT/Vi1nwEvcil.G5Nbc4NiXZaNGStk8PSwGfQv72N2CKPPrVACtLtip/cZ/1GM/O6IND4WQhG."
	got, err := run(t, "", "-m", "sha-512", "-S", "abcdefgh", "secret")
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Errorf("sha-512 hash =\n%s\nwant\n%s", got, want)
	}
}

func TestMD5MatchesReference(t *testing.T) {
	const want = "$1$abcdefgh$cHJi5PXp/ki/ktXzqlk6I1"
	got, err := run(t, "", "-m", "md5", "-S", "abcdefgh", "secret")
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Errorf("md5 hash = %s, want %s", got, want)
	}
}

func TestHashVerifies(t *testing.T) {
	got, err := run(t, "", "-S", "saltsalt", "hunter2")
	if err != nil {
		t.Fatal(err)
	}
	if crypt.NewFromHash(got).Verify(got, []byte("hunter2")) != nil {
		t.Errorf("generated hash does not verify against the password")
	}
}

func TestStdinPassword(t *testing.T) {
	got, err := run(t, "frompipe\n", "-m", "sha-256", "-S", "abcdefgh")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(got, "$5$abcdefgh$") {
		t.Errorf("stdin hash = %s", got)
	}
}

func TestErrors(t *testing.T) {
	if _, err := run(t, "", "-m", "bogus", "secret"); err == nil {
		t.Errorf("an unknown method should fail")
	}
	if _, err := run(t, ""); err == nil { // empty stdin, no operand
		t.Errorf("no password should fail")
	}
}
