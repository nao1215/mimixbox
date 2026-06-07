package uuidgen_test

import (
	"bytes"
	"context"
	"regexp"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/shellutils/uuidgen"
	"github.com/nao1215/mimixbox/internal/command"
)

var uuidRe = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

func run(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: errBuf}
	err := uuidgen.New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func TestRun(t *testing.T) {
	t.Parallel()
	out, errOut, err := run(t)
	if err != nil {
		t.Fatalf("Run error = %v (stderr=%q)", err, errOut)
	}
	got := strings.TrimSpace(out)
	if !uuidRe.MatchString(got) {
		t.Errorf("uuid = %q, does not match %s", got, uuidRe.String())
	}
}

func TestRunUnique(t *testing.T) {
	t.Parallel()
	first, _, err := run(t)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	second, _, err := run(t)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if strings.TrimSpace(first) == strings.TrimSpace(second) {
		t.Errorf("two runs produced the same UUID: %q", strings.TrimSpace(first))
	}
}
