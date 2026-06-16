package command

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func newGate(plan func([]string) (string, error)) PlanGate {
	return PlanGate{
		Plan: plan,
		Gate: func(p string) error { return Failuref("planned action [%s] is gated", p) },
	}
}

func TestPlanGateReportsPlanThenGates(t *testing.T) {
	out := &bytes.Buffer{}
	stdio := IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	fs := NewFlagSet("demo", "ARG...", stdio.Err)
	g := newGate(func(args []string) (string, error) {
		return "demo do " + strings.Join(args, " "), nil
	})

	err := g.Run(fs, stdio, []string{"a", "b"})
	if err == nil || !strings.Contains(err.Error(), "is gated") {
		t.Fatalf("expected gate error, got %v", err)
	}
	if got := out.String(); got != "demo: planned action: demo do a b\n" {
		t.Errorf("stdout = %q", got)
	}
}

func TestPlanGateValidationErrorPrintsNoPlan(t *testing.T) {
	out := &bytes.Buffer{}
	stdio := IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	fs := NewFlagSet("demo", "ARG...", stdio.Err)
	g := newGate(func([]string) (string, error) { return "", errors.New("bad args") })

	err := g.Run(fs, stdio, []string{"x"})
	if err == nil || !strings.Contains(err.Error(), "bad args") {
		t.Fatalf("expected validation error, got %v", err)
	}
	if out.Len() != 0 {
		t.Errorf("plan must not be printed on validation error, got %q", out.String())
	}
}

func TestPlanGateHelpStops(t *testing.T) {
	out := &bytes.Buffer{}
	stdio := IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	fs := NewFlagSet("demo", "ARG...", stdio.Err).WithHelp(Help{Description: "demo"})
	called := false
	g := newGate(func([]string) (string, error) { called = true; return "x", nil })

	if err := g.Run(fs, stdio, []string{"--help"}); err != nil {
		t.Fatalf("--help err = %v", err)
	}
	if called {
		t.Error("Plan must not run when --help short-circuits")
	}
	if !strings.Contains(out.String(), "Usage: demo") {
		t.Errorf("--help out = %q", out.String())
	}
}
