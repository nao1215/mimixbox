package seq_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/shellutils/seq"
	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: errBuf}
	err := seq.New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func TestRun(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"last only", []string{"3"}, "1\n2\n3\n"},
		{"first last", []string{"2", "5"}, "2\n3\n4\n5\n"},
		{"first increment last", []string{"1", "2", "9"}, "1\n3\n5\n7\n9\n"},
		{"separator", []string{"-s", ",", "3"}, "1,2,3\n"},
		{"equal width", []string{"-w", "8", "10"}, "08\n09\n10\n"},
		{"equal width carry", []string{"-w", "98", "102"}, "098\n099\n100\n101\n102\n"},
		{"descending", []string{"5", "-1", "1"}, "5\n4\n3\n2\n1\n"},
		{"float", []string{"1", "0.5", "2.5"}, "1.0\n1.5\n2.0\n2.5\n"},
		{"format", []string{"-f", "%.2f", "1", "3"}, "1.00\n2.00\n3.00\n"},
		{"empty range", []string{"5", "1"}, ""},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			out, _, err := run(t, tt.args...)
			if err != nil {
				t.Fatalf("Run error = %v", err)
			}
			if out != tt.want {
				t.Errorf("out = %q, want %q", out, tt.want)
			}
		})
	}
}

func TestRunInvalidArg(t *testing.T) {
	t.Parallel()
	out, errOut, err := run(t, "foo")
	if err == nil {
		t.Fatal("expected error for invalid argument")
	}
	if out != "" {
		t.Errorf("out = %q, want empty", out)
	}
	if !strings.Contains(errOut, "seq: invalid floating point argument: 'foo'") {
		t.Errorf("stderr = %q, want invalid argument message", errOut)
	}
}

func TestRunMissingOperand(t *testing.T) {
	t.Parallel()
	_, errOut, err := run(t)
	if err == nil {
		t.Fatal("expected error for missing operand")
	}
	if !strings.Contains(errOut, "seq: missing operand") {
		t.Errorf("stderr = %q, want missing operand message", errOut)
	}
}

func TestRunHelpAndVersion(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "--help")
	if err != nil {
		t.Fatalf("--help error = %v", err)
	}
	if !strings.Contains(out, "Usage: seq") {
		t.Errorf("--help out = %q", out)
	}

	out, _, err = run(t, "--version")
	if err != nil {
		t.Fatalf("--version error = %v", err)
	}
	if !strings.Contains(out, "seq (mimixbox)") {
		t.Errorf("--version out = %q", out)
	}
}
