//
// mimixbox/internal/applets/shellutils/mbsh/builtin/cd.go
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
package builtin

import (
	"errors"
	"os"
)

// ErrNoPath is returned when 'cd' was called without a second argument.
var ErrNoPath = errors.New("path required")

func cd(args []string) error {
	// 'cd' to home with empty path not yet supported.
	if len(args) < 2 {
		return ErrNoPath
	}
	// Change the directory and return the error.
	return os.Chdir(args[1])
}
