package probe

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, cmd *Command, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: errBuf}
	err := cmd.Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func TestCapabilityErrorByDefault(t *testing.T) {
	// The default transport must report a capability error, never succeed silently.
	for _, c := range []*Command{NewTraceroute(), NewTraceroute6(), NewPing6(), NewArping()} {
		_, _, err := run(t, c, hostFor(c))
		if err == nil {
			t.Fatalf("%s: expected capability error, got nil", c.Name())
		}
		if !strings.Contains(err.Error(), "raw socket") {
			t.Errorf("%s: err = %v, want capability message", c.Name(), err)
		}
	}
}

// hostFor returns a literal address of the family each applet requires.
func hostFor(c *Command) string {
	switch specs[c.kind].want {
	case ipv6Only:
		return "2001:db8::1"
	default:
		return "192.0.2.1"
	}
}

func TestTransportSuccessPath(t *testing.T) {
	orig := transport
	transport = func(_ context.Context, name string, tgt Target) (string, error) {
		return name + " to " + tgt.Host + "\n", nil
	}
	t.Cleanup(func() { transport = orig })

	out, _, err := run(t, NewTraceroute(), "192.0.2.1")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if !strings.Contains(out, "traceroute to 192.0.2.1") {
		t.Errorf("transport output not printed: %s", out)
	}
}

func TestFamilyValidation(t *testing.T) {
	// IPv6 literal given to an IPv4-only applet must be rejected at parse time.
	if _, _, err := run(t, NewTraceroute(), "2001:db8::1"); err == nil {
		t.Error("traceroute should reject an IPv6 literal")
	}
	// IPv4 literal given to ping6 must be rejected.
	if _, _, err := run(t, NewPing6(), "192.0.2.1"); err == nil {
		t.Error("ping6 should reject an IPv4 literal")
	}
}

func TestNameRequired(t *testing.T) {
	if _, _, err := run(t, NewArping()); err == nil {
		t.Error("expected error with no host")
	}
	if _, _, err := run(t, NewArping(), "a", "b"); err == nil {
		t.Error("expected error with two operands")
	}
}

func TestNamesAndSynopses(t *testing.T) {
	t.Parallel()
	cmds := map[string]*Command{
		"traceroute":  NewTraceroute(),
		"traceroute6": NewTraceroute6(),
		"ping6":       NewPing6(),
		"arping":      NewArping(),
	}
	for want, c := range cmds {
		if c.Name() != want {
			t.Errorf("Name() = %q, want %q", c.Name(), want)
		}
		if c.Synopsis() == "" {
			t.Errorf("%s Synopsis() empty", want)
		}
	}
}
