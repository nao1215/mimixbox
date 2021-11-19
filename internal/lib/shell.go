//
// mimixbox/internal/lib/shell.go
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
package mb

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func ExistCmd(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

func IsRootUser() bool {
	return os.Geteuid() == 0
}

func IsRootDir(path string) bool {
	p, err := filepath.Abs(path)
	if err != nil {
		return false
	}
	return p == "/"
}

func Question(ask string) bool {
	var response string

	fmt.Printf(ask + " [Y/n] ")
	_, err := fmt.Scanln(&response)
	if err != nil {
		// If user input only enter.
		if strings.Contains(err.Error(), "expected newline") {
			return Question(ask)
		}
		fmt.Print(err.Error())
		return false
	}

	switch strings.ToLower(response) {
	case "y", "yes":
		return true
	case "n", "no":
		return false
	default:
		return Question(ask)
	}
}

func Parrot() bool {
	var response string

	_, err := fmt.Scanln(&response)
	if err != nil {
		if !strings.Contains(err.Error(), "expected newline") {
			return false // Ctrl+D or other error.
		}
	}
	fmt.Println(response)
	return true
}

func Input() (string, bool) {
	var response string

	_, err := fmt.Scanln(&response)
	if err != nil {
		if !strings.Contains(err.Error(), "expected newline") {
			return "", false // Ctrl+D or other error.
		}
	}
	return response, true
}
