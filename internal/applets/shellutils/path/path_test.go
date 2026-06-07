package path_test

import (
	"bytes"
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/shellutils/path"
	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: errBuf}
	err := path.New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func TestRun(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"basename", []string{"-b", "/etc/systemd/pstore.conf"}, "pstore.conf\n"},
		{"canonical default (no option)", []string{"cmd/../scripts/installer.sh"}, "scripts/installer.sh\n"},
		{"canonical explicit", []string{"-c", "cmd/../scripts/installer.sh"}, "scripts/installer.sh\n"},
		{"dirname", []string{"-d", "/etc/ssh/ssh_config"}, "/etc/ssh\n"},
		{"extension", []string{"-e", "go.mod"}, ".mod\n"},
		{"long basename", []string{"--basename", "/a/b/c.txt"}, "c.txt\n"},
		{"long dirname", []string{"--dirname", "/a/b/c.txt"}, "/a/b\n"},
		{"long extension", []string{"--extension", "/a/b/c.txt"}, ".txt\n"},
		{"combined base+ext", []string{"-b", "-e", "/a/b/c.txt"}, "c.txt\n.txt\n"},
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

func TestRunAbsolute(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "-a", "foo")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	want, aerr := filepath.Abs("foo")
	if aerr != nil {
		t.Fatalf("filepath.Abs: %v", aerr)
	}
	if out != want+"\n" {
		t.Errorf("out = %q, want %q", out, want+"\n")
	}
}

func TestRunMissingOperand(t *testing.T) {
	t.Parallel()
	_, errOut, err := run(t)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(errOut, "missing operand") {
		t.Errorf("stderr = %q", errOut)
	}
}
