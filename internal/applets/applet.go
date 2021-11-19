//
// mimixbox/internal/applets/applet.go
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
package applets

import (
	"fmt"
	"go/doc"
	"mimixbox/internal/applets/fileutils/mkdir"
	"mimixbox/internal/applets/fileutils/mkfifo"
	"mimixbox/internal/applets/fileutils/mv"
	"mimixbox/internal/applets/fileutils/rm"
	"mimixbox/internal/applets/fileutils/rmdir"
	"mimixbox/internal/applets/fileutils/touch"
	"mimixbox/internal/applets/jokeutils/fakemovie"
	"mimixbox/internal/applets/shellutils/base64"
	"mimixbox/internal/applets/shellutils/basename"
	"mimixbox/internal/applets/shellutils/chroot"
	"mimixbox/internal/applets/shellutils/echo"
	"mimixbox/internal/applets/shellutils/false"
	"mimixbox/internal/applets/shellutils/ghrdc"
	"mimixbox/internal/applets/shellutils/ischroot"
	"mimixbox/internal/applets/shellutils/mbsh"
	"mimixbox/internal/applets/shellutils/path"
	"mimixbox/internal/applets/shellutils/serial"
	"mimixbox/internal/applets/shellutils/sleep"
	"mimixbox/internal/applets/shellutils/true"
	"mimixbox/internal/applets/shellutils/which"
	"mimixbox/internal/applets/textutils/cat"
	"mimixbox/internal/applets/textutils/tac"
	"os"
	"sort"
	"strconv"
)

type EntryPoint func() (int, error)

type Applet struct {
	Ep   EntryPoint
	Desc string
}

var Applets map[string]Applet

func init() {
	Applets = map[string]Applet{
		"base64":    {base64.Run, "Base64 encode/decode from FILR(or STDIN) to STDOUT"},
		"basename":  {basename.Run, "Print basename (PATH without\"/\") from file path"},
		"cat":       {cat.Run, "Concatenate files and print on the standard output"},
		"chroot":    {chroot.Run, "Run command or interactive shell with special root directory"},
		"echo":      {echo.Run, "Display a line of text"},
		"fakemovie": {fakemovie.Run, "Adds a video playback button to the image"},
		"false":     {false.Run, "Do nothing. Return unsuccess(1)"},
		"ghrdc":     {ghrdc.Run, "GitHub Relase Download Counter"},
		"ischroot":  {ischroot.Run, "Detect if running in a chroot"},
		"mbsh":      {mbsh.Run, "Mimix Box Shell"},
		"mkdir":     {mkdir.Run, "Make directories"},
		"mkfifo":    {mkfifo.Run, "Make FIFO (named pipe)"},
		"mv":        {mv.Run, "Rename SOURCE to DESTINATION, or move SOURCE(s) to DIRECTORY"},
		"path":      {path.Run, "Manipulate filename path"},
		"rm":        {rm.Run, "Remove file(s) or directory(s)"},
		"rmdir":     {rmdir.Run, "Remove directory"},
		"serial":    {serial.Run, "Rename the file to the name with a serial number"},
		"sh":        {mbsh.Run, "Mimix Box Shell"},
		"sleep":     {sleep.Run, "Pause for NUMBER seconds(minutes, hours, days)"},
		"tac":       {tac.Run, "Print the file contents from the end to the beginning"},
		"touch":     {touch.Run, "Update the access and modification times of each FILE to the current time"},
		"true":      {true.Run, "Do nothing. Return success(0)"},
		"which":     {which.Run, "Returns the file path which would be executed in the current environment"},
	}
}

func HasApplet(target string) bool {
	_, ok := Applets[target]
	return ok
}

func ListApplets() {
	format := "%" + strconv.Itoa(longestAppletLength()) + "s - %s\n"
	for _, key := range sortApplet() {
		fmt.Printf(format, key, Applets[key].Desc)
	}
}

func ShowAppletsBySpaceSeparated() {
	var app string
	for _, key := range sortApplet() {
		app += key
		app += " "
	}
	doc.ToText(os.Stdout, app, "", "", 60)
}

func sortApplet() []string {
	var keys []string
	for applet := range Applets {
		keys = append(keys, applet)
	}
	sort.Strings(keys)
	return keys
}

func longestAppletLength() int {
	var max int = 0
	for _, key := range sortApplet() {
		if max < len(key) {
			max = len(key)
		}
	}
	return max
}
