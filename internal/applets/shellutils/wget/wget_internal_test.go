package wget

import (
	"errors"
	"net/url"
	"path/filepath"
	"testing"
)

func TestSynopsis(t *testing.T) {
	t.Parallel()
	c := New()
	if c.Name() != "wget" {
		t.Errorf("Name() = %q, want wget", c.Name())
	}
	if c.Synopsis() == "" {
		t.Error("Synopsis() is empty")
	}
}

// TestDestination exercises the pure URL-to-filename derivation: -O overrides
// everything, a path with no usable base name falls back to index.html, and -P
// joins the derived name under a directory prefix.
func TestDestination(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		opts   options
		rawURL string
		want   string
	}{
		{"basename", options{}, "https://example.com/a/b.tar.gz", "b.tar.gz"},
		{"output overrides", options{output: "fixed.bin"}, "https://example.com/a/b.tar.gz", "fixed.bin"},
		{"output stdout", options{output: "-"}, "https://example.com/file", "-"},
		{"empty path falls back", options{}, "https://example.com", "index.html"},
		{"root path falls back", options{}, "https://example.com/", "index.html"},
		{"prefix joins", options{prefix: "downloads"}, "https://example.com/x.txt", filepath.Join("downloads", "x.txt")},
		{"prefix with fallback", options{prefix: "out"}, "https://example.com/", filepath.Join("out", "index.html")},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			parsed, err := url.Parse(tt.rawURL)
			if err != nil {
				t.Fatalf("url.Parse(%q) = %v", tt.rawURL, err)
			}
			if got := destination(tt.opts, parsed); got != tt.want {
				t.Errorf("destination(%+v, %q) = %q, want %q", tt.opts, tt.rawURL, got, tt.want)
			}
		})
	}
}

// TestRetryableWrapping verifies the retryableErr marker: it reports as
// retryable, and Unwrap exposes the underlying error for errors.Is/As.
func TestRetryableWrapping(t *testing.T) {
	t.Parallel()
	base := errors.New("boom")
	wrapped := retryable(base)

	if !isRetryable(wrapped) {
		t.Error("retryable(err) should be reported as retryable")
	}
	if wrapped.Error() != "boom" {
		t.Errorf("Error() = %q, want %q", wrapped.Error(), "boom")
	}
	if !errors.Is(wrapped, base) {
		t.Error("Unwrap should expose the wrapped error to errors.Is")
	}
	if isRetryable(base) {
		t.Error("a plain error must not be reported as retryable")
	}
}
