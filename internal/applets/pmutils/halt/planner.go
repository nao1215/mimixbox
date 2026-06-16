//
// mimixbox/internal/applets/pmutils/halt/planner.go
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

package halt

import (
	"syscall"

	"github.com/nao1215/mimixbox/internal/command"
)

// This file is the shutdown action planner and reboot executor seam: action
// decides which syscall.Reboot command the parsed options map to, and stop runs
// the optional sync followed by the (replaceable) reboot syscall. The wtmp
// encoding lives in wtmp.go and the CLI wiring in halt.go; behavior, exit codes,
// and messages are unchanged.

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
