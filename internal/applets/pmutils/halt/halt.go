//
// mimixbox/internal/applets/pmutils/halt/halt.go
//
// Copyright 2021 Naohiro CHIKAMATSU, polynomialspace
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

// Package halt implements the halt, poweroff and reboot applets: stop the
// system. The same package backs all three command names; the action performed
// depends on the name the command was constructed with.
package halt

import (
	"context"
	"fmt"
	"os"
	"syscall"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command names served by this package.
const (
	nameHalt     = "halt"
	namePoweroff = "poweroff"
	nameReboot   = "reboot"
)

// rebootFn is the dangerous syscall, behind a package variable so tests can
// stub it and never actually stop the machine.
var rebootFn = syscall.Reboot

// isRoot reports whether the current process has the privilege required to stop
// the system. It is a package variable so tests can simulate root.
var isRoot = func() bool {
	return os.Geteuid() == 0 && os.Getuid() == 0
}

// Command is the halt/poweroff/reboot applet. It carries the name it was
// invoked as so the same type can serve all three commands.
type Command struct {
	name string
}

// New returns a command for the given name (one of "halt", "poweroff" or
// "reboot").
func New(name string) *Command { return &Command{name: name} }

// NewHalt returns a halt command.
func NewHalt() *Command { return New(nameHalt) }

// NewPoweroff returns a poweroff command.
func NewPoweroff() *Command { return New(namePoweroff) }

// NewReboot returns a reboot command.
func NewReboot() *Command { return New(nameReboot) }

// Name returns the command name.
func (c *Command) Name() string { return c.name }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string {
	switch c.name {
	case namePoweroff:
		return "Power off the system"
	case nameReboot:
		return "Reboot the system"
	default:
		return "Halt the system"
	}
}

// action returns the syscall.Reboot command constant for this command name.
func (c *Command) action() int {
	switch c.name {
	case nameReboot:
		return syscall.LINUX_REBOOT_CMD_RESTART
	case namePoweroff:
		return syscall.LINUX_REBOOT_CMD_POWER_OFF
	default:
		return syscall.LINUX_REBOOT_CMD_HALT
	}
}

// Run executes halt/poweroff/reboot. It requires root; otherwise it prints a
// permission message and returns a silent failure without touching rebootFn.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]", stdio.Err)
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	if !isRoot() {
		_, _ = fmt.Fprintf(stdio.Err, "%s: you must be root to %s the system\n", c.Name(), c.Name())
		return command.SilentFailure()
	}

	return stop(c.action())
}

// stop synchronizes filesystems and performs the requested reboot action via
// the replaceable rebootFn so tests stay safe.
func stop(action int) error {
	syscall.Sync()
	if err := rebootFn(action); err != nil {
		return command.Failure(err)
	}
	return nil
}
