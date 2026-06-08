// mimixbox/internal/applets/shellutils/mbsh/builtin/cd.go
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
package builtin

import (
	"errors"
	"fmt"
	"os"

	"github.com/nao1215/mimixbox/internal/command"
)

// ErrNoHome is returned when 'cd' with no argument cannot find $HOME.
var ErrNoHome = errors.New("HOME not set")

// ErrNoOldpwd is returned when 'cd -' has no previous directory to return to.
var ErrNoOldpwd = errors.New("OLDPWD not set")

// cd changes the working directory and maintains the PWD/OLDPWD environment
// variables. With no argument it changes to $HOME; "cd -" returns to the
// previous directory (and prints it, like a POSIX shell).
func cd(stdio command.IO, args []string) error {
	target, printDir, err := cdTarget(args)
	if err != nil {
		return err
	}

	old, _ := os.Getwd()
	if err := os.Chdir(target); err != nil {
		return err
	}

	newDir, err := os.Getwd()
	if err != nil {
		return err
	}
	_ = os.Setenv("OLDPWD", old)
	_ = os.Setenv("PWD", newDir)

	if printDir {
		_, _ = fmt.Fprintln(stdio.Out, newDir)
	}
	return nil
}

// cdTarget resolves the directory cd should move to and whether the new
// directory should be echoed (as "cd -" does).
func cdTarget(args []string) (target string, printDir bool, err error) {
	if len(args) < 1 || args[0] == "" {
		home := os.Getenv("HOME")
		if home == "" {
			return "", false, ErrNoHome
		}
		return home, false, nil
	}
	if args[0] == "-" {
		old := os.Getenv("OLDPWD")
		if old == "" {
			return "", false, ErrNoOldpwd
		}
		return old, true, nil
	}
	return args[0], false, nil
}
