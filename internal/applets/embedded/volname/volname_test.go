package volname

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

// makeISO writes a minimal file with an ISO 9660 volume identifier at the
// correct offset, padded with spaces as the format requires.
func makeISO(t *testing.T, label string) string {
	t.Helper()
	size := pvdOffset + volIDFieldOff + volIDFieldBytes
	buf := make([]byte, size)
	field := make([]byte, volIDFieldBytes)
	for i := range field {
		field[i] = ' '
	}
	copy(field, label)
	copy(buf[pvdOffset+volIDFieldOff:], field)
	p := filepath.Join(t.TempDir(), "disc.iso")
	if err := os.WriteFile(p, buf, 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func run(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	var out, errBuf bytes.Buffer
	stdio := command.IO{In: strings.NewReader(""), Out: &out, Err: &errBuf}
	err := New().Run(context.Background(), stdio, args)
	return out.String(), errBuf.String(), err
}

func TestVolnameReadsLabel(t *testing.T) {
	iso := makeISO(t, "MY_VOLUME")
	out, _, err := run(t, iso)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.TrimSpace(out) != "MY_VOLUME" {
		t.Errorf("unexpected label: %q", out)
	}
}

func TestVolnameMissingFile(t *testing.T) {
	_, errOut, err := run(t, filepath.Join(t.TempDir(), "absent.iso"))
	if err == nil {
		t.Fatal("expected error for missing file")
	}
	if !strings.Contains(errOut, "volname:") {
		t.Errorf("missing prefix: %q", errOut)
	}
}

func TestVolnameTooManyArgs(t *testing.T) {
	if _, _, err := run(t, "a", "b"); err == nil {
		t.Fatal("expected error for too many operands")
	}
}
