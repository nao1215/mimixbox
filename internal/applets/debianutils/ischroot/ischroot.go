//
// mimixbox/internal/applets/debianutils/ischroot/ischroot.go
//
// Copyright 2021 Naohiro CHIKAMATSU
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package ischroot implements the ischroot applet: detect whether the current
// process is running inside a chroot.
package ischroot

import (
	"context"
	"os"
	"strings"
	"syscall"

	"github.com/nao1215/mimixbox/internal/command"
	mb "github.com/nao1215/mimixbox/internal/lib"
)

// Exit codes, matching Debian's ischroot: 0 if running in a chroot, 1 if not,
// and 2 if it cannot be detected (e.g. not enough privileges).
const (
	jail         = 0 // running in a chroot
	notJail      = 1 // not running in a chroot
	notSuperUser = 2 // chroot status could not be detected
)

// Command is the ischroot applet.
type Command struct{}

// New returns an ischroot command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "ischroot" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Detect if running in a chroot" }

// Run executes ischroot. It returns nil when running in a chroot, and an
// *command.ExitError carrying status 1 (not a chroot) or 2 (undetectable)
// otherwise.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]", stdio.Err).WithHelp(command.Help{
		Description: "Detect whether the current process is running in a chroot. The result is " +
			"reported through the exit status. When the status cannot be determined, -f and " +
			"-t choose the fallback answer.",
		Examples: []command.Example{
			{Command: "ischroot", Explain: "Exit 0 inside a chroot, 1 outside, 2 if undetermined."},
			{Command: "ischroot -f", Explain: "Assume not in a chroot when detection fails."},
			{Command: "ischroot -t", Explain: "Assume in a chroot when detection fails."},
		},
		ExitStatus: "0  running in a chroot.\n1  not in a chroot.\n2  could not be determined.",
	})
	defaultFalse := fs.BoolP("default-false", "f", false, "return 1 if detection fails (not root)")
	defaultTrue := fs.BoolP("default-true", "t", false, "return 0 if detection fails (not root)")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	code := detect(*defaultFalse, *defaultTrue)
	switch code {
	case jail:
		return nil
	default:
		return &command.ExitError{Code: code}
	}
}

// detect returns the chroot exit code, applying the -f/-t fallbacks when the
// status cannot be determined.
func detect(defaultFalse, defaultTrue bool) int {
	if isFakeChroot() {
		return jail
	}

	exitCode := isChroot()
	if exitCode == notSuperUser {
		if defaultFalse {
			exitCode = notJail
		} else if defaultTrue {
			exitCode = jail
		}
	}
	return exitCode
}

// isFakeChroot reports whether the environment is a FAKECHROOT environment.
// In the FAKECHROOT environment, the library that overwrites the libc (glibc) is preloaded.
// Specifically, the preloaded library is libfakechroot.so. Whether it is preloaded or not
// can be determined by getting the path of libfakechroot.so from the environment variable LD_PRELOAD.
// The libc (glibc) function has been redefined in libfakechroot.so.
// If you run the app with LD_PRELOAD, the app will run using the redefined functions.
func isFakeChroot() bool {
	fakeChroot := os.Getenv("FAKECHROOT")
	if fakeChroot != "true" {
		return false
	}
	fakeChrootBase := os.Getenv("FAKECHROOT_BASE")
	if fakeChrootBase == "" {
		return false
	}
	ldPreload := os.Getenv("LD_PRELOAD")
	return strings.Contains(ldPreload, "libfakechroot.so")
}

func isChroot() int {
	if !canStatRootDir() {
		return notSuperUser
	}

	if !canStatInitProcessRootDir() {
		if !canLstatInitProcessRootDir() {
			return notSuperUser
		}
		if !mb.IsRootUser() {
			return notSuperUser
		}
		// User is root. However, root can't stat "/proc/1/root". It's jail.
		return jail
	}

	if isNotJail() {
		return notJail
	}
	return jail
}

func canStatRootDir() bool {
	_, err := os.Stat("/")
	return err == nil
}

func canStatInitProcessRootDir() bool {
	_, err := os.Stat("/proc/1/root")
	return err == nil
}

func canLstatInitProcessRootDir() bool {
	_, err := os.Lstat("/proc/1/root")
	return err == nil
}

func isNotJail() bool {
	rootStatInfo, err := os.Stat("/")
	if err != nil {
		return false
	}
	internalRootStat := rootStatInfo.Sys().(*syscall.Stat_t)

	procStatInfo, err := os.Stat("/proc/1/root")
	if err != nil {
		return false
	}
	internalProcStat := procStatInfo.Sys().(*syscall.Stat_t)

	return (internalRootStat.Ino == internalProcStat.Ino) && (internalRootStat.Dev == internalProcStat.Dev)
}
