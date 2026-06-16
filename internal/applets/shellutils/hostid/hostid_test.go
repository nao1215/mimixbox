package hostid_test

import (
	"bytes"
	"context"
	"regexp"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/shellutils/hostid"
	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: errBuf}
	err := hostid.New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func TestRun(t *testing.T) {
	t.Parallel()

	out, errOut, err := run(t)
	if err != nil {
		t.Fatalf("Run error = %v, stderr = %q", err, errOut)
	}

	// Like coreutils, hostid prints exactly one line of 8 lowercase hex digits.
	if !regexp.MustCompile(`^[0-9a-f]{8}\n$`).MatchString(out) {
		t.Errorf("hostid output = %q, want one line of 8 hex digits", out)
	}
}

func TestHelpSections(t *testing.T) {
	out, _, err := run(t, "--help")
	if err != nil {
		t.Fatalf("--help error = %v", err)
	}
	if !strings.Contains(out, "Examples:") || !strings.Contains(out, "Exit status:") {
		t.Errorf("--help missing structured sections:\n%s", out)
	}
}
