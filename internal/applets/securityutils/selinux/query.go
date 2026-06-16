package selinux

import (
	"bufio"
	"fmt"
	"sort"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

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
