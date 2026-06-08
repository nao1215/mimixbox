package command_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func TestHandleHelpVersion(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		args        []string
		wantHandled bool
		wantOut     string // substring expected in stdout (empty = no output)
	}{
		{"help first", []string{"--help"}, true, "Usage: demo [OPERAND]..."},
		{"version first", []string{"--version"}, true, "demo (mimixbox)"},
		{"help not first", []string{"x", "--help"}, false, ""},
		{"version not first", []string{"x", "--version"}, false, ""},
		{"no args", nil, false, ""},
		{"unrelated", []string{"foo"}, false, ""},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			out := &bytes.Buffer{}
			stdio := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
			handled := command.HandleHelpVersion(stdio, "demo", "[OPERAND]...", tt.args)
			if handled != tt.wantHandled {
				t.Errorf("handled = %v, want %v", handled, tt.wantHandled)
			}
			if tt.wantOut == "" {
				if out.Len() != 0 {
					t.Errorf("expected no output, got %q", out.String())
				}
				return
			}
			if !strings.Contains(out.String(), tt.wantOut) {
				t.Errorf("out = %q, want substring %q", out.String(), tt.wantOut)
			}
		})
	}
}
