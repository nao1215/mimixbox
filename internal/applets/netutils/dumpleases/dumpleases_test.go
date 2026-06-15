package dumpleases

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/nao1215/mimixbox/internal/command"
)

func TestParseLeases(t *testing.T) {
	t.Parallel()
	in := `# comment line
00:11:22:33:44:55 192.168.1.10 alpha 1000000000

aa:bb:cc:dd:ee:ff 192.168.1.11 * 0
`
	leases, err := ParseLeases(strings.NewReader(in))
	if err != nil {
		t.Fatalf("ParseLeases error: %v", err)
	}
	if len(leases) != 2 {
		t.Fatalf("got %d leases, want 2", len(leases))
	}
	if leases[0].Hostname != "alpha" || leases[0].IP.String() != "192.168.1.10" {
		t.Errorf("lease 0 = %+v", leases[0])
	}
	if !leases[1].Expires.IsZero() {
		t.Errorf("lease 1 should be static (zero expiry), got %v", leases[1].Expires)
	}
}

func TestParseLeasesErrors(t *testing.T) {
	t.Parallel()
	tests := []string{
		"toofew fields here",
		"zz:zz:zz:zz:zz:zz 1.2.3.4 host 1",
		"00:11:22:33:44:55 not-an-ip host 1",
		"00:11:22:33:44:55 1.2.3.4 host notanumber",
	}
	for _, in := range tests {
		if _, err := ParseLeases(strings.NewReader(in)); err == nil {
			t.Errorf("expected error for %q, got nil", in)
		}
	}
}

func TestFormatExpiry(t *testing.T) {
	t.Parallel()
	now := time.Unix(1000, 0).UTC()
	static := Lease{}
	if got := formatExpiry(static, false, now); got != "never" {
		t.Errorf("static lease expiry = %q, want never", got)
	}
	future := Lease{Expires: time.Unix(1000+3661, 0).UTC()}
	if got := formatExpiry(future, false, now); got != "01:01:01" {
		t.Errorf("remaining = %q, want 01:01:01", got)
	}
	past := Lease{Expires: time.Unix(500, 0).UTC()}
	if got := formatExpiry(past, false, now); got != "expired" {
		t.Errorf("past = %q, want expired", got)
	}
	if got := formatExpiry(future, true, now); got != "1970-01-01T01:17:41Z" {
		t.Errorf("absolute = %q", got)
	}
}

func TestRunReadsFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "leases.db")
	if err := os.WriteFile(path, []byte("00:11:22:33:44:55 10.0.0.5 host1 0\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	out := &bytes.Buffer{}
	stdio := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	if err := New().Run(context.Background(), stdio, []string{path}); err != nil {
		t.Fatalf("Run error: %v", err)
	}
	if !strings.Contains(out.String(), "10.0.0.5") || !strings.Contains(out.String(), "host1") {
		t.Errorf("output missing lease data:\n%s", out.String())
	}
}

func TestRunMissingFile(t *testing.T) {
	t.Parallel()
	stdio := command.IO{In: strings.NewReader(""), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	if err := New().Run(context.Background(), stdio, []string{"/no/such/lease/file"}); err == nil {
		t.Fatal("expected error for missing file")
	}
}
