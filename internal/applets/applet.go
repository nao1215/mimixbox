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
	"os"
	"sort"
	"strconv"

	"github.com/nao1215/mimixbox/internal/applets/fileutils/cp"
	"github.com/nao1215/mimixbox/internal/applets/fileutils/ln"
	"github.com/nao1215/mimixbox/internal/applets/fileutils/mkdir"
	"github.com/nao1215/mimixbox/internal/applets/fileutils/mkfifo"
	"github.com/nao1215/mimixbox/internal/applets/fileutils/mv"
	"github.com/nao1215/mimixbox/internal/applets/fileutils/rm"
	"github.com/nao1215/mimixbox/internal/applets/fileutils/rmdir"
	"github.com/nao1215/mimixbox/internal/applets/fileutils/touch"
	"github.com/nao1215/mimixbox/internal/applets/jokeutils/cowsay"
	"github.com/nao1215/mimixbox/internal/applets/jokeutils/fakemovie"
	"github.com/nao1215/mimixbox/internal/applets/shellutils/base64"
	"github.com/nao1215/mimixbox/internal/applets/shellutils/basename"
	"github.com/nao1215/mimixbox/internal/applets/shellutils/chroot"
	"github.com/nao1215/mimixbox/internal/applets/shellutils/echo"
	"github.com/nao1215/mimixbox/internal/applets/shellutils/false"
	"github.com/nao1215/mimixbox/internal/applets/shellutils/ghrdc"
	"github.com/nao1215/mimixbox/internal/applets/shellutils/ischroot"
	"github.com/nao1215/mimixbox/internal/applets/shellutils/mbsh"
	"github.com/nao1215/mimixbox/internal/applets/shellutils/path"
	"github.com/nao1215/mimixbox/internal/applets/shellutils/serial"
	"github.com/nao1215/mimixbox/internal/applets/shellutils/sleep"
	"github.com/nao1215/mimixbox/internal/applets/shellutils/true"
	"github.com/nao1215/mimixbox/internal/applets/shellutils/which"
	"github.com/nao1215/mimixbox/internal/applets/shellutils/whoami"
	"github.com/nao1215/mimixbox/internal/applets/textutils/cat"
	"github.com/nao1215/mimixbox/internal/applets/textutils/dos2unix"
	"github.com/nao1215/mimixbox/internal/applets/textutils/expand"
	"github.com/nao1215/mimixbox/internal/applets/textutils/head"
	"github.com/nao1215/mimixbox/internal/applets/textutils/nl"
	"github.com/nao1215/mimixbox/internal/applets/textutils/tac"
	"github.com/nao1215/mimixbox/internal/applets/textutils/tail"
	"github.com/nao1215/mimixbox/internal/applets/textutils/unexpand"
	"github.com/nao1215/mimixbox/internal/applets/textutils/unix2dos"
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
		"cowsay":    {cowsay.Run, "Print message with cow's ASCII art"},
		"chroot":    {chroot.Run, "Run command or interactive shell with special root directory"},
		"cp":        {cp.Run, "Copy file(s) otr Directory(s)"},
		"dos2unix":  {dos2unix.Run, "Change CRLF to LF"},
		"echo":      {echo.Run, "Display a line of text"},
		"expand":    {expand.Run, "Convert TAB to N space (default:N=8)"},
		"fakemovie": {fakemovie.Run, "Adds a video playback button to the image"},
		"false":     {false.Run, "Do nothing. Return unsuccess(1)"},
		"ghrdc":     {ghrdc.Run, "GitHub Relase Download Counter"},
		"head":      {head.Run, "Print the first NUMBER(default=10) lines"},
		"ischroot":  {ischroot.Run, "Detect if running in a chroot"},
		"ln":        {ln.Run, "Create hard or symbolic link"},
		"mbsh":      {mbsh.Run, "Mimix Box Shell"},
		"mkdir":     {mkdir.Run, "Make directories"},
		"mkfifo":    {mkfifo.Run, "Make FIFO (named pipe)"},
		"mv":        {mv.Run, "Rename SOURCE to DESTINATION, or move SOURCE(s) to DIRECTORY"},
		"nl":        {nl.Run, "Write each FILE to standard output with line numbers added"},
		"path":      {path.Run, "Manipulate filename path"},
		"rm":        {rm.Run, "Remove file(s) or directory(s)"},
		"rmdir":     {rmdir.Run, "Remove directory"},
		"serial":    {serial.Run, "Rename the file to the name with a serial number"},
		"sh":        {mbsh.Run, "Mimix Box Shell"},
		"sleep":     {sleep.Run, "Pause for NUMBER seconds(minutes, hours, days)"},
		"tac":       {tac.Run, "Print the file contents from the end to the beginning"},
		"tail":      {tail.Run, "Print the last NUMBER(default=10) lines"},
		"touch":     {touch.Run, "Update the access and modification times of each FILE to the current time"},
		"true":      {true.Run, "Do nothing. Return success(0)"},
		"unexpand":  {unexpand.Run, "Convert N space to TAB(default:N=8)"},
		"unix2dos":  {unix2dos.Run, "Change LF to CRLF"},
		"which":     {which.Run, "Returns the file path which would be executed in the current environment"},
		"whoami":    {whoami.Run, "Print login user name"},
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
