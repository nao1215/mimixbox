package command_test

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

// stub is a Command whose Run is supplied by the test.
type stub struct {
	name string
	run  func(ctx context.Context, io command.IO, args []string) error
}

func (s stub) Name() string     { return s.name }
func (s stub) Synopsis() string { return "stub command" }
func (s stub) Run(ctx context.Context, io command.IO, args []string) error {
	return s.run(ctx, io, args)
}

func newIO() (command.IO, *bytes.Buffer, *bytes.Buffer) {
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	return command.IO{In: strings.NewReader(""), Out: out, Err: errBuf}, out, errBuf
}

func TestExecute(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		run      func(ctx context.Context, io command.IO, args []string) error
		wantCode int
		wantErr  string // substring expected on stderr
	}{
		{
			name:     "success returns 0 and prints nothing",
			run:      func(context.Context, command.IO, []string) error { return nil },
			wantCode: command.ExitSuccess,
			wantErr:  "",
		},
		{
			name:     "plain error maps to failure with name prefix",
			run:      func(context.Context, command.IO, []string) error { return errors.New("boom") },
			wantCode: command.ExitFailure,
			wantErr:  "demo: boom",
		},
		{
			name: "ExitError carries its own code",
			run: func(context.Context, command.IO, []string) error {
				return &command.ExitError{Code: 3, Err: errors.New("nope")}
			},
			wantCode: 3,
			wantErr:  "demo: nope",
		},
		{
			name:     "silent failure prints nothing extra",
			run:      func(context.Context, command.IO, []string) error { return command.SilentFailure() },
			wantCode: command.ExitFailure,
			wantErr:  "",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			io, _, errBuf := newIO()
			got := command.Execute(context.Background(), stub{name: "demo", run: tt.run}, io, nil)
			if got != tt.wantCode {
				t.Errorf("exit code = %d, want %d", got, tt.wantCode)
			}
			if tt.wantErr == "" {
				if errBuf.Len() != 0 {
					t.Errorf("stderr = %q, want empty", errBuf.String())
				}
			} else if !strings.Contains(errBuf.String(), tt.wantErr) {
				t.Errorf("stderr = %q, want to contain %q", errBuf.String(), tt.wantErr)
			}
		})
	}
}

func TestFlagSetHelp(t *testing.T) {
	t.Parallel()
	io, out, _ := newIO()
	fs := command.NewFlagSet("demo", "[OPTION]... [FILE]...", io.Err)
	fs.Bool("number", false, "number all output lines")

	proceed, err := fs.Parse(io, []string{"--help"})
	if err != nil {
		t.Fatalf("Parse(--help) error = %v", err)
	}
	if proceed {
		t.Fatal("Parse(--help) proceed = true, want false")
	}
	if !strings.Contains(out.String(), "Usage: demo [OPTION]... [FILE]...") {
		t.Errorf("help output missing usage line: %q", out.String())
	}
	if !strings.Contains(out.String(), "--number") {
		t.Errorf("help output missing --number: %q", out.String())
	}
}

func TestFlagSetVersion(t *testing.T) {
	t.Parallel()
	io, out, _ := newIO()
	fs := command.NewFlagSet("demo", "", io.Err)

	proceed, err := fs.Parse(io, []string{"--version"})
	if err != nil || proceed {
		t.Fatalf("Parse(--version) = (%v, %v), want (false, nil)", proceed, err)
	}
	if !strings.Contains(out.String(), "demo (mimixbox)") {
		t.Errorf("version output = %q", out.String())
	}
}

func TestFlagSetUnknownFlag(t *testing.T) {
	t.Parallel()
	io, _, errBuf := newIO()
	fs := command.NewFlagSet("demo", "", io.Err)

	proceed, err := fs.Parse(io, []string{"--nope"})
	if err == nil || proceed {
		t.Fatalf("Parse(--nope) = (%v, %v), want (false, error)", proceed, err)
	}
	if !strings.Contains(errBuf.String(), "unknown flag") {
		t.Errorf("stderr = %q, want unknown flag message", errBuf.String())
	}
	if !strings.Contains(errBuf.String(), "Try 'demo --help'") {
		t.Errorf("stderr = %q, want hint", errBuf.String())
	}
}

func TestFlagSetGNUClustering(t *testing.T) {
	t.Parallel()
	io, _, _ := newIO()
	fs := command.NewFlagSet("demo", "", io.Err)
	a := fs.BoolP("all", "a", false, "")
	b := fs.BoolP("bytes", "b", false, "")

	proceed, err := fs.Parse(io, []string{"-ab", "operand"})
	if err != nil || !proceed {
		t.Fatalf("Parse(-ab) = (%v, %v), want (true, nil)", proceed, err)
	}
	if !*a || !*b {
		t.Errorf("clustered flags not set: a=%v b=%v", *a, *b)
	}
	if got := fs.Args(); len(got) != 1 || got[0] != "operand" {
		t.Errorf("operands = %v, want [operand]", got)
	}
}
