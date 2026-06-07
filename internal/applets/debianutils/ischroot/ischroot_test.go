package ischroot_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/debianutils/ischroot"
	"github.com/nao1215/mimixbox/internal/command"
)

func TestNew(t *testing.T) {
	t.Parallel()
	if ischroot.New() == nil {
		t.Fatal("New() returned nil")
	}
}

func TestNameSynopsis(t *testing.T) {
	t.Parallel()
	c := ischroot.New()
	if got := c.Name(); got != "ischroot" {
		t.Errorf("Name() = %q, want %q", got, "ischroot")
	}
	if got := c.Synopsis(); got != "Detect if running in a chroot" {
		t.Errorf("Synopsis() = %q, want %q", got, "Detect if running in a chroot")
	}
}

// TestRunExitCode asserts that running ischroot returns one of the valid Debian
// exit codes (0 in a chroot, 1 if not, 2 if undetectable) and writes nothing to
// stdout. The test environment may or may not be a chroot, so the whole valid
// set is accepted.
func TestRunExitCode(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		args []string
	}{
		{"no option", nil},
		{"default-false", []string{"-f"}},
		{"default-true", []string{"-t"}},
		{"long default-false", []string{"--default-false"}},
		{"long default-true", []string{"--default-true"}},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			out := &bytes.Buffer{}
			errBuf := &bytes.Buffer{}
			io := command.IO{In: strings.NewReader(""), Out: out, Err: errBuf}

			code := command.Execute(context.Background(), ischroot.New(), io, tt.args)

			switch code {
			case 0, 1, 2:
				// valid
			default:
				t.Errorf("exit code = %d, want one of 0, 1, 2", code)
			}
			if out.Len() != 0 {
				t.Errorf("stdout = %q, want empty", out.String())
			}
		})
	}
}
