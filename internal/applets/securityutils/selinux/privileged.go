package selinux

import (
	"github.com/nao1215/mimixbox/internal/command"
)

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
