package beep

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

type beepCall struct{ freq, length int }

func stub(t *testing.T, err error) *[]beepCall {
	t.Helper()
	var calls []beepCall
	orig := beepFn
	beepFn = func(freq, length int) error {
		calls = append(calls, beepCall{freq, length})
		return err
	}
	t.Cleanup(func() { beepFn = orig })
	return &calls
}

func run(t *testing.T, args ...string) error {
	t.Helper()
	io := command.IO{In: strings.NewReader(""), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	return New().Run(context.Background(), io, args)
}

func TestDefaultBeep(t *testing.T) {
	calls := stub(t, nil)
	if err := run(t); err != nil {
		t.Fatal(err)
	}
	if len(*calls) != 1 || (*calls)[0] != (beepCall{defaultFreq, defaultLength}) {
		t.Errorf("calls = %v", *calls)
	}
}

func TestCustomAndRepeats(t *testing.T) {
	calls := stub(t, nil)
	if err := run(t, "-f", "1000", "-l", "50", "-r", "3"); err != nil {
		t.Fatal(err)
	}
	if len(*calls) != 3 {
		t.Fatalf("got %d beeps, want 3", len(*calls))
	}
	for _, c := range *calls {
		if c != (beepCall{1000, 50}) {
			t.Errorf("beep = %v, want {1000 50}", c)
		}
	}
}

func TestErrors(t *testing.T) {
	stub(t, nil)
	if err := run(t, "-f", "0"); err == nil {
		t.Errorf("a zero frequency should fail")
	}
	if err := run(t, "-r", "0"); err == nil {
		t.Errorf("a zero repeat count should fail")
	}
	stub(t, errors.New("no console"))
	if err := run(t); err == nil {
		t.Errorf("a speaker failure should fail")
	}
}
