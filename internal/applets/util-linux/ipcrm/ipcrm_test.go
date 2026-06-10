package ipcrm

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

type removal struct {
	kind string
	id   int
}

func withStub(t *testing.T, failKind string) *[]removal {
	t.Helper()
	var calls []removal
	orig := removeIPC
	removeIPC = func(kind string, id int) error {
		calls = append(calls, removal{kind, id})
		if kind == failKind {
			return errors.New("boom")
		}
		return nil
	}
	t.Cleanup(func() { removeIPC = orig })
	return &calls
}

func run(t *testing.T, args ...string) error {
	t.Helper()
	io := command.IO{In: strings.NewReader(""), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	return New().Run(context.Background(), io, args)
}

func TestRemovesEachKind(t *testing.T) {
	calls := withStub(t, "")
	if err := run(t, "-q", "1", "-m", "2", "-s", "3"); err != nil {
		t.Fatal(err)
	}
	want := []removal{{"msg", 1}, {"shm", 2}, {"sem", 3}}
	if len(*calls) != 3 {
		t.Fatalf("calls = %v", *calls)
	}
	for i, w := range want {
		if (*calls)[i] != w {
			t.Errorf("call %d = %v, want %v", i, (*calls)[i], w)
		}
	}
}

func TestMultipleIds(t *testing.T) {
	calls := withStub(t, "")
	if err := run(t, "-q", "10", "-q", "20"); err != nil {
		t.Fatal(err)
	}
	if len(*calls) != 2 || (*calls)[0].id != 10 || (*calls)[1].id != 20 {
		t.Errorf("calls = %v", *calls)
	}
}

func TestNothingToRemove(t *testing.T) {
	withStub(t, "")
	if err := run(t); err == nil {
		t.Errorf("no ids should fail")
	}
}

func TestRemovalFailure(t *testing.T) {
	withStub(t, "shm")
	if err := run(t, "-m", "5"); err == nil {
		t.Errorf("a failed removal should fail")
	}
}
