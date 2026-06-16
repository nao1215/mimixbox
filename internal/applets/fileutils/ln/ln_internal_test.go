package ln

import (
	"errors"
	"os"
	"testing"
)

func TestCapitalize(t *testing.T) {
	t.Parallel()
	tests := []struct {
		in, want string
	}{
		{"", ""},
		{"file exists", "File exists"},
		{"Already capital", "Already capital"},
		{"1 leading digit", "1 leading digit"},
	}
	for _, tt := range tests {
		if got := capitalize(tt.in); got != tt.want {
			t.Errorf("capitalize(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestReason(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		err  error
		want string
	}{
		{"exist", os.ErrExist, "File exists"},
		{"not exist", os.ErrNotExist, "No such file or directory"},
		{"permission", os.ErrPermission, "Permission denied"},
		{
			name: "link error",
			err:  &os.LinkError{Op: "symlink", Old: "a", New: "b", Err: errors.New("invalid argument")},
			want: "Invalid argument",
		},
		{
			name: "path error",
			err:  &os.PathError{Op: "remove", Path: "x", Err: errors.New("directory not empty")},
			want: "Directory not empty",
		},
		{"plain", errors.New("boom"), "Boom"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if got := reason(tt.err); got != tt.want {
				t.Errorf("reason(%v) = %q, want %q", tt.err, got, tt.want)
			}
		})
	}
}

func TestKeepFirstError(t *testing.T) {
	t.Parallel()
	// With no prior error, keep returns a failure sentinel.
	if keep(nil) == nil {
		t.Error("keep(nil) = nil, want a failure error")
	}
	// With a prior error, keep preserves it.
	prior := errors.New("first")
	if got := keep(prior); got != prior {
		t.Errorf("keep(prior) = %v, want %v", got, prior)
	}
}

func TestRemoveExisting(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	// A missing path is a no-op (not an error).
	if err := removeExisting(dir + "/missing"); err != nil {
		t.Errorf("removeExisting(missing) = %v, want nil", err)
	}

	// An existing file is removed.
	existing := dir + "/present"
	if err := os.WriteFile(existing, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := removeExisting(existing); err != nil {
		t.Fatalf("removeExisting(present) = %v, want nil", err)
	}
	if _, err := os.Lstat(existing); !os.IsNotExist(err) {
		t.Error("file was not removed")
	}
}
