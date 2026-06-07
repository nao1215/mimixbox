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

	hex8 := regexp.MustCompile(`^[0-9a-f]{8}$`)
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		if line == "" {
			// Host has no non-loopback IPv4 address; nothing to print.
			continue
		}
		if !hex8.MatchString(line) {
			t.Errorf("hostid output line = %q, want 8 hex digits", line)
		}
	}
}
