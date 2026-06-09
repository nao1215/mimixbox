package users

import (
	"bytes"
	"context"
	"encoding/binary"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

// record builds one 384-byte utmp record of the given type and user.
func record(utype int16, user string) []byte {
	b := make([]byte, recordSize)
	binary.LittleEndian.PutUint16(b[typeOffset:], uint16(utype))
	copy(b[userOffset:userOffset+userLen], user)
	return b
}

func TestParse(t *testing.T) {
	t.Parallel()
	var data []byte
	data = append(data, record(userProcess, "alice")...)
	data = append(data, record(8, "dead")...) // DEAD_PROCESS is ignored
	data = append(data, record(userProcess, "bob")...)
	got := parse(data)
	if len(got) != 2 || got[0] != "alice" || got[1] != "bob" {
		t.Errorf("parse = %v, want [alice bob]", got)
	}
}

func run(t *testing.T, args ...string) string {
	t.Helper()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	if err := New().Run(context.Background(), io, args); err != nil {
		t.Fatalf("Run error = %v", err)
	}
	return out.String()
}

func TestRunSortsAndJoins(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	f := filepath.Join(dir, "utmp")
	var data []byte
	for _, u := range []string{"carol", "alice", "bob"} {
		data = append(data, record(userProcess, u)...)
	}
	if err := os.WriteFile(f, data, 0o644); err != nil {
		t.Fatal(err)
	}
	if got := run(t, f); got != "alice bob carol\n" {
		t.Errorf("users = %q, want %q", got, "alice bob carol\n")
	}
}

func TestRunMissingFile(t *testing.T) {
	t.Parallel()
	// A missing utmp means nobody is logged in: a single empty line.
	if got := run(t, "/no/such/users/utmp"); got != "\n" {
		t.Errorf("missing utmp = %q, want a blank line", got)
	}
}

func TestRunDefaultPath(t *testing.T) {
	// Not parallel: it mutates the package-level utmpPath.
	dir := t.TempDir()
	f := filepath.Join(dir, "utmp")
	if err := os.WriteFile(f, record(userProcess, "dave"), 0o644); err != nil {
		t.Fatal(err)
	}
	orig := utmpPath
	utmpPath = f
	defer func() { utmpPath = orig }()
	if got := run(t); got != "dave\n" {
		t.Errorf("default path = %q, want dave", got)
	}
}
