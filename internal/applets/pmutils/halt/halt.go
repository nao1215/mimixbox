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
// depends on the name the command was constructed with. The options follow the
// util-linux/sysvinit man pages (-f, -n, -w, -d, and -p for halt).
package halt

import (
	"context"
	"encoding/binary"
	"fmt"
	"os"
	"syscall"
	"time"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command names served by this package.
const (
	nameHalt     = "halt"
	namePoweroff = "poweroff"
	nameReboot   = "reboot"
)

// wtmpFile is the login-record file a shutdown event is written to. It is a
// package variable so tests can point it at a temporary file.
var wtmpFile = "/var/log/wtmp"

// rebootFn is the dangerous syscall, behind a package variable so tests can
// stub it and never actually stop the machine.
var rebootFn = syscall.Reboot

// syncFn flushes filesystem buffers; a variable so tests observe it.
var syncFn = syscall.Sync

// isRoot reports whether the current process has the privilege required to stop
// the system. It is a package variable so tests can simulate root.
var isRoot = func() bool {
	return os.Geteuid() == 0 && os.Getuid() == 0
}

// nowFn returns the current time; a variable so tests get a deterministic clock.
var nowFn = time.Now

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

type options struct {
	force    bool // -f: force, do not sync or wait
	noSync   bool // -n: do not sync before the action
	wtmpOnly bool // -w: only write the wtmp record, do not stop the system
	noWtmp   bool // -d: do not write the wtmp record
	poweroff bool // -p: when called as halt, power off instead of halting
}

// action returns the syscall.Reboot command constant for this command, honoring
// "halt -p" (power off instead of halt).
func (c *Command) action(opts options) int {
	switch c.name {
	case nameReboot:
		return syscall.LINUX_REBOOT_CMD_RESTART
	case namePoweroff:
		return syscall.LINUX_REBOOT_CMD_POWER_OFF
	default:
		if opts.poweroff {
			return syscall.LINUX_REBOOT_CMD_POWER_OFF
		}
		return syscall.LINUX_REBOOT_CMD_HALT
	}
}

// Run executes halt/poweroff/reboot. It requires root; otherwise it prints a
// permission message and returns a silent failure without touching rebootFn.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]", stdio.Err).WithHelp(command.Help{
		Description: "Halt, power off, or reboot the machine (this applet runs as " + c.Name() + "). It " +
			"writes a wtmp shutdown record and then asks the kernel to stop or restart the system. " +
			"Root privileges are required.",
		Examples: []command.Example{
			{Command: c.Name(), Explain: "Sync and " + c.Name() + " the system (requires root)."},
			{Command: c.Name() + " -f", Explain: "Skip the sync and act immediately."},
		},
		ExitStatus: "0  the shutdown request was issued.\n1  not run as root, or the request failed.",
	})
	force := fs.BoolP("force", "f", false, "force immediate halt/power-off/reboot; do not sync")
	noSync := fs.BoolP("no-sync", "n", false, "do not sync before halt or reboot")
	wtmpOnly := fs.BoolP("wtmp-only", "w", false, "only write the wtmp shutdown record, do not stop the system")
	noWtmp := fs.BoolP("no-wtmp", "d", false, "do not write the wtmp shutdown record")
	var poweroff *bool
	if c.name == nameHalt {
		poweroff = fs.BoolP("poweroff", "p", false, "power off the machine instead of halting")
	}

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	opts := options{force: *force, noSync: *noSync, wtmpOnly: *wtmpOnly, noWtmp: *noWtmp}
	if poweroff != nil {
		opts.poweroff = *poweroff
	}

	if !isRoot() {
		_, _ = fmt.Fprintf(stdio.Err, "%s: you must be root to %s the system\n", c.Name(), c.Name())
		return command.SilentFailure()
	}

	if !opts.noWtmp {
		if werr := writeWtmp(wtmpFile, nowFn()); werr != nil {
			// Match util-linux: a wtmp failure is reported but does not abort.
			_, _ = fmt.Fprintf(stdio.Err, "%s: cannot write %s: %v\n", c.Name(), wtmpFile, werr)
		}
	}

	if opts.wtmpOnly {
		return nil
	}

	return c.stop(opts)
}

// stop syncs filesystems (unless suppressed) and performs the requested action
// via the replaceable rebootFn so tests stay safe.
func (c *Command) stop(opts options) error {
	if !opts.noSync && !opts.force {
		syncFn()
	}
	if err := rebootFn(c.action(opts)); err != nil {
		return command.Failure(err)
	}
	return nil
}

// utmpRecordSize is the size of a Linux struct utmp record.
const utmpRecordSize = 384

// utmp record type for a run-level (shutdown) entry.
const runLevel = 1

// writeWtmp appends a "shutdown" login record to path, as sysvinit's halt does,
// so utilities such as "who" and "last" can report the shutdown time. The
// record layout matches Linux's struct utmp.
func writeWtmp(path string, now time.Time) error {
	rec := make([]byte, utmpRecordSize)
	binary.LittleEndian.PutUint16(rec[0:], runLevel)                        // ut_type
	copy(rec[8:40], "~~")                                                   // ut_line[32]
	copy(rec[40:44], "~~")                                                  // ut_id[4]
	copy(rec[44:76], "shutdown")                                            // ut_user[32]
	binary.LittleEndian.PutUint32(rec[340:], uint32(now.Unix()))            // ut_tv.tv_sec
	binary.LittleEndian.PutUint32(rec[344:], uint32(now.Nanosecond()/1000)) // ut_tv.tv_usec

	f, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644) //nolint:gosec // wtmp is world-readable
	if err != nil {
		return err
	}
	defer func() {
		if cerr := f.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()
	_, err = f.Write(rec)
	return err
}
