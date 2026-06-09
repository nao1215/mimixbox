// mimixbox/internal/lib/shell.go
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
package mb

import (
	"fmt"
	"io"
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

func IsRootDir(path string) bool {
	p, err := filepath.Abs(path)
	if err != nil {
		return false
	}
	return p == "/"
}

// QuestionFrom prompts ask on out and reads a yes/no answer from in. It returns
// true for "y"/"yes", false for "n"/"no", re-prompts on a blank or invalid line,
// and returns false on EOF or a read error. Injecting in/out keeps it
// unit-testable with a strings.Reader and a bytes.Buffer.
func QuestionFrom(in io.Reader, out io.Writer, ask string) bool {
	for {
		var response string
		fmt.Fprintf(out, ask+" [Y/n] ")
		_, err := fmt.Fscanln(in, &response)
		if err != nil {
			// An empty line (just Enter) reports "unexpected newline"; re-ask.
			if strings.Contains(err.Error(), "expected newline") {
				continue
			}
			return false
		}

		switch strings.ToLower(response) {
		case "y", "yes":
			return true
		case "n", "no":
			return false
		default:
			continue
		}
	}
}

// Question is QuestionFrom wired to the process stdin/stdout.
func Question(ask string) bool { return QuestionFrom(os.Stdin, os.Stdout, ask) }

// ParrotFrom echoes each line read from in to out, optionally with a line
// number, until in is exhausted. Injecting in/out keeps it testable.
func ParrotFrom(in io.Reader, out io.Writer, withNl bool) {
	nl := 1
	for {
		var response string
		_, err := fmt.Fscanln(in, &response)
		if err != nil {
			if !strings.Contains(err.Error(), "expected newline") {
				break // Ctrl+D or other error.
			}
		}
		if withNl {
			PrintStrWithNumberLineTo(out, nl, "  %6d  %s", response) // respect Coreutils
			nl++
		} else {
			fmt.Fprintln(out, response)
		}
	}
}

// Parrot is ParrotFrom wired to the process stdin/stdout.
func Parrot(withNl bool) { ParrotFrom(os.Stdin, os.Stdout, withNl) }

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

func Concatenate(path []string) ([]string, error) {
	var strList []string
	var index int

	for _, file := range path {
		list, err := ReadFileToStrList(os.ExpandEnv(file))
		if err != nil {
			return nil, err
		}

		index = len(strList) - 1
		if index > 0 && !strings.HasSuffix(strList[index], "\n") {
			strList[index] = strList[index] + list[0]
			strList = append(strList, list[1:]...)
		} else {
			strList = append(strList, list...)
		}
	}
	return strList, nil
}

// PrintStrListWithNumberLineTo writes the numbered lines to w.
func PrintStrListWithNumberLineTo(w io.Writer, strList []string, countEmpryLine bool) {
	nl := 1
	for _, s := range strList {
		if s == "\n" && !countEmpryLine {
			fmt.Fprint(w, s)
			continue
		}
		PrintStrWithNumberLineTo(w, nl, "%6d  %s", s)
		nl++
	}
}

func PrintStrListWithNumberLine(strList []string, countEmpryLine bool) {
	PrintStrListWithNumberLineTo(os.Stdout, strList, countEmpryLine)
}

// PrintStrWithNumberLineTo writes one numbered line to w.
func PrintStrWithNumberLineTo(w io.Writer, nl int, format string, message string) {
	fmt.Fprintf(w, format, nl, message)
}

func PrintStrWithNumberLine(nl int, format string, message string) {
	PrintStrWithNumberLineTo(os.Stdout, nl, format, message)
}

func FromPIPE() (string, error) {
	if HasPipeData() {
		b, err := io.ReadAll(os.Stdin)
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
		return strings.TrimSuffix(line, "\n")
	}
	return line
}

// DumpTo writes lines to w, optionally numbered.
func DumpTo(w io.Writer, lines []string, withNumber bool) {
	if withNumber {
		PrintStrListWithNumberLineTo(w, lines, true)
	} else {
		for _, line := range lines {
			fmt.Fprint(w, line)
		}
	}
}

func Dump(lines []string, withNumber bool) { DumpTo(os.Stdout, lines, withNumber) }

func Groups(uname string) ([]user.Group, error) {
	u, err := user.Lookup(uname)
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

// DumpGroupsTo writes the space-separated group names or GIDs to w.
func DumpGroupsTo(w io.Writer, groups []user.Group, showName bool) {
	resultLine := ""
	if showName {
		for _, g := range groups {
			resultLine = resultLine + g.Name + " "
		}

	} else {
		for _, g := range groups {
			resultLine = resultLine + g.Gid + " "
		}
	}
	fmt.Fprintln(w, strings.TrimSuffix(resultLine, " "))
}

func DumpGroups(groups []user.Group, showName bool) {
	DumpGroupsTo(os.Stdout, groups, showName)
}

func HasOperand(args []string, cmdName string) bool {
	for _, v := range args {
		if strings.HasPrefix(v, "-") && len(v) == 2 {
			continue
		}
		if strings.HasPrefix(v, "--") {
			continue
		}
		if v == cmdName && args[0] == cmdName {
			continue
		}
		return true
	}
	return false
}

func HasNoOperand(args []string, cmdName string) bool {
	return !HasOperand(args, cmdName)
}
