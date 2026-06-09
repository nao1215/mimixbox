package hexdump

import (
	"bytes"
	"context"
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
}
