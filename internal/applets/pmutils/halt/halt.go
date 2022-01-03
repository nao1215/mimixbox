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
package halt

import (
	"errors"
	"fmt"
	"os"
	"syscall"

	mb "github.com/nao1215/mimixbox/internal/lib"

	"github.com/jessevdk/go-flags"
)

var cmdName string = "halt"

const version = "1.0.1"

var osExit = os.Exit

// Exit code
const (
	ExitSuccess int = iota // 0
	ExitFailure
)

type haltOpts struct {
	Version bool `short:"v" long:"version" description:"Show halt command version"`
}

type poweroffOpts struct {
	Version bool `short:"v" long:"version" description:"Show poweroff command version"`
}

type rebootOpts struct {
	Version bool `short:"v" long:"version" description:"Show reboot command version"`
}

type allOptions struct {
	halt   haltOpts
	po     poweroffOpts
	reboot rebootOpts
}

func Run() (int, error) {
	var allOpts allOptions = allOptions{}
	var args []string
	var err error

	setCmdName(os.Args[0])
	if args, err = parseArgs(&allOpts); err != nil {
		return ExitFailure, nil
	}

	switch cmdName {
	case "halt":
		return halt(args, allOpts.halt)
	case "poweroff":
		return poweroff(args, allOpts.po)
	case "reboot":
		return reboot(args, allOpts.reboot)
	}
	return ExitFailure, errors.New("mimixbox failed to parse the argument (not halt, poweroff, reboot error)")
}

func halt(args []string, opts haltOpts) (int, error) {
	fmt.Fprintln(os.Stdout, "The system is going down NOW !!")

	recordWtmp()
	if err := powerOffSystem(); err != nil {
		return ExitFailure, err
	}
	return ExitSuccess, nil
}

func poweroff(args []string, opts poweroffOpts) (int, error) {
	if err := powerOffSystem(); err != nil {
		return ExitFailure, err
	}
	return ExitSuccess, nil
}

func reboot(args []string, opts rebootOpts) (int, error) {
	if err := rebootSystem(); err != nil {
		return ExitFailure, err
	}
	return ExitSuccess, nil
}

func powerOffSystem() error {
	process, err := os.FindProcess(1)
	if err != nil {
		return err
	}
	err = process.Signal(syscall.Signal(mb.ConvSignalNameToNum("SIGUSR1")))
	if err != nil {
		return err
	}
	syscall.Sync()
	// 0x4321fedc == LINUX_REBOOT_CMD_POWER_OFF; see reboot(2)
	// LINUX_REBOOT_CMD_HALT is semantically correct, but
	// implementations vary (halt(8)), and most users will
	// want power off.
	return syscall.Reboot(syscall.LINUX_REBOOT_CMD_POWER_OFF)
}

func rebootSystem() error {
	process, err := os.FindProcess(1)
	if err != nil {
		return err
	}
	err = process.Signal(syscall.Signal(mb.ConvSignalNameToNum("SIGUSR2")))
	if err != nil {
		return err
	}
	syscall.Sync()
	return syscall.Reboot(syscall.LINUX_REBOOT_CMD_RESTART)
}

func recordWtmp() {
	return // TODO:
}

func setCmdName(name string) {
	cmdName = name
}

func parseArgs(opts *allOptions) ([]string, error) {
	p := initParser(opts)

	args, err := p.Parse()
	if err != nil {
		return nil, err
	}
	showVersionAndExitIfNeeded(opts)
	return args, nil
}

func initParser(opts *allOptions) *flags.Parser {
	parser := flags.NewParser(opts, flags.Default)
	parser.Name = cmdName
	parser.Usage = "[OPTIONS]"
	return parser
}

func showVersionAndExitIfNeeded(opts *allOptions) {
	if opts.halt.Version || opts.po.Version || opts.reboot.Version {
		mb.ShowVersion(cmdName, version)
		osExit(ExitSuccess)
	}
}
