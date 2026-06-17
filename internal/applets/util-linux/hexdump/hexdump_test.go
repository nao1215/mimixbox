package hexdump

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, c *Command, in string, args ...string) (string, string) {
	t.Helper()
	out, errBuf := &bytes.Buffer{}, &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(in), Out: out, Err: errBuf}
	if err := c.Run(context.Background(), io, args); err != nil {
		t.Fatalf("Run error = %v (stderr=%q)", err, errBuf.String())
	}
	return out.String(), errBuf.String()
}

// dripReader returns its data one byte per Read call so rows are assembled
// across read boundaries.
type dripReader struct {
	data []byte
	pos  int
}

func (d *dripReader) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	if d.pos >= len(d.data) {
		return 0, io.EOF
	}
	p[0] = d.data[d.pos]
	d.pos++
	return 1, nil
}

// failWriter fails every write, modeling a closed downstream pipe.
type failWriter struct{}

func (failWriter) Write([]byte) (int, error) { return 0, io.ErrClosedPipe }

func TestWriteErrorIsReported(t *testing.T) {
	t.Parallel()
	// A failing output stream must surface an error rather than silently
	// draining the input and exiting 0.
	io1 := command.IO{In: bytes.NewReader(bytes.Repeat([]byte("x"), 200000)), Out: failWriter{}, Err: &bytes.Buffer{}}
	if err := NewHexdump().Run(context.Background(), io1, nil); err == nil {
		t.Error("hexdump must report a write error to stdout")
	}
}

func TestStreamingMatchesSingleRead(t *testing.T) {
	t.Parallel()
	// Streaming a large input one byte at a time must match a single read, for
	// both the canonical and default layouts (issue #952).
	data := bytes.Repeat([]byte{0x00, 0x41, 0xff, 0x7f, 0x10, '\n'}, 5000) // ~30 KiB
	for _, canonical := range []bool{false, true} {
		var args []string
		if canonical {
			args = []string{"-C"}
		}
		whole := &bytes.Buffer{}
		io1 := command.IO{In: bytes.NewReader(data), Out: whole, Err: &bytes.Buffer{}}
		if err := NewHexdump().Run(context.Background(), io1, args); err != nil {
			t.Fatalf("whole run error = %v", err)
		}
		drip := &bytes.Buffer{}
		io2 := command.IO{In: &dripReader{data: data}, Out: drip, Err: &bytes.Buffer{}}
		if err := NewHexdump().Run(context.Background(), io2, args); err != nil {
			t.Fatalf("drip run error = %v", err)
		}
		if whole.String() != drip.String() {
			t.Errorf("canonical=%v: streaming output differs from single read", canonical)
		}
	}
}

// These golden strings were verified byte-for-byte against util-linux hd /
// hexdump.
func TestHdCanonical(t *testing.T) {
	t.Parallel()
	want := "00000000  68 65 6c 6c 6f 20 77 6f  72 6c 64 0a              |hello world.|\n0000000c\n"
	if got, _ := run(t, NewHd(), "hello world\n"); got != want {
		t.Errorf("hd =\n%q\nwant\n%q", got, want)
	}
	// hexdump -C is the same as hd.
	if got, _ := run(t, NewHexdump(), "hello world\n", "-C"); got != want {
		t.Errorf("hexdump -C =\n%q\nwant\n%q", got, want)
	}
}

func TestHexdumpTwoByte(t *testing.T) {
	t.Parallel()
	want := "0000000 6568 6c6c 206f 6f77 6c72 0a64          \n000000c\n"
	if got, _ := run(t, NewHexdump(), "hello world\n"); got != want {
		t.Errorf("hexdump =\n%q\nwant\n%q", got, want)
	}
}

func TestLengthAndSkip(t *testing.T) {
	t.Parallel()
	// -n 4 keeps the first four bytes; the trailing offset is 00000004.
	got, _ := run(t, NewHd(), "abcdefgh", "-n", "4")
	if !strings.HasPrefix(got, "00000000  61 62 63 64 ") || !strings.Contains(got, "|abcd|") {
		t.Errorf("hd -n 4 = %q", got)
	}
	if !strings.HasSuffix(got, "00000004\n") {
		t.Errorf("hd -n 4 final offset = %q", got)
	}
	// -s 4 skips the first four bytes.
	got, _ = run(t, NewHd(), "abcdefgh", "-s", "4")
	if !strings.Contains(got, "|efgh|") || strings.Contains(got, "|abcd|") {
		t.Errorf("hd -s 4 = %q", got)
	}
}

func TestEmptyInput(t *testing.T) {
	t.Parallel()
	if got, _ := run(t, NewHd(), ""); got != "00000000\n" {
		t.Errorf("hd of empty = %q, want %q", got, "00000000\n")
	}
}

func TestHelp(t *testing.T) {
	t.Parallel()
	out, _ := run(t, NewHexdump(), "", "--help")
	if !strings.Contains(out, "Usage: hexdump") {
		t.Errorf("--help = %q", out)
	}
	if !strings.Contains(out, "Exit status:") {
		t.Errorf("--help missing exit status section = %q", out)
	}
	// hd shares the parameterized command; its help also documents exit status.
	hdOut, _ := run(t, NewHd(), "", "--help")
	if !strings.Contains(hdOut, "Exit status:") {
		t.Errorf("hd --help missing exit status section = %q", hdOut)
	}
}
