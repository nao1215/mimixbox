package selinux

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func withBackend(t *testing.T, b Backend) {
	t.Helper()
	old := backend
	backend = b
	t.Cleanup(func() { backend = old })
}

func enabledBackend() *fixtureBackend {
	return &fixtureBackend{
		enabled: true,
		mode:    Enforcing,
		policyV: "33",
		booleans: map[string]bool{
			"httpd_can_network_connect": true,
			"ssh_sysadm_login":          false,
		},
		contexts: map[string]string{
			"/etc/passwd": "system_u:object_r:passwd_file_t:s0",
		},
	}
}

func runCmd(t *testing.T, c *Command, args ...string) (string, string, error) {
	t.Helper()
	out, errBuf := &bytes.Buffer{}, &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: errBuf}
	err := c.Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func TestGetenforce(t *testing.T) {
	withBackend(t, enabledBackend())
	out, _, err := runCmd(t, NewGetenforce())
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(out) != "Enforcing" {
		t.Errorf("getenforce = %q, want Enforcing", out)
	}
}

func TestGetenforceDisabled(t *testing.T) {
	withBackend(t, &fixtureBackend{enabled: false})
	out, _, err := runCmd(t, NewGetenforce())
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(out) != "Disabled" {
		t.Errorf("getenforce = %q, want Disabled", out)
	}
}

func TestSelinuxenabled(t *testing.T) {
	withBackend(t, enabledBackend())
	if _, _, err := runCmd(t, NewSelinuxenabled()); err != nil {
		t.Errorf("expected exit 0 when enabled, got %v", err)
	}

	withBackend(t, &fixtureBackend{enabled: false})
	_, _, err := runCmd(t, NewSelinuxenabled())
	ee, ok := err.(*command.ExitError)
	if !ok || ee.Code != command.ExitFailure {
		t.Errorf("expected ExitFailure when disabled, got %v", err)
	}
}

func TestSestatus(t *testing.T) {
	withBackend(t, enabledBackend())
	out, _, err := runCmd(t, NewSestatus())
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"SELinux status:", "enabled", "Current mode:", "enforcing", "33"} {
		if !strings.Contains(out, want) {
			t.Errorf("sestatus output missing %q:\n%s", want, out)
		}
	}
}

func TestGetsebool(t *testing.T) {
	withBackend(t, enabledBackend())

	out, _, err := runCmd(t, NewGetsebool(), "-a")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "httpd_can_network_connect --> on") {
		t.Errorf("getsebool -a missing on boolean:\n%s", out)
	}
	if !strings.Contains(out, "ssh_sysadm_login --> off") {
		t.Errorf("getsebool -a missing off boolean:\n%s", out)
	}

	out, _, err = runCmd(t, NewGetsebool(), "ssh_sysadm_login")
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(out) != "ssh_sysadm_login --> off" {
		t.Errorf("getsebool single = %q", out)
	}

	_, _, err = runCmd(t, NewGetsebool(), "no_such_boolean")
	if err == nil {
		t.Error("expected error for unknown boolean")
	}
}

func TestMatchpathcon(t *testing.T) {
	withBackend(t, enabledBackend())
	out, _, err := runCmd(t, NewMatchpathcon(), "/etc/passwd")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "passwd_file_t") {
		t.Errorf("matchpathcon = %q", out)
	}

	if _, _, err := runCmd(t, NewMatchpathcon(), "/no/context"); err == nil {
		t.Error("expected error for path with no context")
	}
}

func TestPrivilegedGated(t *testing.T) {
	for _, c := range []*Command{
		NewSetenforce(), NewSetsebool(), NewChcon(), NewRuncon(),
		NewRestorecon(), NewSetfiles(), NewLoadPolicy(),
	} {
		_, _, err := runCmd(t, c, "arg")
		if err == nil {
			t.Errorf("%s: expected deterministic failure", c.Name())
			continue
		}
		if !strings.Contains(err.Error(), "intentionally not implemented") {
			t.Errorf("%s: error %q lacks documented reason", c.Name(), err)
		}
	}
}

func TestPrivilegedHelpStillWorks(t *testing.T) {
	out, _, err := runCmd(t, NewSetenforce(), "--help")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Usage: setenforce") || !strings.Contains(out, "CAP_MAC_ADMIN") {
		t.Errorf("setenforce --help unexpected:\n%s", out)
	}
}
