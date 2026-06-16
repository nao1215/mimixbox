package pwcrack

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

// sha512crypt hash of the password "secret" with salt "abcdefgh".
const secretHash = "$6$abcdefgh$ltjgWl6579NluT/Vi1nwEvcil.G5Nbc4NiXZaNGStk8PSwGfQv72N2CKPPrVACtLtip/cZ/1GM/O6IND4WQhG."

func run(t *testing.T, stdin string, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(stdin), Out: out, Err: errBuf}
	err := New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func writeFile(t *testing.T, name, content string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return p
}

func exitCode(err error) int {
	var ee *command.ExitError
	if errors.As(err, &ee) {
		return ee.Code
	}
	return 0
}

func TestCrackHelperFindsPassword(t *testing.T) {
	t.Parallel()
	got, ok := crack(secretHash, []string{"wrong", "nope", "secret", "extra"})
	if !ok || got != "secret" {
		t.Errorf("crack = %q, %v; want secret, true", got, ok)
	}
}

func TestCrackHelperNoMatch(t *testing.T) {
	t.Parallel()
	if _, ok := crack(secretHash, []string{"a", "b", "c"}); ok {
		t.Error("did not expect a match")
	}
}

func TestCrackUnsupportedHash(t *testing.T) {
	t.Parallel()
	if _, ok := crack("not-a-crypt-hash", []string{"x"}); ok {
		t.Error("unsupported hash should not match")
	}
}

func TestRunHashOperand(t *testing.T) {
	t.Parallel()
	wl := writeFile(t, "words", "alpha\nsecret\nbeta\n")
	out, _, err := run(t, "", "-w", wl, secretHash)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if !strings.Contains(out, ": secret") {
		t.Errorf("out = %q", out)
	}
}

func TestRunHashFromStdin(t *testing.T) {
	t.Parallel()
	wl := writeFile(t, "words", "secret\n")
	out, _, err := run(t, secretHash+"\n", "-w", wl)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if !strings.Contains(out, ": secret") {
		t.Errorf("out = %q", out)
	}
}

func TestRunShadowFile(t *testing.T) {
	t.Parallel()
	wl := writeFile(t, "words", "secret\n")
	shadow := writeFile(t, "shadow", "alice:"+secretHash+":19000:0:99999:7:::\n")
	out, _, err := run(t, "", "-w", wl, "--shadow", shadow)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if !strings.Contains(out, "alice: secret") {
		t.Errorf("out = %q", out)
	}
}

func TestRunNoMatchExitsOne(t *testing.T) {
	t.Parallel()
	wl := writeFile(t, "words", "nope\n")
	out, _, err := run(t, "", "-w", wl, secretHash)
	if code := exitCode(err); code != command.ExitFailure {
		t.Errorf("exit code = %d, want %d", code, command.ExitFailure)
	}
	if !strings.Contains(out, "(not found)") {
		t.Errorf("out = %q", out)
	}
}

func TestRunNoWordlist(t *testing.T) {
	t.Parallel()
	_, _, err := run(t, "", secretHash)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "wordlist is required") {
		t.Errorf("err = %v", err)
	}
}

func TestRunMissingWordlistFile(t *testing.T) {
	t.Parallel()
	_, _, err := run(t, "", "-w", "/no/such/wordlist", secretHash)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRunNoTarget(t *testing.T) {
	t.Parallel()
	wl := writeFile(t, "words", "secret\n")
	_, _, err := run(t, "", "-w", wl)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "no hash to crack") {
		t.Errorf("err = %v", err)
	}
}

func TestNameSynopsis(t *testing.T) {
	t.Parallel()
	c := New()
	if c.Name() != "pwcrack" {
		t.Errorf("Name() = %q", c.Name())
	}
	if c.Synopsis() == "" {
		t.Error("Synopsis() is empty")
	}
}

func TestHelpSections(t *testing.T) {
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	if err := New().Run(context.Background(), io, []string{"--help"}); err != nil {
		t.Fatalf("Run --help error = %v", err)
	}
	got := out.String()
	if !strings.Contains(got, "Examples:") {
		t.Errorf("--help missing Examples section:\n%s", got)
	}
	if !strings.Contains(got, "Exit status:") {
		t.Errorf("--help missing Exit status section:\n%s", got)
	}
}
