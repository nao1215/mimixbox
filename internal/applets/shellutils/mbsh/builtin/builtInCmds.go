//
// mimixbox/internal/applets/shellutils/mbsh/builtin/buitInCmds.go
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
)

// ビルトインコマンドの書式
type run func(args []string) error

// ビルトインコマンドのマップ
type builtin map[string]run

var buitInCmds builtin

func init() {
	buitInCmds = builtin{
		"cd": cd,
	}
}

// 文字列がビルトインコマンドではない
var ErrNotBuiltinCmd = errors.New("the command is not built-in")

// 引数の文字列がビルトインコマンドかどうかを返す。
// 文字列がビルトインコマンドに含まれればtrue、
// 含まれないもしくはcommand == nilの場合はfalse
func IsBuiltinCmd(command string) bool {
	_, ok := buitInCmds[command]
	return ok
}

func Run(key string, args []string) error {
	cmd, ok := buitInCmds[key]
	if !ok {
		return ErrNotBuiltinCmd
	}
	return cmd(args)
}
