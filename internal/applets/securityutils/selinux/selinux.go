// Package selinux implements the read-only SELinux query applets (getenforce,
// selinuxenabled, sestatus, getsebool, matchpathcon) plus deterministic,
// documented stubs for the privileged mutating applets (setenforce, setsebool,
// chcon, runcon, restorecon, setfiles, load_policy).
//
// All state is read through an injectable backend so the commands can be tested
// hermetically without touching the host's /sys/fs/selinux mount.
package selinux

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
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

// command name constants for the multi-applet package.
const (
	cmdGetenforce     = "getenforce"
	cmdSelinuxenabled = "selinuxenabled"
	cmdSestatus       = "sestatus"
	cmdGetsebool      = "getsebool"
	cmdMatchpathcon   = "matchpathcon"
	cmdSetenforce     = "setenforce"
	cmdSetsebool      = "setsebool"
	cmdChcon          = "chcon"
	cmdRuncon         = "runcon"
	cmdRestorecon     = "restorecon"
	cmdSetfiles       = "setfiles"
	cmdLoadPolicy     = "load_policy"
)

// Command is one SELinux applet, distinguished by name.
type Command struct {
	name string
}

// Name returns the command name.
func (c *Command) Name() string { return c.name }

// Constructors for each applet.

// NewGetenforce returns the getenforce applet.
func NewGetenforce() *Command { return &Command{name: cmdGetenforce} }

// NewSelinuxenabled returns the selinuxenabled applet.
func NewSelinuxenabled() *Command { return &Command{name: cmdSelinuxenabled} }

// NewSestatus returns the sestatus applet.
func NewSestatus() *Command { return &Command{name: cmdSestatus} }

// NewGetsebool returns the getsebool applet.
func NewGetsebool() *Command { return &Command{name: cmdGetsebool} }

// NewMatchpathcon returns the matchpathcon applet.
func NewMatchpathcon() *Command { return &Command{name: cmdMatchpathcon} }

// NewSetenforce returns the setenforce applet.
func NewSetenforce() *Command { return &Command{name: cmdSetenforce} }

// NewSetsebool returns the setsebool applet.
func NewSetsebool() *Command { return &Command{name: cmdSetsebool} }

// NewChcon returns the chcon applet.
func NewChcon() *Command { return &Command{name: cmdChcon} }

// NewRuncon returns the runcon applet.
func NewRuncon() *Command { return &Command{name: cmdRuncon} }

// NewRestorecon returns the restorecon applet.
func NewRestorecon() *Command { return &Command{name: cmdRestorecon} }

// NewSetfiles returns the setfiles applet.
func NewSetfiles() *Command { return &Command{name: cmdSetfiles} }

// NewLoadPolicy returns the load_policy applet.
func NewLoadPolicy() *Command { return &Command{name: cmdLoadPolicy} }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string {
	switch c.name {
	case cmdGetenforce:
		return "Print the current SELinux enforcing mode"
	case cmdSelinuxenabled:
		return "Exit 0 if SELinux is enabled, 1 otherwise"
	case cmdSestatus:
		return "Show the SELinux status summary"
	case cmdGetsebool:
		return "Show the state of SELinux booleans"
	case cmdMatchpathcon:
		return "Show the default file context for a path"
	case cmdSetenforce:
		return "Set the SELinux enforcing mode (privileged)"
	case cmdSetsebool:
		return "Set the state of an SELinux boolean (privileged)"
	case cmdChcon:
		return "Change the SELinux security context of files (privileged)"
	case cmdRuncon:
		return "Run a program in a different SELinux context (privileged)"
	case cmdRestorecon:
		return "Restore default SELinux contexts on files (privileged)"
	case cmdSetfiles:
		return "Set file SELinux contexts from a spec file (privileged)"
	case cmdLoadPolicy:
		return "Load a new SELinux policy into the kernel (privileged)"
	}
	return "SELinux utility"
}

// Run dispatches to the per-command implementation.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	switch c.name {
	case cmdGetenforce:
		return c.runGetenforce(stdio, args)
	case cmdSelinuxenabled:
		return c.runSelinuxenabled(stdio, args)
	case cmdSestatus:
		return c.runSestatus(stdio, args)
	case cmdGetsebool:
		return c.runGetsebool(stdio, args)
	case cmdMatchpathcon:
		return c.runMatchpathcon(stdio, args)
	default:
		return c.runPrivileged(stdio, args)
	}
}

func (c *Command) runGetenforce(stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.name, "", stdio.Err).WithHelp(command.Help{
		Description: "Print the current SELinux enforcing mode: Enforcing, Permissive, or Disabled. " +
			"This is a read-only query against the kernel selinuxfs mount.",
		Examples:   []command.Example{{Command: "getenforce", Explain: "Print the current mode."}},
		ExitStatus: "0  always (when arguments parse).",
	})
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}
	if !backend.Enabled() {
		_, _ = fmt.Fprintln(stdio.Out, Disabled)
		return nil
	}
	_, _ = fmt.Fprintln(stdio.Out, backend.Enforce())
	return nil
}

func (c *Command) runSelinuxenabled(stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.name, "", stdio.Err).WithHelp(command.Help{
		Description: "Exit with status 0 if SELinux is enabled (a policy is loaded), or status 1 if it " +
			"is not. Prints nothing. Useful in shell scripts to gate SELinux-specific logic.",
		Examples:   []command.Example{{Command: "selinuxenabled && echo on", Explain: "Run only when SELinux is enabled."}},
		ExitStatus: "0  SELinux is enabled.\n1  SELinux is disabled.",
	})
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}
	if backend.Enabled() {
		return nil
	}
	return &command.ExitError{Code: command.ExitFailure}
}

func (c *Command) runSestatus(stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.name, "", stdio.Err).WithHelp(command.Help{
		Description: "Print a summary of the SELinux subsystem: whether the selinuxfs mount is present, " +
			"the current and configured mode, and the loaded policy version. Read-only.",
		Examples:   []command.Example{{Command: "sestatus", Explain: "Show the SELinux status summary."}},
		ExitStatus: "0  always (when arguments parse).",
	})
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}
	status := "disabled"
	if backend.Enabled() {
		status = "enabled"
	}
	out := bufio.NewWriter(stdio.Out)
	fmt.Fprintf(out, "SELinux status:                 %s\n", status)
	if backend.Enabled() {
		fmt.Fprintf(out, "SELinuxfs mount:                /sys/fs/selinux\n")
		fmt.Fprintf(out, "Current mode:                   %s\n", strings.ToLower(backend.Enforce().String()))
		if v := backend.PolicyVersion(); v != "" {
			fmt.Fprintf(out, "Max kernel policy version:      %s\n", v)
		}
	}
	return out.Flush()
}

func (c *Command) runGetsebool(stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.name, "[-a] [BOOLEAN]...", stdio.Err).WithHelp(command.Help{
		Description: "Show the state (on/off) of SELinux booleans. With -a, or with no names, every " +
			"boolean is listed; otherwise only the named booleans are shown. Read-only query.",
		Examples: []command.Example{
			{Command: "getsebool -a", Explain: "List every boolean and its state."},
			{Command: "getsebool httpd_can_network_connect", Explain: "Show one boolean's state."},
		},
		ExitStatus: "0  success.\n1  a requested boolean does not exist.",
	})
	all := fs.BoolP("all", "a", false, "show all booleans")
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}
	bools := backend.Booleans()
	names := fs.Args()
	if *all || len(names) == 0 {
		keys := make([]string, 0, len(bools))
		for k := range bools {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			fmt.Fprintf(stdio.Out, "%s --> %s\n", k, onOff(bools[k]))
		}
		return nil
	}
	var failed bool
	for _, n := range names {
		v, ok := bools[n]
		if !ok {
			fmt.Fprintf(stdio.Err, "%s: %s: no such boolean\n", c.name, n)
			failed = true
			continue
		}
		fmt.Fprintf(stdio.Out, "%s --> %s\n", n, onOff(v))
	}
	if failed {
		return command.SilentFailure()
	}
	return nil
}

func onOff(b bool) string {
	if b {
		return "on"
	}
	return "off"
}

func (c *Command) runMatchpathcon(stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.name, "PATH...", stdio.Err).WithHelp(command.Help{
		Description: "Print the default SELinux file context that the loaded policy assigns to each PATH, " +
			"in the form 'PATH context'. Read-only: it consults policy file-context rules and does " +
			"not change any file's label.",
		Examples:   []command.Example{{Command: "matchpathcon /etc/passwd", Explain: "Show the default context for /etc/passwd."}},
		ExitStatus: "0  success.\n1  a path has no matching context, or none were given.",
	})
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}
	paths := fs.Args()
	if len(paths) == 0 {
		fmt.Fprintf(stdio.Err, "%s: no path given\n", c.name)
		return command.SilentFailure()
	}
	var failed bool
	for _, p := range paths {
		ctx, ok := backend.MatchPathCon(p)
		if !ok {
			fmt.Fprintf(stdio.Err, "%s: %s: no default context\n", c.name, p)
			failed = true
			continue
		}
		fmt.Fprintf(stdio.Out, "%s\t%s\n", p, ctx)
	}
	if failed {
		return command.SilentFailure()
	}
	return nil
}

// runPrivileged handles the mutating SELinux applets. They validate arguments
// and report the capability/policy requirements deterministically rather than
// silently mutating the host. --help and --version still work.
func (c *Command) runPrivileged(stdio command.IO, args []string) error {
	usage, desc := privilegedHelp(c.name)
	fs := command.NewFlagSet(c.name, usage, stdio.Err).WithHelp(command.Help{
		Description: desc,
		Examples: []command.Example{
			{Command: c.name + " " + exampleArgs(c.name), Explain: "Validate the request, then report the capability/policy requirement."},
		},
		ExitStatus: "1  always in this build: the privileged operation is intentionally gated.",
		Notes: []string{
			"Mutating SELinux operations require CAP_MAC_ADMIN and a loaded policy; this build refuses them deterministically instead of partially applying changes.",
		},
	})
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}
	return command.Failuref(
		"%s: refusing to mutate SELinux state: requires CAP_MAC_ADMIN and a loaded policy; "+
			"this operation is intentionally not implemented in the hermetic build", c.name)
}

// privilegedHelp returns the usage operand summary and description paragraph for
// each privileged applet.
func privilegedHelp(name string) (usage, desc string) {
	switch name {
	case cmdSetenforce:
		return "[Enforcing|Permissive|1|0]", "Set the SELinux enforcing mode. Requires CAP_MAC_ADMIN; intentionally gated in this build."
	case cmdSetsebool:
		return "BOOLEAN VALUE...", "Set the state of one or more SELinux booleans. Requires CAP_MAC_ADMIN; intentionally gated in this build."
	case cmdChcon:
		return "CONTEXT FILE...", "Change the SELinux security context of files. Requires CAP_FOWNER/CAP_MAC_ADMIN; intentionally gated in this build."
	case cmdRuncon:
		return "CONTEXT PROG [ARG]...", "Run PROG in the given SELinux context. Requires a loaded policy and permission to transition; intentionally gated in this build."
	case cmdRestorecon:
		return "FILE...", "Restore the default SELinux contexts on FILEs from policy. Requires CAP_MAC_ADMIN; intentionally gated in this build."
	case cmdSetfiles:
		return "SPEC_FILE FILE...", "Set file SELinux contexts according to a file-contexts SPEC_FILE. Requires CAP_MAC_ADMIN; intentionally gated in this build."
	case cmdLoadPolicy:
		return "", "Load a new SELinux policy into the running kernel. Requires CAP_MAC_ADMIN; intentionally gated in this build."
	}
	return "", "Privileged SELinux operation, intentionally gated in this build."
}

// exampleArgs returns representative operands for a privileged applet's worked
// --help example.
func exampleArgs(name string) string {
	switch name {
	case cmdSetenforce:
		return "Permissive"
	case cmdSetsebool:
		return "httpd_can_network_connect on"
	case cmdChcon:
		return "-t httpd_sys_content_t /var/www/index.html"
	case cmdRuncon:
		return "system_u:system_r:httpd_t /usr/sbin/httpd"
	case cmdRestorecon:
		return "-R /var/www"
	case cmdSetfiles:
		return "file_contexts /var/www"
	case cmdLoadPolicy:
		return "" // load_policy takes no operands
	}
	return ""
}
