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
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"

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

	// Symlink invocation: the binary was called as an applet name.
	if invoked != cmdName {
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

// runApplet runs the named applet with args. The applet entry points still read
// the process-level os.Args and streams (via internal/command.Adapt), so os.Args
// is set to "<applet> <args...>" before the call. runApplet is a variable so
// tests can substitute a fake dispatcher.
var runApplet = func(name string, args []string, stdio command.IO) int {
	app, ok := applets.Applets[name]
	if !ok {
		return unsupported(name, stdio)
	}
	os.Args = append([]string{name}, args...)
	status, err := app.Ep()
	if err != nil {
		_, _ = fmt.Fprintln(stdio.Err, name+": "+err.Error())
	}
	return status
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
		applets.ListAppletsTo(stdio.Out)
		return command.ExitSuccess
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
func unsupported(name string, stdio command.IO) int {
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
`)
}

func install(mimixboxPath string, installPath string, full bool, stdio command.IO) error {
	targetPath := os.ExpandEnv(installPath)
	if !mb.IsDir(targetPath) {
		return errors.New(targetPath + ": no such directory")
	}

	mimixboxAbsPath, err := resolveSelf(mimixboxPath)
	if err != nil {
		return err
	}

	for _, applet := range applets.SortApplet() {
		if !full && mb.ExistCmd(applet) {
			_, _ = fmt.Fprintf(stdio.Err, "Same name command(%s) already exists. Not create symbolic link.\n", applet)
			continue // if same name command already exists, not install for safety.
		}

		// If a symbolic link with the same name already exists, delete it. If a
		// real binary has the same name, leave it: the former is likely ours,
		// the latter probably belongs to another package.
		newPath := filepath.Join(targetPath, applet)
		if mb.IsSymlink(newPath) {
			if err := os.Remove(newPath); err != nil {
				_, _ = fmt.Fprintln(stdio.Err, err)
				continue
			}
			_, _ = fmt.Fprintf(stdio.Out, "Delete              : %s\n", newPath)
		}

		if err := os.Symlink(mimixboxAbsPath, newPath); err != nil {
			_, _ = fmt.Fprintln(stdio.Err, err)
			continue
		}
		_, _ = fmt.Fprintf(stdio.Out, "Create symbolic link: %s\n", newPath)
	}
	return nil
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
