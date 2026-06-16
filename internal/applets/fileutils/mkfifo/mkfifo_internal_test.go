package mkfifo

import (
	"errors"
	"os"
	"strings"
	"syscall"
	"testing"
)

// TestParseMode checks octal mode parsing accepts valid octal strings and
// rejects malformed ones.
func TestParseMode(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		in      string
		want    os.FileMode
		wantErr bool
	}{
		{"three digit", "644", 0o644, false},
		{"leading zero", "0755", 0o755, false},
		{"owner only", "600", 0o600, false},
		{"non-octal digit", "8", 0, true},
		{"empty", "", 0, true},
		{"garbage", "rwx", 0, true},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := parseMode(tt.in)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseMode(%q) err = %v, wantErr %v", tt.in, err, tt.wantErr)
			}
			if err == nil && got != tt.want {
				t.Errorf("parseMode(%q) = %o, want %o", tt.in, got, tt.want)
			}
		})
	}
}

// TestReason extracts the lower-case message from a syscall errno and falls
// back to the error string for a non-errno error.
func TestReason(t *testing.T) {
	t.Parallel()
	if got := reason(syscall.ENOENT); got != "no such file or directory" {
		t.Errorf("reason(ENOENT) = %q, want %q", got, "no such file or directory")
	}
	plain := errors.New("boom")
	if got := reason(plain); got != "boom" {
		t.Errorf("reason(plain) = %q, want %q", got, "boom")
	}
}

// TestFifoError formats the already-exist sentinel and a generic syscall error
// the way the integration spec expects.
func TestFifoError(t *testing.T) {
	t.Parallel()
	if got := fifoError("/tmp/p", errExist); got != "can't make /tmp/p: already exist" {
		t.Errorf("fifoError(exist) = %q", got)
	}
	got := fifoError("/no/such/p", syscall.ENOENT)
	if !strings.HasPrefix(got, "/no/such/p: ") || !strings.Contains(got, "no such file or directory") {
		t.Errorf("fifoError(ENOENT) = %q, want path + lower-case reason", got)
	}
}

// TestMakeFifoExistingPath confirms makeFifo reports the already-exist sentinel
// when the path is taken.
func TestMakeFifoExistingPath(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	existing := dir + "/taken"
	if err := os.WriteFile(existing, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := makeFifo(existing, 0o644); !errors.Is(err, errExist) {
		t.Errorf("makeFifo on existing path err = %v, want errExist", err)
	}
}
