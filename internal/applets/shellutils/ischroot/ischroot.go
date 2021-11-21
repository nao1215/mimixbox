//
// mimixbox/internal/applets/shellutils/ischroot/ischroot.go
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
package ischroot

import (
	"fmt"
	"os"
	"strings"
	"syscall"

	mb "github.com/nao1215/mimixbox/internal/lib"

	"github.com/jessevdk/go-flags"
)

const cmdName string = "ischroot"
const version = "1.0.0"

var osExit = os.Exit

type options struct {
	DefaultFalse bool `short:"f" long:"--default-false" description:"Return false if user(not root user) use ischroot"`
	DefaultTrue  bool `short:"t" long:"--default-true" description:"Return true if user(not root user) use ischroot"`
	Version      bool `short:"v" long:"version" description:"Show ischroot command version"`
}

// Exit code
const (
	Jail    int = iota // 0
	NotJail            // NotJail = ExitFailuer
	NotSuperUser
)

func Run() (int, error) {
	var opts options
	var err error

	if _, err = parseArgs(&opts); err != nil {
		return NotJail, nil
	}

	if opts.DefaultFalse && opts.DefaultTrue {
		return NotJail, nil
	}

	if isFakeChroot() {
		return Jail, nil
	}

	exitCode := isChroot()
	if exitCode == NotSuperUser {
		if opts.DefaultFalse {
			exitCode = NotJail
		} else if opts.DefaultTrue {
			exitCode = Jail
		}
	}
	return exitCode, nil
}

// isFakeChroot() checks if the environment is a FAKECHROOT environment.
// In the FAKECHROOT environment, the library that overwrites the libc (glibc) is preloaded.
// Specifically, the preloaded library is libfakechroot.so. Whether it is preloaded or not
// can be determinedby getting the path of libfakechroot.so from the environment variable LD_PRELOAD.
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
		return NotSuperUser
	}

	if !canStatInitProcessRootDir() {
		if !canLstatInitProcessRootDir() {
			return NotSuperUser
		}
		if !mb.IsRootUser() {
			return NotSuperUser
		}
		// User is root. However, root can't stat "/proc/1/root". It's jail.
		return Jail
	}

	if isNotJail() {
		return NotJail
	}
	return Jail
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

func parseArgs(opts *options) ([]string, error) {
	p := initParser(opts)

	args, err := p.Parse()
	if err != nil {
		return nil, err
	}

	if opts.Version {
		showVersion()
		osExit(0)
	}

	return args, nil
}

func initParser(opts *options) *flags.Parser {
	parser := flags.NewParser(opts, flags.Default)
	parser.Name = cmdName
	parser.Usage = "[OPTION]"

	return parser
}

func showVersion() {
	description := cmdName + " version " + version + " (under Apache License verison 2.0)\n"
	fmt.Print(description)
}

func showHelp(p *flags.Parser) {
	p.WriteHelp(os.Stdout)
}
