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
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"syscall"

	"golang.org/x/term"
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

func Parrot(withNl bool) {
	var response string
	var nl int = 1
	for {
		response = ""
		_, err := fmt.Scanln(&response)
		if err != nil {
			if !strings.Contains(err.Error(), "expected newline") {
				break // Ctrl+D or other error.
			}
		}
		if withNl {
			PrintStrWithNumberLine(nl, response)
		} else {
			fmt.Println(response)
		}
	}
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

func WrapString(src string, column int) string {
	var buf []string

	if column <= 0 {
		return src
	}

	for i := 0; i < len(src); i += column {
		if i+column < len(src) {
			buf = append(buf, src[i:(i+column)])
		} else {
			buf = append(buf, src[i:])
		}
	}
	return strings.Join(buf, "\n")
}

func Concatenate(path []string, lfAtTheJoint bool) ([]string, error) {
	var strList []string
	var index int

	for _, file := range path {
		list, err := ReadFileToStrList(file)
		if err != nil {
			return nil, err
		}

		// In the case of the cat command, the beginning of the new file is
		// concatenated to the end of the previous file.
		// In the case of nl command, do not concatenate the new file at the end
		// of the previous file. The end of file (EOF) is replaced with a newline.
		index = len(strList) - 1
		if lfAtTheJoint { // for nl command
			list[len(list)-1] = list[len(list)-1] + "\n"
			strList = append(strList, list...)
		} else { // for cat command
			if index > 0 {
				strList[index] = strList[index] + list[0]
				strList = append(strList, list[1:]...)
			} else {
				strList = append(strList, list...)
			}
		}
	}
	return strList, nil
}

func PrintStrListWithNumberLine(strList []string, countEmpryLine bool) {
	var nl int = 1
	for _, s := range strList {
		if s == "\n" && !countEmpryLine {
			fmt.Print(s)
			continue
		}
		PrintStrWithNumberLine(nl, s)
		nl++
	}
}

func PrintStrWithNumberLine(nl int, str string) {
	fmt.Printf("%6d  %s", nl, str)
}

func FromPIPE() (string, error) {
	if HasPipeData() {
		b, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			return "", err
		}
		return string(b), nil
	}
	return "", nil
}

func HasPipeData() bool {
	return !term.IsTerminal(syscall.Stdin)
}

func ChopAll(lines []string) []string {
	var newLines []string
	for _, v := range lines {
		newLines = append(newLines, Chop(v))
	}
	return newLines
}

func Chop(line string) string {
	if strings.HasSuffix(line, "\n") {
		return strings.TrimRight(line, "\n")
	}
	return line
}

func Dump(lines []string, withNumber bool) {
	if withNumber {
		PrintStrListWithNumberLine(lines, true)
	} else {
		for _, line := range lines {
			fmt.Print(line)
		}
	}
}

func Groups() ([]user.Group, error) {
	u, err := user.Current()
	if err != nil {
		return nil, err
	}

	groups, err := u.GroupIds()
	if err != nil {
		return nil, err
	}

	var groupList []user.Group
	for _, v := range groups {
		group, err := user.LookupGroupId(v)
		if err != nil {
			return nil, err
		}
		groupList = append(groupList, *group)
	}
	return groupList, nil
}
