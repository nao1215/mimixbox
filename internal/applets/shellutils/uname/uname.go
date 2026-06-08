// Package uname implements the uname applet: print system information such as
// the kernel name, hostname, release, version and machine architecture.
package uname

import (
	"context"
	"fmt"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
	"golang.org/x/sys/unix"
)

// Command is the uname applet.
type Command struct{}

// New returns a uname command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "uname" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Print system information" }

// info holds the pieces of system information uname can print.
type info struct {
	sysname  string
	nodename string
	release  string
	version  string
	machine  string
	os       string
}

// sysInfo is the source of system information; tests replace it.
var sysInfo = uts

// Run executes uname.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]...", stdio.Err)
	all := fs.BoolP("all", "a", false, "print all information")
	sysname := fs.BoolP("kernel-name", "s", false, "print the kernel name")
	nodename := fs.BoolP("nodename", "n", false, "print the network node hostname")
	release := fs.BoolP("kernel-release", "r", false, "print the kernel release")
	version := fs.BoolP("kernel-version", "v", false, "print the kernel version")
	machine := fs.BoolP("machine", "m", false, "print the machine hardware name")
	osName := fs.BoolP("operating-system", "o", false, "print the operating system")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	in, err := sysInfo()
	if err != nil {
		return command.Failuref("cannot get system information: %v", err)
	}

	var parts []string
	if *all {
		parts = []string{in.sysname, in.nodename, in.release, in.version, in.machine, in.os}
	} else {
		if *sysname {
			parts = append(parts, in.sysname)
		}
		if *nodename {
			parts = append(parts, in.nodename)
		}
		if *release {
			parts = append(parts, in.release)
		}
		if *version {
			parts = append(parts, in.version)
		}
		if *machine {
			parts = append(parts, in.machine)
		}
		if *osName {
			parts = append(parts, in.os)
		}
		if len(parts) == 0 {
			parts = []string{in.sysname}
		}
	}

	if _, err := fmt.Fprintln(stdio.Out, strings.Join(parts, " ")); err != nil {
		return command.Failure(err)
	}
	return nil
}

// uts reads the real system information via the uname(2) system call.
func uts() (info, error) {
	var u unix.Utsname
	if err := unix.Uname(&u); err != nil {
		return info{}, err
	}
	return info{
		sysname:  charsToString(u.Sysname[:]),
		nodename: charsToString(u.Nodename[:]),
		release:  charsToString(u.Release[:]),
		version:  charsToString(u.Version[:]),
		machine:  charsToString(u.Machine[:]),
		os:       "GNU/Linux",
	}, nil
}

// charsToString turns a NUL-terminated C character array into a Go string.
func charsToString(ca []byte) string {
	n := 0
	for n < len(ca) && ca[n] != 0 {
		n++
	}
	return string(ca[:n])
}
