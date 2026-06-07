// mimixbox/internal/applets/shellutils/mbsh/builtin/buitInCmds.go
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

	"github.com/nao1215/mimixbox/internal/command"
)

// run is the signature of a built-in command. It receives the injected I/O
// streams so a built-in can read from and write to the same streams the shell
// is driven by, which keeps the shell testable in memory.
type run func(stdio command.IO, args []string) error

// builtin maps a command name to its implementation.
type builtin map[string]run

var buitInCmds builtin

func init() {
	buitInCmds = builtin{
		"cd": cd,
	}
}

// ErrNotBuiltinCmd reports that the string is not a built-in command.
var ErrNotBuiltinCmd = errors.New("the command is not built-in")

// IsBuiltinCmd reports whether command is one of the shell's built-ins. It
// returns false for an empty string or a name that is not registered.
func IsBuiltinCmd(command string) bool {
	_, ok := buitInCmds[command]
	return ok
}

// Run executes the built-in named key, wiring it to stdio. It returns
// ErrNotBuiltinCmd when key is not a built-in.
func Run(stdio command.IO, key string, args []string) error {
	cmd, ok := buitInCmds[key]
	if !ok {
		return ErrNotBuiltinCmd
	}
	return cmd(stdio, args)
}
