package cksum_test

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/textutils/cksum"
	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, stdin string, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(stdin), Out: out, Err: errBuf}
	err := cksum.New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func TestRunStdin(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		stdin string
		want  string
	}{
		{"empty", "", "4294967295 0\n"},
		{"hello", "hello\n", "3015617425 6\n"},
		{"digits", "123456789", "930766865 9\n"},
		{"one byte", "a", "1220704766 1\n"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			out, _, err := run(t, tt.stdin)
			if err != nil {
				t.Fatalf("Run error = %v", err)
			}
			if out != tt.want {
				t.Errorf("out = %q, want %q", out, tt.want)
			}
		})
	}
}

// dripReader returns its data one byte per Read call, forcing the consumer to
// accumulate across many read boundaries.
type dripReader struct {
	data []byte
	pos  int
}

func (d *dripReader) Read(p []byte) (int, error) {
	if d.pos >= len(d.data) {
		return 0, io.EOF
	}
	p[0] = d.data[d.pos]
	d.pos++
	return 1, nil
}

func TestStreamingChecksumMatchesSingleRead(t *testing.T) {
	t.Parallel()
	// The streaming CRC must produce the same checksum whether the input arrives
	// as one chunk or one byte at a time (issue #952).
	data := strings.Repeat("mimixbox cksum streaming check\n", 5000)

	whole := &bytes.Buffer{}
	io1 := command.IO{In: strings.NewReader(data), Out: whole, Err: &bytes.Buffer{}}
	if err := cksum.New().Run(context.Background(), io1, nil); err != nil {
		t.Fatalf("whole-read error = %v", err)
	}

	drip := &bytes.Buffer{}
	io2 := command.IO{In: &dripReader{data: []byte(data)}, Out: drip, Err: &bytes.Buffer{}}
	if err := cksum.New().Run(context.Background(), io2, nil); err != nil {
		t.Fatalf("drip-read error = %v", err)
	}

	if whole.String() != drip.String() {
		t.Errorf("checksum differs by read chunking: whole=%q drip=%q", whole.String(), drip.String())
	}
}

func TestRunFileNameInOutput(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "", "-")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	// "-" is treated as a file operand, so the name is echoed.
	if !strings.Contains(out, " -\n") {
		t.Errorf("out = %q, want name suffix", out)
	}
}

func TestRunMissingFile(t *testing.T) {
	t.Parallel()
	_, errOut, err := run(t, "", "/no/such/file")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(errOut, "cksum: /no/such/file:") {
		t.Errorf("stderr = %q", errOut)
	}
}

func TestNameSynopsis(t *testing.T) {
	t.Parallel()
	c := cksum.New()
	if c.Name() != "cksum" {
		t.Errorf("Name() = %q", c.Name())
	}
	if c.Synopsis() == "" {
		t.Error("Synopsis() is empty")
	}
}

// TestHelpSections verifies that --help renders both the Examples and the
// Exit status sections supplied through WithHelp.
func TestHelpSections(t *testing.T) {
	out, _, err := run(t, "", "--help")
	if err != nil {
		t.Fatalf("--help err = %v", err)
	}
	for _, want := range []string{"Examples:", "Exit status:"} {
		if !strings.Contains(out, want) {
			t.Errorf("--help missing %q section:\n%s", want, out)
		}
	}
}
