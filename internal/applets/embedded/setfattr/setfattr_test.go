package setfattr

import (
	"bytes"
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

// fakeBackend records the mutations requested of it.
type fakeBackend struct {
	store   map[string]map[string][]byte
	setErr  error
	removed []string
}

func newFake() *fakeBackend {
	return &fakeBackend{store: map[string]map[string][]byte{}}
}

func (f *fakeBackend) Set(path, name string, value []byte, _ bool) error {
	if f.setErr != nil {
		return f.setErr
	}
	if f.store[path] == nil {
		f.store[path] = map[string][]byte{}
	}
	f.store[path][name] = value
	return nil
}

func (f *fakeBackend) Remove(path, name string, _ bool) error {
	f.removed = append(f.removed, path+":"+name)
	if f.store[path] != nil {
		delete(f.store[path], name)
	}
	return nil
}

func withBackend(t *testing.T, b Backend) {
	t.Helper()
	prev := xattrBackend
	xattrBackend = b
	t.Cleanup(func() { xattrBackend = prev })
}

func run(t *testing.T, args ...string) (string, error) {
	t.Helper()
	var out, errBuf bytes.Buffer
	stdio := command.IO{In: strings.NewReader(""), Out: &out, Err: &errBuf}
	err := New().Run(context.Background(), stdio, args)
	return errBuf.String(), err
}

func TestPlanSetText(t *testing.T) {
	act, err := plan("user.demo", "hello", "", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if act.remove || act.name != "user.demo" || string(act.value) != "hello" {
		t.Errorf("unexpected action: %+v", act)
	}
}

func TestPlanSetHex(t *testing.T) {
	act, err := plan("user.bin", "0xdeadbeef", "", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []byte{0xde, 0xad, 0xbe, 0xef}
	if !reflect.DeepEqual(act.value, want) {
		t.Errorf("hex decode wrong: %v", act.value)
	}
}

func TestPlanSetBase64(t *testing.T) {
	act, err := plan("user.bin", "0saGk=", "", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(act.value) != "hi" {
		t.Errorf("base64 decode wrong: %q", act.value)
	}
}

func TestPlanRemove(t *testing.T) {
	act, err := plan("", "", "user.demo", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !act.remove || act.name != "user.demo" {
		t.Errorf("unexpected action: %+v", act)
	}
}

func TestPlanErrors(t *testing.T) {
	tests := []struct {
		name                   string
		nameOpt, value, remove string
		valueGiven             bool
	}{
		{"both -n and -x", "user.a", "", "user.b", false},
		{"neither", "", "", "", false},
		{"missing value", "user.a", "", "", false},
		{"bad hex", "user.a", "0xZZ", "", true},
		{"bad base64", "user.a", "0s!!!", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := plan(tt.nameOpt, tt.value, tt.remove, tt.valueGiven); err == nil {
				t.Error("expected error")
			}
		})
	}
}

func TestPlanEmptyValueAllowed(t *testing.T) {
	act, err := plan("user.demo", "", "", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(act.value) != 0 {
		t.Errorf("expected empty value, got %v", act.value)
	}
}

func TestRunSet(t *testing.T) {
	fake := newFake()
	withBackend(t, fake)
	if _, err := run(t, "-n", "user.demo", "-v", "hello", "file.txt"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(fake.store["file.txt"]["user.demo"]) != "hello" {
		t.Errorf("value not stored: %+v", fake.store)
	}
}

func TestRunRemove(t *testing.T) {
	fake := newFake()
	withBackend(t, fake)
	if _, err := run(t, "-x", "user.demo", "file.txt"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(fake.removed) != 1 || fake.removed[0] != "file.txt:user.demo" {
		t.Errorf("remove not recorded: %v", fake.removed)
	}
}

func TestRunNoFiles(t *testing.T) {
	withBackend(t, newFake())
	if _, err := run(t, "-n", "user.demo", "-v", "x"); err == nil {
		t.Fatal("expected error with no files")
	}
}

func TestRunBackendError(t *testing.T) {
	fake := newFake()
	fake.setErr = errors.New("operation not supported")
	withBackend(t, fake)
	errOut, err := run(t, "-n", "user.demo", "-v", "x", "file.txt")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(errOut, "setfattr:") {
		t.Errorf("missing prefix: %q", errOut)
	}
}
