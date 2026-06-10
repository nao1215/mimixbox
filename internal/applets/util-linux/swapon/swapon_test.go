package swapon

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func setup(t *testing.T, swaponErr error) *[]string {
	t.Helper()
	dir := t.TempDir()
	f := filepath.Join(dir, "swaps")
	content := "Filename\tType\tSize\tUsed\tPriority\n/dev/sda2 partition 2097148 0 -2\n"
	if err := os.WriteFile(f, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	var called []string
	op, of := swapsPath, swaponFn
	swapsPath = f
	swaponFn = func(path string) error {
		called = append(called, path)
		return swaponErr
	}
	t.Cleanup(func() { swapsPath, swaponFn = op, of })
	return &called
}

func run(t *testing.T, args ...string) (string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	err := New().Run(context.Background(), io, args)
	return out.String(), err
}

func TestSummary(t *testing.T) {
	setup(t, nil)
	out, err := run(t, "-s")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "/dev/sda2") || !strings.Contains(out, "Filename") {
		t.Errorf("summary wrong:\n%s", out)
	}
}

func TestEnable(t *testing.T) {
	called := setup(t, nil)
	if _, err := run(t, "/swapfile"); err != nil {
		t.Fatal(err)
	}
	if len(*called) != 1 || (*called)[0] != "/swapfile" {
		t.Errorf("swaponFn called with %v", *called)
	}
}

func TestEnableFails(t *testing.T) {
	setup(t, errors.New("operation not permitted"))
	if _, err := run(t, "/swapfile"); err == nil {
		t.Errorf("an enable failure should fail")
	}
}
