//
// mimixbox/internal/applets/debianutils/valid-shell/valid-shell.go
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

// Package validShell implements the valid-shell applet: verify that every
// shell listed in /etc/shells exists and is executable.
package validShell

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// shellsPath is the file valid-shell checks by default.
const shellsPath = "/etc/shells"

// Command is the valid-shell applet.
type Command struct{}

// New returns a valid-shell command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "valid-shell" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Verify if /etc/shells is valid" }

// Run executes valid-shell.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[FILE]", stdio.Err)

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	path := shellsPath
	if rest := fs.Args(); len(rest) > 0 {
		path = rest[0]
	}

	ok, err := validateShells(path, stdio.Out)
	if err != nil {
		fmt.Fprintf(stdio.Err, "%s: %v\n", c.Name(), err)
		return command.SilentFailure()
	}
	if !ok {
		return command.SilentFailure()
	}
	return nil
}

// validateShells reads the shells file at path and writes an "OK:"/"NG:" line
// to out for each listed shell. Comment lines (starting with '#') and blank
// lines are ignored. It returns ok=true only when every listed shell exists and
// is executable.
func validateShells(path string, out io.Writer) (bool, error) {
	f, err := os.Open(path) //nolint:gosec // operating on the named shells file is the point
	if err != nil {
		return false, err
	}
	defer f.Close()

	ok := true
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if isExecutable(line) {
			fmt.Fprintf(out, "OK: %s\n", line)
		} else {
			fmt.Fprintf(out, "NG: %s (not exist in the system)\n", line)
			ok = false
		}
	}
	if err := sc.Err(); err != nil {
		return false, err
	}
	return ok, nil
}

// isExecutable reports whether path is an existing regular file with an execute
// bit set.
func isExecutable(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	if info.IsDir() {
		return false
	}
	return info.Mode().Perm()&0111 != 0
}
