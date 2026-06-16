package stat

import (
	"errors"
	"io/fs"
	"os"
	"testing"
)

// TestFileType covers every mode-to-description branch of fileType, including
// the device sub-cases that the on-disk fixtures in the external test cannot
// portably exercise.
func TestFileType(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		mode fs.FileMode
		want string
	}{
		{"regular", 0o644, "regular file"},
		{"directory", fs.ModeDir | 0o755, "directory"},
		{"symlink", fs.ModeSymlink | 0o777, "symbolic link"},
		{"fifo", fs.ModeNamedPipe | 0o644, "fifo"},
		{"socket", fs.ModeSocket | 0o644, "socket"},
		{"char device", fs.ModeDevice | fs.ModeCharDevice | 0o644, "character special file"},
		{"block device", fs.ModeDevice | 0o644, "block special file"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := fileType(tt.mode); got != tt.want {
				t.Errorf("fileType(%v) = %q, want %q", tt.mode, got, tt.want)
			}
		})
	}
}

// TestUnwrap verifies unwrap peels a *os.PathError but passes plain errors
// through unchanged.
func TestUnwrap(t *testing.T) {
	t.Parallel()
	inner := errors.New("boom")
	pe := &os.PathError{Op: "stat", Path: "/x", Err: inner}
	if got := unwrap(pe); got != inner {
		t.Errorf("unwrap(PathError) = %v, want %v", got, inner)
	}
	plain := errors.New("plain")
	if got := unwrap(plain); got != plain {
		t.Errorf("unwrap(plain) = %v, want %v", got, plain)
	}
}

// TestUnescape covers the backslash escapes honoured inside a -c format.
func TestUnescape(t *testing.T) {
	t.Parallel()
	if got := unescape(`a\nb\tc\\d`); got != "a\nb\tc\\d" {
		t.Errorf("unescape = %q", got)
	}
}
