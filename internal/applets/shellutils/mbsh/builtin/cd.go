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
	"os"

	"github.com/nao1215/mimixbox/internal/command"
)

// ErrNoPath is returned when 'cd' was called without a second argument.
var ErrNoPath = errors.New("path required")

// cd changes the process working directory. The stdio streams are accepted to
// satisfy the built-in signature; cd produces no output of its own. args holds
// the operands after the command name, so the target path is args[0].
func cd(_ command.IO, args []string) error {
	// 'cd' to home with an empty path is not yet supported.
	if len(args) < 1 {
		return ErrNoPath
	}
	// Change the directory and return the error.
	return os.Chdir(args[0])
}
