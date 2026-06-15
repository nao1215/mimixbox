package getfattr

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

// fakeBackend is an in-memory xattr store keyed by path then name.
type fakeBackend struct {
	data    map[string]map[string][]byte
	listErr error
}

func (f *fakeBackend) List(path string, _ bool) ([]string, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	attrs, ok := f.data[path]
	if !ok {
		return nil, errors.New("no such file or directory")
	}
	var names []string
	for n := range attrs {
		names = append(names, n)
	}
	return names, nil
}

func (f *fakeBackend) Get(path, name string, _ bool) ([]byte, error) {
	attrs, ok := f.data[path]
	if !ok {
		return nil, errors.New("no such file or directory")
	}
	v, ok := attrs[name]
	if !ok {
		return nil, errors.New("no such attribute")
	}
	return v, nil
}

func withBackend(t *testing.T, b Backend) {
	t.Helper()
	prev := xattrBackend
	xattrBackend = b
	t.Cleanup(func() { xattrBackend = prev })
}

func run(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	var out, errBuf bytes.Buffer
	stdio := command.IO{In: strings.NewReader(""), Out: &out, Err: &errBuf}
	err := New().Run(context.Background(), stdio, args)
	return out.String(), errBuf.String(), err
}

func TestGetfattrListNames(t *testing.T) {
	withBackend(t, &fakeBackend{data: map[string]map[string][]byte{
		"file.txt": {
			"user.demo":     []byte("hello"),
			"user.author":   []byte("nao"),
			"security.selinux": []byte("system_u"),
		},
	}})
	out, _, err := run(t, "file.txt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Only user.* shown by default, sorted, no values.
	want := "# file: file.txt\nuser.author\nuser.demo\n\n"
	if out != want {
		t.Errorf("output mismatch:\n got %q\nwant %q", out, want)
	}
}

func TestGetfattrDumpText(t *testing.T) {
	withBackend(t, &fakeBackend{data: map[string]map[string][]byte{
		"file.txt": {"user.demo": []byte("hello")},
	}})
	out, _, err := run(t, "-d", "file.txt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "# file: file.txt\nuser.demo=\"hello\"\n\n"
	if out != want {
		t.Errorf("got %q want %q", out, want)
	}
}

func TestGetfattrDumpHex(t *testing.T) {
	withBackend(t, &fakeBackend{data: map[string]map[string][]byte{
		"file.txt": {"user.demo": []byte{0x00, 0xff, 0x41}},
	}})
	out, _, err := run(t, "-d", "-e", "hex", "file.txt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "user.demo=0x00ff41") {
		t.Errorf("hex encoding missing: %q", out)
	}
}

func TestGetfattrDumpBase64(t *testing.T) {
	withBackend(t, &fakeBackend{data: map[string]map[string][]byte{
		"file.txt": {"user.demo": []byte("hi")},
	}})
	out, _, err := run(t, "-d", "-e", "base64", "file.txt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "user.demo=0saGk=") {
		t.Errorf("base64 encoding missing: %q", out)
	}
}

func TestGetfattrSingleName(t *testing.T) {
	withBackend(t, &fakeBackend{data: map[string]map[string][]byte{
		"file.txt": {"user.demo": []byte("hello"), "security.x": []byte("y")},
	}})
	out, _, err := run(t, "-n", "security.x", "file.txt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// -n names an attribute explicitly, so the namespace filter is bypassed
	// and the value is dumped.
	if !strings.Contains(out, "security.x=\"y\"") {
		t.Errorf("named attribute not dumped: %q", out)
	}
}

func TestGetfattrMatch(t *testing.T) {
	withBackend(t, &fakeBackend{data: map[string]map[string][]byte{
		"file.txt": {"user.demo": []byte("a"), "security.x": []byte("b")},
	}})
	out, _, err := run(t, "-m", "security", "file.txt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "security.x") || strings.Contains(out, "user.demo") {
		t.Errorf("match filter wrong: %q", out)
	}
}

func TestGetfattrNoFiles(t *testing.T) {
	withBackend(t, &fakeBackend{})
	_, _, err := run(t)
	if err == nil {
		t.Fatal("expected error when no file operand given")
	}
}

func TestGetfattrBadEncoding(t *testing.T) {
	withBackend(t, &fakeBackend{})
	_, _, err := run(t, "-e", "rot13", "file.txt")
	if err == nil {
		t.Fatal("expected error for unknown encoding")
	}
}

func TestGetfattrReadError(t *testing.T) {
	withBackend(t, &fakeBackend{listErr: errors.New("not supported")})
	_, errOut, err := run(t, "file.txt")
	if err == nil {
		t.Fatal("expected error when backend fails")
	}
	if !strings.Contains(errOut, "getfattr:") {
		t.Errorf("missing error prefix: %q", errOut)
	}
}
