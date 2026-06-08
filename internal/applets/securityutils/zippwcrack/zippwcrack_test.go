package zippwcrack

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

// encryptedZipB64 is a ZIP archive holding data.txt ("secret content here\n"),
// encrypted with traditional ZipCrypto using the password "hunter2". It was
// produced by the system `zip -P hunter2`.
const encryptedZipB64 = "UEsDBAoACQAAAPeJyFzf7VBXIAAAABQAAAAIABwAZGF0YS50eHRVVAkAAzF6JmoxeiZqdXgLAAEE6AMAAAToAwAAEbWyZmDC+DuePSekzjycZ7l/UpFToIZ+b2pmJIB3fahQSwcI3+1QVyAAAAAUAAAAUEsBAh4DCgAJAAAA94nIXN/tUFcgAAAAFAAAAAgAGAAAAAAAAQAAAKSBAAAAAGRhdGEudHh0VVQFAAMxeiZqdXgLAAEE6AMAAAToAwAAUEsFBgAAAAABAAEATgAAAHIAAAAAAA=="

func fixture(t *testing.T) []byte {
	t.Helper()
	data, err := base64.StdEncoding.DecodeString(encryptedZipB64)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func run(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: errBuf}
	err := New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func writeFile(t *testing.T, name string, content []byte) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(p, content, 0o600); err != nil {
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

func TestVerifyCorrectPassword(t *testing.T) {
	t.Parallel()
	e, err := firstEncrypted(fixture(t))
	if err != nil {
		t.Fatal(err)
	}
	if !verifyPassword(e, "hunter2") {
		t.Error("hunter2 should verify against the fixture")
	}
}

func TestVerifyWrongPassword(t *testing.T) {
	t.Parallel()
	e, err := firstEncrypted(fixture(t))
	if err != nil {
		t.Fatal(err)
	}
	for _, wrong := range []string{"hunter1", "password", "", "Hunter2"} {
		if verifyPassword(e, wrong) {
			t.Errorf("%q should not verify", wrong)
		}
	}
}

func TestRunFindsPassword(t *testing.T) {
	t.Parallel()
	archive := writeFile(t, "test.zip", fixture(t))
	wl := writeFile(t, "words", []byte("alpha\nbeta\nhunter2\ngamma\n"))
	out, _, err := run(t, archive, "-w", wl)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if !strings.Contains(out, "password found: hunter2") {
		t.Errorf("out = %q", out)
	}
}

func TestRunPasswordNotInList(t *testing.T) {
	t.Parallel()
	archive := writeFile(t, "test.zip", fixture(t))
	wl := writeFile(t, "words", []byte("alpha\nbeta\n"))
	out, _, err := run(t, archive, "-w", wl)
	if code := exitCode(err); code != command.ExitFailure {
		t.Errorf("exit code = %d, want %d", code, command.ExitFailure)
	}
	if !strings.Contains(out, "not found") {
		t.Errorf("out = %q", out)
	}
}

func TestRunNoWordlist(t *testing.T) {
	t.Parallel()
	archive := writeFile(t, "test.zip", fixture(t))
	_, _, err := run(t, archive)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "wordlist is required") {
		t.Errorf("err = %v", err)
	}
}

func TestRunMissingArchive(t *testing.T) {
	t.Parallel()
	wl := writeFile(t, "words", []byte("x\n"))
	_, _, err := run(t, "/no/such/archive.zip", "-w", wl)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRunNotAZip(t *testing.T) {
	t.Parallel()
	bogus := writeFile(t, "bogus.zip", []byte("not a zip file"))
	wl := writeFile(t, "words", []byte("x\n"))
	_, _, err := run(t, bogus, "-w", wl)
	if err == nil {
		t.Fatal("expected error for a non-zip file")
	}
}

func TestNameSynopsis(t *testing.T) {
	t.Parallel()
	c := New()
	if c.Name() != "zip-pwcrack" {
		t.Errorf("Name() = %q", c.Name())
	}
	if c.Synopsis() == "" {
		t.Error("Synopsis() is empty")
	}
}
