package reformime

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

const singlePart = "MIME-Version: 1.0\r\n" +
	"Content-Type: text/plain\r\n" +
	"Content-Transfer-Encoding: base64\r\n" +
	"\r\n" +
	"aGVsbG8gd29ybGQ=\r\n"

const multiPart = "MIME-Version: 1.0\r\n" +
	"Content-Type: multipart/mixed; boundary=BOUND\r\n" +
	"\r\n" +
	"--BOUND\r\n" +
	"Content-Type: text/plain\r\n" +
	"\r\n" +
	"first part\r\n" +
	"--BOUND\r\n" +
	"Content-Type: application/octet-stream\r\n" +
	"Content-Transfer-Encoding: base64\r\n" +
	"\r\n" +
	"c2Vjb25k\r\n" +
	"--BOUND--\r\n"

func run(t *testing.T, input string, args ...string) (string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(input), Out: out, Err: &bytes.Buffer{}}
	err := New().Run(context.Background(), io, args)
	return out.String(), err
}

// headerThenErr yields the message header on its first read and then fails,
// modeling a body that becomes unreadable after the header is parsed.
type headerThenErr struct {
	header []byte
	done   bool
}

func (r *headerThenErr) Read(p []byte) (int, error) {
	if !r.done {
		r.done = true
		return copy(p, r.header), nil
	}
	return 0, errors.New("body read failed")
}

func TestParseUnparseableContentTypeSurfacesBodyError(t *testing.T) {
	// "text/plain; x" is a non-empty but unparseable Content-Type, so parse()
	// takes the opaque-fallback path. A failing body read must be reported, not
	// swallowed into an empty part.
	hdr := "Content-Type: text/plain; x\r\n\r\n"
	if _, err := parse(&headerThenErr{header: []byte(hdr)}); err == nil {
		t.Fatal("parse must surface a body read error on the invalid Content-Type path")
	}
}

func TestListSingle(t *testing.T) {
	out, err := run(t, singlePart)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "1\ttext/plain") {
		t.Errorf("expected single part listing, got %q", out)
	}
}

func TestExtractSingle(t *testing.T) {
	out, err := run(t, singlePart, "-x", "1")
	if err != nil {
		t.Fatal(err)
	}
	if out != "hello world" {
		t.Errorf("decoded part = %q, want %q", out, "hello world")
	}
}

func TestListMultipart(t *testing.T) {
	out, err := run(t, multiPart)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "1\ttext/plain") || !strings.Contains(out, "2\tapplication/octet-stream") {
		t.Errorf("multipart listing wrong:\n%s", out)
	}
}

func TestExtractMultipart(t *testing.T) {
	out, err := run(t, multiPart, "-x", "2")
	if err != nil {
		t.Fatal(err)
	}
	if out != "second" {
		t.Errorf("decoded part 2 = %q, want %q", out, "second")
	}
}

func TestExtractOutOfRange(t *testing.T) {
	if _, err := run(t, multiPart, "-x", "9"); err == nil {
		t.Fatal("expected error for out-of-range part")
	}
}
