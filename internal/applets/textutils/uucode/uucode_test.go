package uucode

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func encode(t *testing.T, data []byte, args ...string) string {
	t.Helper()
	out := &bytes.Buffer{}
	io := command.IO{In: bytes.NewReader(data), Out: out, Err: &bytes.Buffer{}}
	if err := NewUuencode().Run(context.Background(), io, args); err != nil {
		t.Fatalf("uuencode error = %v", err)
	}
	return out.String()
}

func decodeToStdout(t *testing.T, encoded string) []byte {
	t.Helper()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(encoded), Out: out, Err: &bytes.Buffer{}}
	if err := NewUudecode().Run(context.Background(), io, []string{"-o", "-"}); err != nil {
		t.Fatalf("uudecode error = %v", err)
	}
	return out.Bytes()
}

func TestRoundTrip(t *testing.T) {
	t.Parallel()
	payloads := [][]byte{
		[]byte("hello world\n"),
		{0x00, 0x01, 0x02, 0xff, 0xfe, '\n', 'x'},
		bytes.Repeat([]byte("The quick brown fox. "), 20),
		{},
	}
	for _, mode := range []string{"uu", "base64"} {
		for i, p := range payloads {
			var enc string
			if mode == "base64" {
				enc = encode(t, p, "-m", "name")
			} else {
				enc = encode(t, p, "name")
			}
			got := decodeToStdout(t, enc)
			if !bytes.Equal(got, p) {
				t.Errorf("%s payload %d round-trip mismatch:\n got %v\nwant %v", mode, i, got, p)
			}
		}
	}
}

func TestStreamsLargeInputRoundTrip(t *testing.T) {
	t.Parallel()
	// A large payload exercises the streaming encoders (45-byte UU lines and the
	// streaming base64 wrapper) and must round-trip unchanged (issue #961).
	payload := bytes.Repeat([]byte{0x00, 0x01, 0xfe, 0xff, 'a', '\n'}, 100000) // ~600 KiB
	for _, mode := range []string{"uu", "base64"} {
		var enc string
		if mode == "base64" {
			enc = encode(t, payload, "-m", "name")
		} else {
			enc = encode(t, payload, "name")
		}
		if got := decodeToStdout(t, enc); !bytes.Equal(got, payload) {
			t.Errorf("%s large round-trip mismatch: got %d bytes, want %d", mode, len(got), len(payload))
		}
	}
}

func TestHeaderName(t *testing.T) {
	t.Parallel()
	enc := encode(t, []byte("x"), "myfile.bin")
	if !strings.HasPrefix(enc, "begin 644 myfile.bin\n") {
		t.Errorf("header = %q", enc[:20])
	}
	if !strings.HasSuffix(enc, "end\n") {
		t.Errorf("missing end trailer: %q", enc)
	}
}

func TestDecodeToNamedFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	target := filepath.Join(dir, "decoded.bin")
	enc := encode(t, []byte("payload\n"), target)
	io := command.IO{In: strings.NewReader(enc), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	if err := NewUudecode().Run(context.Background(), io, nil); err != nil {
		t.Fatalf("uudecode error = %v", err)
	}
	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "payload\n" {
		t.Errorf("decoded file = %q", got)
	}
}

func TestDecodeNoBegin(t *testing.T) {
	t.Parallel()
	io := command.IO{In: strings.NewReader("not encoded data\n"), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	if err := NewUudecode().Run(context.Background(), io, []string{"-o", "-"}); err == nil {
		t.Errorf("input without a begin line should fail")
	}
}

func TestHelpExitStatus(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name string
		cmd  *Command
	}{
		{"uuencode", NewUuencode()},
		{"uudecode", NewUudecode()},
	} {
		out := &bytes.Buffer{}
		io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
		if err := tc.cmd.Run(context.Background(), io, []string{"--help"}); err != nil {
			t.Fatalf("%s --help err = %v", tc.name, err)
		}
		if !strings.Contains(out.String(), "Exit status:") {
			t.Errorf("%s --help missing Exit status section = %q", tc.name, out.String())
		}
	}
}
