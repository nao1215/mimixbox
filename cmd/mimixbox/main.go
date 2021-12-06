//
// mimixbox/cmd/mimixbox/main.go
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
package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	mb "github.com/nao1215/mimixbox/internal/lib"

	"github.com/nao1215/mimixbox/internal/applets"

	"github.com/jessevdk/go-flags"
)

const cmdName string = "mimixbox"

// TODO: Change go-flags library to another one.
// go-flags is not suitable for parsing "mimixbox options" and "applet options"
// at the same time. There are other problems.
// If the option type is string, an unnecessary "=" will be included in the description
// of the long option. It's like this.
//
// Application Options:
//  -i, --install=      Create symbolic links for commands that don't exist on the system.
//  -f, --full-install= Create symbolic links regardless of system state.
//
type options struct {
	Install     bool `short:"i" long:"install" description:"Create symbolic links for commands that don't exist on the system."`
	FullInstall bool `short:"f" long:"full-install" description:"Create symbolic links regardless of system state."`
	List        bool `short:"l" long:"list" description:"Show command name provided by mimixbox"`
	Remove      bool `short:"r" long:"remove" description:"Remove symbolic links for commands provided by mimixbox."`
	Version     bool `short:"v" long:"version" description:"Show mimixbox command version"`
}

var osExit = os.Exit

const version = "0.27.20"

const (
	ExitSuccess int = iota // 0
	ExitFailuer
)

func main() {
	var opts options
	var status int
	var err error
	parser := initParser(&opts)

	// The contents of os.Args [0] are different when mimixbox is
	// executed directly and when it is executed via a symbolic link.
	// [e.g.]
	// $ mimixbox cat
	//   --> os.Args[0] = mimixbox
	// $ cat                        ※ This is symbolic link for mimixbox.
	//                                 cat --->   /bin/mimixbox
	//   --> os.Args[0] = cat
	if strings.Contains(os.Args[0], cmdName) {
		handleMimixBoxOptionsIfNeeded(parser, &opts)
		os.Args = os.Args[1:]
	}

	// If the specified command(applet) is not built in mimixbox.
	if !hasAppletName() {
		fmt.Fprintf(os.Stderr, "%s is not provided by mimixbox.\n\n", os.Args[0])
		fmt.Fprintln(os.Stderr, "[Commands supported by MimixBox]")
		applets.ShowAppletsBySpaceSeparated()
		osExit(ExitFailuer)
	}

	applet := path.Base(os.Args[0])
	app := applets.Applets[applet]
	if status, err = app.Ep(); err != nil {
		fmt.Fprintln(os.Stderr, applet+": "+err.Error())
		osExit(status)
	}
	osExit(status)
}

// If the mimixbox option exists, execute the processing for the option and exit.
// TODO: Rewrite this function. This method is too complicated
func handleMimixBoxOptionsIfNeeded(parser *flags.Parser, opts *options) {
	mimixBoxPath := os.Args[0]

	// As a temporary workaround for the bug, if the Applet name is included in the argument,
	// it is considered not to be an argument of mimixbox.
	// The directory with the same name as an applet can no longer be an installation directory.
	if hasAppletName() {
		return
	}

	// If user specify help option for applet command.
	if hasHelpOption() && hasAppletName() {
		return
	}

	// Only mimixbox. no option and no argument.
	if len(os.Args) == 1 {
		showHelp(parser)
		osExit(ExitFailuer)
	}

	args, err := parser.Parse()
	if err != nil {
		osExit(ExitFailuer)
	}

	if len(args) == 0 && opts.Version {
		mb.ShowVersion(cmdName, version)
		osExit(ExitSuccess)
	}

	if len(args) == 0 && opts.List {
		applets.ListApplets()
		osExit(ExitSuccess)
	}

	if len(args) == 1 && opts.Install {
		if err = minimumInstall(mimixBoxPath, args[0]); err != nil {
			fmt.Fprintln(os.Stderr, err)
			osExit(ExitFailuer)
		}
		osExit(ExitSuccess)
	}

	if len(args) == 1 && opts.Remove {
		if err = remove(args[0]); err != nil {
			fmt.Fprintln(os.Stderr, err)
			osExit(ExitFailuer)
		}
		osExit(ExitSuccess)
	}

	if len(args) == 1 && opts.FullInstall {
		if err = fullInstall(mimixBoxPath, args[0]); err != nil {
			fmt.Fprintln(os.Stderr, err)
			osExit(ExitFailuer)
		}
		osExit(ExitSuccess)
	}

	if len(args) == 0 && (opts.FullInstall || opts.Install || opts.Remove) {
		showHelp(parser)
		osExit(ExitFailuer)
	}
}

// If go-flags find the help option while parsing the option,
// go-flags show help message immediately.
// The help option needs to determine whether it is specified for mimixbox or not.
// The situation where a hack to the library is needed is not desirable.
func hasHelpOption() bool {
	for _, s := range os.Args[1:] {
		if s == "--help" || s == "-h" || s == "--full-install" || s == "-f" {
			return true
		}
	}
	return false
}

func hasAppletName() bool {
	for _, app := range os.Args {
		if applets.HasApplet(path.Base(app)) {
			return true
		}
	}
	return false
}

func initParser(opts *options) *flags.Parser {
	parser := flags.NewParser(opts, flags.Default)
	parser.Name = cmdName
	parser.Usage = "[applet [arguments]...] [OPTIONS]"

	return parser
}

func showHelp(p *flags.Parser) {
	p.WriteHelp(os.Stdout)
}

func minimumInstall(mimixboxPath string, installPath string) error {
	return __install(mimixboxPath, installPath, false)
}

func fullInstall(mimixboxPath string, installPath string) error {
	return __install(mimixboxPath, installPath, true)
}

func __install(mimixboxPath string, installPath string, full bool) error {
	targetPath := os.ExpandEnv(installPath)
	if !mb.IsDir(targetPath) {
		return errors.New(targetPath + ": no such directory")
	}

	mimixboxAbsPath, err := getMimixBoxAbsPath(targetPath)
	if err != nil {
		return err
	}

	for applet := range applets.Applets {
		if !full && mb.ExistCmd(applet) {
			fmt.Fprintf(os.Stderr, "Same name command(%s) already exists. Not create symbolic link.\n", applet)
			continue // if same name command already exists, not install for safety.
		}

		// If a symbolic link with the same name already exists,
		// delete that link. If the binary has the same name,
		// do not delete it. The former is likely to have been
		// created by mimixbox, while the latter may be binaries
		// provided by other packages.
		newPath := filepath.Join(targetPath, applet)
		if mb.IsSymlink(newPath) {
			err := os.Remove(newPath) // Remove  even BusyBox's symbolic link
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				continue
			}
			fmt.Fprintf(os.Stdout, "Delete              : %s\n", newPath)
		}

		err = os.Symlink(mimixboxAbsPath, newPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			continue
		}

		fmt.Fprintf(os.Stdout, "Create symbolic link: %s\n", newPath)
	}
	return nil
}

func getMimixBoxAbsPath(mimixboxPath string) (string, error) {
	// If mimixbox is installed on system (mimixbox in $PATH)
	path, err := exec.LookPath(cmdName)
	if err == nil {
		return path, nil
	}

	path, err = filepath.Abs(os.Args[0])
	if err != nil {
		return "", err
	}
	return path, nil
}

func remove(installPath string) error {
	targetPath := os.ExpandEnv(installPath)

	if !mb.IsDir(targetPath) {
		return errors.New(targetPath + ": no such directory")
	}

	for name := range applets.Applets {
		symbolicPath := filepath.Join(targetPath, name)

		if !mb.IsSymlink(symbolicPath) {
			continue
		}

		realPath, err := os.Readlink(symbolicPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			continue
		}

		if strings.Contains(realPath, cmdName) {
			err := os.Remove(symbolicPath)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				continue
			}
		}
		fmt.Fprintf(os.Stdout, "Delete symbolic link: %s\n", symbolicPath)
	}
	return nil
}
