package command

import "github.com/nao1215/mimixbox/internal/version"

// HandleHelpVersion implements the standard --help/--version contract for
// applets that parse their own arguments instead of using a FlagSet (echo,
// true, false, and similar custom parsers). When the first argument is exactly
// "--help" it writes a GNU-style usage block built from name and usage; when it
// is "--version" it writes the version line. It reports whether it handled the
// argument so the caller can return early.
//
// Only the first argument is inspected, matching the behavior of GNU's
// standalone echo and true: a "--help" or "--version" that appears later is an
// ordinary operand (so `echo foo --help` still prints "foo --help"). Keeping
// this logic in one place is what stops custom-parsed applets from drifting
// into inconsistent --help/--version behavior.
func HandleHelpVersion(stdio IO, name, usage string, args []string) bool {
	if len(args) == 0 {
		return false
	}
	switch args[0] {
	case "--help":
		NewFlagSet(name, usage, stdio.Err).WriteUsage(stdio.Out)
		return true
	case "--version":
		version.Print(stdio.Out, name)
		return true
	}
	return false
}
