//
// mimixbox/internal/applets/debianutils/add-shell/add-shell.go
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

// Package addShell implements the add-shell applet: append shell names to
// /etc/shells if they are not already listed.
package addShell

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// shellsPath is the file add-shell operates on by default.
const shellsPath = "/etc/shells"

// Command is the add-shell applet.
type Command struct{}

// New returns an add-shell command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "add-shell" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Add shell name to /etc/shells" }

// Run executes add-shell.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "SHELLNAME...", stdio.Err).WithHelp(command.Help{
		Description: "Add each SHELLNAME to /etc/shells, appending only the names that are\n" +
			"not already listed. The file is created if it does not yet exist.",
		Examples: []command.Example{
			{Command: "add-shell /bin/zsh", Explain: "add /bin/zsh to /etc/shells if absent"},
			{Command: "add-shell /usr/bin/fish /bin/dash", Explain: "add several shells at once"},
		},
		ExitStatus: "0  success.\n1  no shell name was given, or /etc/shells could not be written.",
	})

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	shells := fs.Args()
	if len(shells) == 0 {
		_, _ = fmt.Fprintf(stdio.Err, "%s: shellname [shellname ...]\n", c.Name())
		return command.SilentFailure()
	}

	if err := addShells(shellsPath, shells); err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "%s: %v\n", c.Name(), err)
		return command.SilentFailure()
	}
	return nil
}

// addShells appends each shell in shells to the file at path, skipping any that
// are already present. It mirrors Debian add-shell: existing entries are left
// untouched and new ones are appended in order.
func addShells(path string, shells []string) error {
	existing, err := readShells(path)
	if err != nil {
		return err
	}
	present := make(map[string]bool, len(existing))
	for _, s := range existing {
		present[s] = true
	}

	var toAdd []string
	for _, s := range shells {
		if present[s] {
			continue
		}
		present[s] = true
		toAdd = append(toAdd, s)
	}
	if len(toAdd) == 0 {
		return nil
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644) //nolint:gosec // /etc/shells is world-readable
	if err != nil {
		return err
	}

	for _, s := range toAdd {
		if _, err := fmt.Fprintln(f, s); err != nil {
			_ = f.Close()
			return err
		}
	}
	return f.Close()
}

// readShells returns the non-empty, whitespace-trimmed lines of the file at
// path. A missing file is treated as an empty list, matching add-shell's
// tolerance for a not-yet-created /etc/shells.
func readShells(path string) ([]string, error) {
	f, err := os.Open(path) //nolint:gosec // operating on the named shells file is the point
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer func() { _ = f.Close() }()

	var lines []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		lines = append(lines, line)
	}
	return lines, sc.Err()
}
