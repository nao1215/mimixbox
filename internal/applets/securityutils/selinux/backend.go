package selinux

import (
	"os"
	"path/filepath"
	"strings"
)

// Mode is an SELinux enforcing mode.
type Mode int

const (
	// Disabled means SELinux is not active.
	Disabled Mode = iota
	// Permissive means policy is loaded but only logs denials.
	Permissive
	// Enforcing means policy is loaded and enforced.
	Enforcing
)

func (m Mode) String() string {
	switch m {
	case Enforcing:
		return "Enforcing"
	case Permissive:
		return "Permissive"
	default:
		return "Disabled"
	}
}

// Backend exposes the SELinux state the query commands need. Tests provide a
// fixture implementation; the default reads the kernel's selinuxfs mount.
type Backend interface {
	// Enabled reports whether SELinux is present (a policy is loaded).
	Enabled() bool
	// Enforce returns the current enforcing mode.
	Enforce() Mode
	// PolicyVersion returns the loaded policy version, or "" if unknown.
	PolicyVersion() string
	// Booleans returns the SELinux boolean name->state map.
	Booleans() map[string]bool
	// MatchPathCon returns the file context for path, or ("", false).
	MatchPathCon(path string) (string, bool)
}

// backend is the active backend; tests swap it out.
var backend Backend = &fsBackend{root: "/sys/fs/selinux"}

// fsBackend reads SELinux state from a selinuxfs mount root.
type fsBackend struct {
	root string
}

func (b *fsBackend) Enabled() bool {
	_, err := os.Stat(filepath.Join(b.root, "enforce"))
	return err == nil
}

func (b *fsBackend) Enforce() Mode {
	if !b.Enabled() {
		return Disabled
	}
	data, err := os.ReadFile(filepath.Join(b.root, "enforce")) //nolint:gosec // fixed selinuxfs path
	if err != nil {
		return Disabled
	}
	if strings.TrimSpace(string(data)) == "1" {
		return Enforcing
	}
	return Permissive
}

func (b *fsBackend) PolicyVersion() string {
	data, err := os.ReadFile(filepath.Join(b.root, "policyvers")) //nolint:gosec // fixed selinuxfs path
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func (b *fsBackend) Booleans() map[string]bool {
	out := map[string]bool{}
	entries, err := os.ReadDir(filepath.Join(b.root, "booleans"))
	if err != nil {
		return out
	}
	for _, e := range entries {
		data, err := os.ReadFile(filepath.Join(b.root, "booleans", e.Name())) //nolint:gosec // fixed selinuxfs path
		if err != nil {
			continue
		}
		// Format is "current pending"; treat 1 as on.
		out[e.Name()] = strings.HasPrefix(strings.TrimSpace(string(data)), "1")
	}
	return out
}

func (b *fsBackend) MatchPathCon(string) (string, bool) {
	// The selinuxfs mount does not expose file_contexts; default backend
	// cannot resolve contexts without libselinux. Tests use a fixture.
	return "", false
}

// fixtureBackend is a static, in-memory backend used by tests and by the
// SELINUX_FIXTURE environment hook.
type fixtureBackend struct {
	enabled  bool
	mode     Mode
	policyV  string
	booleans map[string]bool
	contexts map[string]string
}

func (f *fixtureBackend) Enabled() bool             { return f.enabled }
func (f *fixtureBackend) Enforce() Mode             { return f.mode }
func (f *fixtureBackend) PolicyVersion() string     { return f.policyV }
func (f *fixtureBackend) Booleans() map[string]bool { return f.booleans }
func (f *fixtureBackend) MatchPathCon(p string) (string, bool) {
	c, ok := f.contexts[p]
	return c, ok
}
