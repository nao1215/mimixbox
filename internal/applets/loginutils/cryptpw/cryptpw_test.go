package cryptpw

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

func TestStdinDefaultSHA512(t *testing.T) {
	const want = "$6$abcdefgh$ltjgWl6579NluT/Vi1nwEvcil.G5Nbc4NiXZaNGStk8PSwGfQv72N2CKPPrVACtLtip/cZ/1GM/O6IND4WQhG."
	got, err := run(t, "secret\n", "-S", "abcdefgh")
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Errorf("cryptpw =\n%s\nwant\n%s", got, want)
	}
}

func TestOperandOverridesStdin(t *testing.T) {
	got, err := run(t, "ignored\n", "-m", "md5", "-S", "abcdefgh", "secret")
	if err != nil {
		t.Fatal(err)
	}
	if got != "$1$abcdefgh$cHJi5PXp/ki/ktXzqlk6I1" {
		t.Errorf("operand password hash = %s", got)
	}
}

func TestVerifies(t *testing.T) {
	got, err := run(t, "", "-S", "saltsalt", "hunter2")
	if err != nil {
		t.Fatal(err)
	}
	if crypt.NewFromHash(got).Verify(got, []byte("hunter2")) != nil {
		t.Errorf("hash does not verify")
	}
}

func TestErrors(t *testing.T) {
	if _, err := run(t, "", "-m", "bogus", "x"); err == nil {
		t.Errorf("unknown method should fail")
	}
	if _, err := run(t, ""); err == nil {
		t.Errorf("no password should fail")
	}
}
