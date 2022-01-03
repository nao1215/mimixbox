//
// mimixbox/internal/applets/shellutils/chroot/chroot.go
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
package chroot

import (
	"os"
	"syscall"

	mb "github.com/nao1215/mimixbox/internal/lib"

	"github.com/jessevdk/go-flags"
)

const cmdName string = "chroot"
const version = "1.0.2"

var osExit = os.Exit

type options struct {
	Version bool `short:"v" long:"version" description:"Show chroot command version"`
	Help    bool `short:"h" long:"help" description:"Show this message"`
}

// Exit code
const (
	ExitSuccess int = iota // 0
	ExitFailure
)

type command struct {
	name       string
	withOption []string
	env        []string
}

func Run() (int, error) {
	var opts options
	var err error
	var cmd command

	parseArgs(&opts)

	err = syscall.Chroot(os.ExpandEnv(os.Args[1]))
	if err != nil {
		return ExitFailure, err
	}

	//----------------From here, in the prison-------------------
	err = os.Chdir("/")
	if err != nil {
		return ExitFailure, err
	}

	decideExecCommand(&cmd)
	// TODO: Reset UID and GID.
	// "/etc/passwd (uid name resolution file)" and "/etc/group (gid name resolution file)" may
	// be different between the original environment and the jail environment.
	// So, reset uid and gid in the jail environment.

	err = syscall.Exec(cmd.name, cmd.withOption, cmd.env)
	if err != nil {
		return ExitFailure, err
	}
	return ExitSuccess, nil
}

// Execute this method after chroot.
// If the user has not specified a command to be executed in the jail environment,
// the execution command is set to the environment variable $SHELL.
// If there is no $SHELL in the jail environment, use /bin/sh.
func decideExecCommand(cmd *command) error {
	if len(os.Args) == 2 {
		shell := os.Getenv("SHELL")
		if shell != "" && mb.ExistCmd(shell) {
			cmd.name = shell
		} else {
			cmd.name = "/bin/sh"
		}
		cmd.withOption = []string{cmd.name, "-i"}
	} else {
		cmd.name = os.Args[2]
		cmd.withOption = os.Args[2:]
	}

	// Reset the environment variable SHELL for the Jail environment.
	if err := os.Setenv("SHELL", cmd.name); err != nil {
		return err
	}

	cmd.env = os.Environ()
	return nil
}

func parseArgs(opts *options) {
	p := initParser(opts)

	if hasVersionOption() {
		mb.ShowVersion(cmdName, version)
		osExit(ExitSuccess)
	}
	if !isValidArgNr(os.Args) {
		showHelp(p)
		osExit(ExitFailure)
	}
}

func hasVersionOption() bool {
	for _, s := range os.Args[1:] {
		if s == "--version" || s == "-v" {
			return true
		}
	}
	return false
}

func initParser(opts *options) *flags.Parser {
	parser := flags.NewParser(opts, flags.Default)
	parser.Name = cmdName
	parser.Usage = "[OPTION] NEWROOT [COMMAND [ARG]...]"

	return parser
}

func isValidArgNr(args []string) bool {
	// 0:chroot, 1:root dir, 2:command(option)
	return len(args) >= 2
}

func showHelp(p *flags.Parser) {
	p.WriteHelp(os.Stdout)
}
