// mimixbox/cmd/mimixbox/main.go
//
// # Copyright 2021 Naohiro CHIKAMATSU
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/nao1215/mimixbox/internal/applets"
	"github.com/nao1215/mimixbox/internal/command"
	"github.com/nao1215/mimixbox/internal/version"

	mb "github.com/nao1215/mimixbox/internal/lib"
)

const cmdName string = "mimixbox"

var osExit = os.Exit

func main() {
	stdio := command.IO{In: os.Stdin, Out: os.Stdout, Err: os.Stderr}
	osExit(run(os.Args, stdio))
}

// run is the whole program as a testable function: it returns the process exit
// code instead of calling os.Exit, and writes through the injected IO instead
// of touching the process streams directly.
//
// Dispatch is decided by the first argument. When MimixBox is invoked through a
// symlink (argv[0] is an applet name), or when its first argument is a known
// applet, that applet runs and receives the remaining arguments — so an
// applet's own flags (cat --help, cp -f, ...) always reach the applet. Only
// when the first argument is not an applet are MimixBox's own options parsed.
func run(argv []string, stdio command.IO) int {
	invoked := path.Base(argv[0])

	// Symlink invocation: the binary was called through a name that matches a
	// known applet (e.g. a "cat" symlink). The basename alone is not enough —
	// the binary may simply have been renamed, copied, or wrapped under another
	// name (mimixbox-review, mimixbox.bin, ...). In that case it is still the
	// MimixBox driver, so fall through to normal option/applet parsing instead
	// of treating the binary name as an unknown applet (issue #949).
	if invoked != cmdName && applets.HasApplet(invoked) {
		return runApplet(invoked, argv[1:], stdio)
	}

	rest := argv[1:]
	if len(rest) == 0 {
		// No applet and no option: a usage error. Help goes to stderr.
		writeHelp(stdio.Err)
		return command.ExitFailure
	}

	first := rest[0]
	if applets.HasApplet(first) {
		return runApplet(first, rest[1:], stdio)
	}

	return runOption(first, rest[1:], stdio)
}

// runApplet runs the named applet with args. The applet is executed through
// internal/command.Execute with the injected IO, so dispatch touches neither
// os.Args nor the process streams and can be exercised entirely in memory.
// runApplet is a variable so tests can substitute a fake dispatcher.
var runApplet = func(name string, args []string, stdio command.IO) int {
	app, ok := applets.Applets[name]
	if !ok {
		return unsupported(name, stdio)
	}
	return command.Execute(context.Background(), app.Cmd, stdio, args)
}

// runOption handles MimixBox's own options (everything that is not an applet).
func runOption(first string, params []string, stdio command.IO) int {
	switch first {
	case "-h", "--help":
		writeHelp(stdio.Out)
		return command.ExitSuccess
	case "-v", "--version":
		version.Print(stdio.Out, cmdName)
		return command.ExitSuccess
	case "-l", "--list":
		return runList(params, stdio)
	case "-i", "--install":
		return runInstall(params, stdio, false)
	case "-f", "--full-install":
		return runInstall(params, stdio, true)
	case "-r", "--remove":
		return runRemove(params, stdio)
	default:
		return unsupported(first, stdio)
	}
}

// runList implements -l/--list, including JSON output and filtering.
//
// Accepted parameters (any order):
//
//	--json                  emit a JSON array instead of the text table
//	--filter=PREFIX         keep only applets whose name starts with PREFIX
//	--subsystem=NAME        keep only applets in subsystem NAME
//	PREFIX or PREFIX*       bare prefix operand (trailing "*" glob optional)
//
// With no parameters the behavior is unchanged: the full text table on stdout.
func runList(params []string, stdio command.IO) int {
	jsonOut := false
	var filter applets.ListFilter
	for _, p := range params {
		switch {
		case p == "--json":
			jsonOut = true
		case strings.HasPrefix(p, "--filter="):
			filter.Prefix = strings.TrimPrefix(p, "--filter=")
		case strings.HasPrefix(p, "--subsystem="):
			filter.Subsystem = strings.TrimPrefix(p, "--subsystem=")
		case strings.HasPrefix(p, "-"):
			_, _ = fmt.Fprintf(stdio.Err, "%s: --list: unknown option %q\n", cmdName, p)
			return command.ExitFailure
		default:
			// A bare operand is a name prefix; an optional trailing "*" glob is
			// accepted so both "cat" and "cat*" select cat-prefixed applets.
			filter.Prefix = strings.TrimSuffix(p, "*")
		}
	}

	if jsonOut {
		if err := applets.ListAppletsJSONTo(stdio.Out, filter); err != nil {
			_, _ = fmt.Fprintln(stdio.Err, err)
			return command.ExitFailure
		}
		return command.ExitSuccess
	}
	applets.ListAppletsTo(stdio.Out, filter)
	return command.ExitSuccess
}

// runInstall implements -i/--install (full=false) and -f/--full-install
// (full=true). The single parameter is the directory to populate; it may share
// a basename with an applet because options are no longer guessed from arg names.
func runInstall(params []string, stdio command.IO, full bool) int {
	if len(params) != 1 {
		_, _ = fmt.Fprintf(stdio.Err, "%s: install requires a single DIRECTORY operand\n", cmdName)
		return command.ExitFailure
	}
	if err := install(os.Args[0], params[0], full, stdio); err != nil {
		_, _ = fmt.Fprintln(stdio.Err, err)
		return command.ExitFailure
	}
	return command.ExitSuccess
}

// runRemove implements -r/--remove.
func runRemove(params []string, stdio command.IO) int {
	if len(params) != 1 {
		_, _ = fmt.Fprintf(stdio.Err, "%s: remove requires a single DIRECTORY operand\n", cmdName)
		return command.ExitFailure
	}
	if err := remove(os.Args[0], params[0], stdio); err != nil {
		_, _ = fmt.Fprintln(stdio.Err, err)
		return command.ExitFailure
	}
	return command.ExitSuccess
}

// unsupported reports an unknown command/option and the supported applet list,
// both on stderr so a script's stdout is never polluted by an error.
//
// When the input looks like a command name (not a "-"/"--" option), the
// error is shown error-first with up to five nearest-applet suggestions by
// Levenshtein distance, e.g.
//
//	mimixbox: 'lss' is not a mimixbox command. Did you mean: ls?
//
// The full applet wall is still printed afterwards, but only as a secondary
// fallback below the concise suggestion line.
func unsupported(name string, stdio command.IO) int {
	if !strings.HasPrefix(name, "-") {
		_, _ = fmt.Fprintf(stdio.Err, "%s: '%s' is not a mimixbox command.", cmdName, name)
		if suggestions := applets.SuggestApplets(name, 5); len(suggestions) > 0 {
			_, _ = fmt.Fprintf(stdio.Err, " Did you mean: %s?", strings.Join(suggestions, ", "))
		}
		_, _ = fmt.Fprint(stdio.Err, "\n\n")
		_, _ = fmt.Fprintln(stdio.Err, "[Commands supported by MimixBox]")
		applets.ShowAppletsBySpaceSeparatedTo(stdio.Err)
		return command.ExitFailure
	}

	_, _ = fmt.Fprintf(stdio.Err, "%s: %q is not a mimixbox command or option\n\n", cmdName, name)
	_, _ = fmt.Fprintln(stdio.Err, "[Commands supported by MimixBox]")
	applets.ShowAppletsBySpaceSeparatedTo(stdio.Err)
	return command.ExitFailure
}

// writeHelp prints the top-level usage in the same GNU style the applets use.
func writeHelp(w io.Writer) {
	_, _ = fmt.Fprint(w, `Usage: mimixbox [OPTION] | mimixbox APPLET [ARG]... | APPLET [ARG]...

MimixBox packs many Unix commands (applets) into a single binary. Run an applet
by name, either as "mimixbox APPLET ..." or through a symlink named APPLET.

Options:
  -i, --install DIR        create symlinks for applets that are not already on the system
  -f, --full-install DIR   create symlinks for every applet, regardless of system state
  -r, --remove DIR         remove the symlinks MimixBox created in DIR
  -l, --list               list the applets MimixBox provides
  -v, --version            print version information and exit
  -h, --help               print this help and exit

List options (used after -l/--list):
  --json                   emit the applet list as a JSON array
  --filter=PREFIX          show only applets whose name starts with PREFIX
                           (a bare "PREFIX" or "PREFIX*" operand works too)
  --subsystem=NAME         show only applets in subsystem NAME (e.g. textutils)

Examples:
  mimixbox --list                    Show every applet and its one-line description.
  mimixbox --list --json             Show every applet as a JSON array.
  mimixbox --list --filter=ca        Show only applets whose name starts with "ca".
  mimixbox --list --subsystem=textutils   Show only the textutils applets.
  mimixbox cat file.txt              Run the cat applet directly.
  cat file.txt                       Same, when invoked through an installed symlink.
  mimixbox APPLET --help             Show the applet's own help, options, and examples.
  sudo mimixbox --full-install /usr/local/bin   Install a symlink for every applet.

Run "mimixbox APPLET --help" for an applet's description, options, and examples.
`)
}

// slot classifies what currently occupies an applet name in the target
// directory, so install() can decide what to do without ever consulting the
// host PATH.
type slot int

const (
	slotFree    slot = iota // nothing occupies the name; safe to create
	slotOwned               // a symlink that already points at this MimixBox
	slotForeign             // a real file, or a symlink owned by something else
)

// classifySlot reports the state of path relative to the running binary self.
// Anything that is not provably ours (a real file, a foreign symlink, or an
// entry we cannot inspect) is treated as foreign so it is never removed.
func classifySlot(path, self string) slot {
	fi, err := os.Lstat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return slotFree
		}
		return slotForeign
	}
	if fi.Mode()&os.ModeSymlink == 0 {
		return slotForeign // a real file or directory occupies the name
	}
	target, err := os.Readlink(path)
	if err != nil {
		return slotForeign
	}
	if ownedBySelf(target, self) {
		return slotOwned
	}
	return slotForeign
}

func install(mimixboxPath string, installPath string, full bool, stdio command.IO) error {
	targetPath := os.ExpandEnv(installPath)
	if !mb.IsDir(targetPath) {
		return errors.New(targetPath + ": no such directory")
	}

	self, err := resolveSelf(mimixboxPath)
	if err != nil {
		return err
	}

	for _, applet := range applets.SortApplet() {
		newPath := filepath.Join(targetPath, applet)
		// The decision depends only on the target directory and explicit
		// ownership, never on whether a same-named command exists somewhere on
		// the host PATH (issue #948).
		switch classifySlot(newPath, self) {
		case slotFree:
			createLink(self, newPath, stdio)
		case slotOwned:
			// Already an up-to-date MimixBox symlink. Plain --install leaves it
			// alone; --full-install refreshes it so its target is corrected if
			// the binary moved.
			if full {
				if err := os.Remove(newPath); err != nil {
					_, _ = fmt.Fprintln(stdio.Err, err)
					continue
				}
				_, _ = fmt.Fprintf(stdio.Out, "Delete              : %s\n", newPath)
				createLink(self, newPath, stdio)
			}
		case slotForeign:
			// A real file or a symlink owned by another package occupies this
			// name. Never remove it; both install modes skip it for safety.
			_, _ = fmt.Fprintf(stdio.Err, "%s already exists and is not owned by MimixBox. Not creating symbolic link.\n", newPath)
		}
	}
	return nil
}

// createLink creates a DIR/applet symlink pointing at the running binary and
// reports the action, or the error if it fails.
func createLink(self, newPath string, stdio command.IO) {
	if err := os.Symlink(self, newPath); err != nil {
		_, _ = fmt.Fprintln(stdio.Err, err)
		return
	}
	_, _ = fmt.Fprintf(stdio.Out, "Create symbolic link: %s\n", newPath)
}

// osExecutable is os.Executable, indirected so tests can substitute it.
var osExecutable = os.Executable

// resolveSelf returns the absolute path of the exact MimixBox binary that is
// running now. install() must link applet symlinks to this binary, not to some
// other "mimixbox" that happens to be earlier on PATH, so that --install always
// targets the binary the user actually invoked or just installed. invoked is the
// argv[0] fallback used only when the executable path cannot be determined.
func resolveSelf(invoked string) (string, error) {
	if p, err := osExecutable(); err == nil {
		return filepath.Clean(p), nil
	}
	return filepath.Abs(invoked)
}

func remove(mimixboxPath string, installPath string, stdio command.IO) error {
	targetPath := os.ExpandEnv(installPath)
	if !mb.IsDir(targetPath) {
		return errors.New(targetPath + ": no such directory")
	}

	self, err := resolveSelf(mimixboxPath)
	if err != nil {
		return err
	}

	for _, name := range applets.SortApplet() {
		symbolicPath := filepath.Join(targetPath, name)
		if !mb.IsSymlink(symbolicPath) {
			continue
		}
		realPath, err := os.Readlink(symbolicPath)
		if err != nil {
			_, _ = fmt.Fprintln(stdio.Err, err)
			continue
		}
		// Only remove symlinks provably owned by this MimixBox install: their
		// target must be exactly the running binary. A foreign symlink whose
		// target merely contains "mimixbox" (e.g. cat -> /opt/other-mimixbox)
		// is left untouched.
		if !ownedBySelf(realPath, self) {
			continue
		}
		if err := os.Remove(symbolicPath); err != nil {
			_, _ = fmt.Fprintln(stdio.Err, err)
			continue
		}
		_, _ = fmt.Fprintf(stdio.Out, "Delete symbolic link: %s\n", symbolicPath)
	}
	return nil
}

// ownedBySelf reports whether a symlink target points at the exact MimixBox
// binary identified by self. Both paths are cleaned so equivalent spellings
// compare equal.
func ownedBySelf(target, self string) bool {
	return filepath.Clean(target) == filepath.Clean(self)
}
