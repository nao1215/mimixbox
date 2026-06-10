package blockdev

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func withStub(t *testing.T, values map[string]uint64, fail bool) *[]string {
	t.Helper()
	var asked []string
	orig := blockQuery
	blockQuery = func(_, name string) (uint64, error) {
		asked = append(asked, name)
		if fail {
			return 0, errors.New("permission denied")
		}
		return values[name], nil
	}
	t.Cleanup(func() { blockQuery = orig })
	return &asked
}

func run(t *testing.T, args ...string) (string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	err := New().Run(context.Background(), io, args)
	return strings.TrimSpace(out.String()), err
}

func TestGetSize64(t *testing.T) {
	withStub(t, map[string]uint64{"getsize64": 1073741824}, false)
	out, err := run(t, "--getsize64", "/dev/sda")
	if err != nil {
		t.Fatal(err)
	}
	if out != "1073741824" {
		t.Errorf("--getsize64 = %q", out)
	}
}

func TestSectorSize(t *testing.T) {
	asked := withStub(t, map[string]uint64{"getss": 512}, false)
	out, err := run(t, "--getss", "/dev/sda")
	if err != nil {
		t.Fatal(err)
	}
	if out != "512" {
		t.Errorf("--getss = %q", out)
	}
	if len(*asked) != 1 || (*asked)[0] != "getss" {
		t.Errorf("asked = %v", *asked)
	}
}

func TestMultipleQueries(t *testing.T) {
	withStub(t, map[string]uint64{"getss": 512, "getro": 1}, false)
	out, err := run(t, "--getss", "--getro", "/dev/sda")
	if err != nil {
		t.Fatal(err)
	}
	if out != "512\n1" {
		t.Errorf("multiple = %q, want \"512\\n1\"", out)
	}
}

func TestGetszAliasesGetsize64(t *testing.T) {
	asked := withStub(t, map[string]uint64{"getsize64": 999}, false)
	if _, err := run(t, "--getsz", "/dev/sda"); err != nil {
		t.Fatal(err)
	}
	if len(*asked) != 1 || (*asked)[0] != "getsize64" {
		t.Errorf("--getsz should query getsize64, asked = %v", *asked)
	}
}

func TestErrors(t *testing.T) {
	withStub(t, nil, false)
	if _, err := run(t, "/dev/sda"); err == nil {
		t.Errorf("no query flag should fail")
	}
	if _, err := run(t, "--getro"); err == nil {
		t.Errorf("no device should fail")
	}
	withStub(t, nil, true)
	if _, err := run(t, "--getss", "/dev/sda"); err == nil {
		t.Errorf("a query failure should fail")
	}
}
