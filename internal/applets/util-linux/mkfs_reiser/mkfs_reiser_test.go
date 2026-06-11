package mkfsreiser

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, args ...string) (string, error) {
	t.Helper()
	errb := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: &bytes.Buffer{}, Err: errb}
	err := New().Run(context.Background(), io, args)
	return errb.String(), err
}

func TestRefusesWithExplanation(t *testing.T) {
	out, err := run(t, "/tmp/disk.img")
	if err == nil {
		t.Fatalf("mkfs.reiser should fail deterministically")
	}
	if !strings.Contains(out, "deprecated") || !strings.Contains(out, "mke2fs") {
		t.Errorf("explanation missing: %q", out)
	}
}

func TestRequiresDevice(t *testing.T) {
	if _, err := run(t); err == nil {
		t.Errorf("missing device should fail")
	}
}
