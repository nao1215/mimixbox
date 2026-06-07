// mimixbox/internal/applets/applet.go
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
package applets

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"

	gzipCmd "github.com/nao1215/mimixbox/internal/applets/archival/gzip"
	"github.com/nao1215/mimixbox/internal/applets/console-tools/clear"
	"github.com/nao1215/mimixbox/internal/applets/console-tools/reset"
	addShell "github.com/nao1215/mimixbox/internal/applets/debianutils/add-shell"
	"github.com/nao1215/mimixbox/internal/applets/debianutils/ischroot"
	removeShell "github.com/nao1215/mimixbox/internal/applets/debianutils/remove-shell"
	"github.com/nao1215/mimixbox/internal/applets/debianutils/which"
	"github.com/nao1215/mimixbox/internal/applets/fileutils/chgrp"
	"github.com/nao1215/mimixbox/internal/applets/fileutils/chown"
	"github.com/nao1215/mimixbox/internal/applets/fileutils/cp"
	"github.com/nao1215/mimixbox/internal/applets/fileutils/ln"
	"github.com/nao1215/mimixbox/internal/applets/fileutils/mkdir"
	"github.com/nao1215/mimixbox/internal/applets/fileutils/mkfifo"
	"github.com/nao1215/mimixbox/internal/applets/fileutils/mv"
	"github.com/nao1215/mimixbox/internal/applets/fileutils/rm"
	"github.com/nao1215/mimixbox/internal/applets/fileutils/rmdir"
	"github.com/nao1215/mimixbox/internal/applets/fileutils/touch"
	"github.com/nao1215/mimixbox/internal/applets/games/lifegame"
	"github.com/nao1215/mimixbox/internal/applets/jokeutils/cowsay"
	"github.com/nao1215/mimixbox/internal/applets/jokeutils/fakemovie"
	"github.com/nao1215/mimixbox/internal/applets/jokeutils/sl"

	//"github.com/nao1215/mimixbox/internal/applets/loginutils/chsh"
	validShell "github.com/nao1215/mimixbox/internal/applets/debianutils/valid-shell"
	"github.com/nao1215/mimixbox/internal/applets/pmutils/halt"
	"github.com/nao1215/mimixbox/internal/applets/shellutils/base64"
	"github.com/nao1215/mimixbox/internal/applets/shellutils/basename"
	"github.com/nao1215/mimixbox/internal/applets/shellutils/chroot"
	"github.com/nao1215/mimixbox/internal/applets/shellutils/dirname"
	"github.com/nao1215/mimixbox/internal/applets/shellutils/echo"
	"github.com/nao1215/mimixbox/internal/applets/shellutils/false"
	"github.com/nao1215/mimixbox/internal/applets/shellutils/ghrdc"
	"github.com/nao1215/mimixbox/internal/applets/shellutils/groups"
	"github.com/nao1215/mimixbox/internal/applets/shellutils/hostid"
	"github.com/nao1215/mimixbox/internal/applets/shellutils/id"
	"github.com/nao1215/mimixbox/internal/applets/shellutils/kill"
	"github.com/nao1215/mimixbox/internal/applets/shellutils/mbsh"
	"github.com/nao1215/mimixbox/internal/applets/shellutils/path"
	"github.com/nao1215/mimixbox/internal/applets/shellutils/printenv"
	"github.com/nao1215/mimixbox/internal/applets/shellutils/printf"
	"github.com/nao1215/mimixbox/internal/applets/shellutils/pwd"
	"github.com/nao1215/mimixbox/internal/applets/shellutils/sddf"
	"github.com/nao1215/mimixbox/internal/applets/shellutils/seq"
	"github.com/nao1215/mimixbox/internal/applets/shellutils/serial"
	"github.com/nao1215/mimixbox/internal/applets/shellutils/sleep"
	"github.com/nao1215/mimixbox/internal/applets/shellutils/sync"
	"github.com/nao1215/mimixbox/internal/applets/shellutils/true"
	"github.com/nao1215/mimixbox/internal/applets/shellutils/uuidgen"
	"github.com/nao1215/mimixbox/internal/applets/shellutils/wget"
	"github.com/nao1215/mimixbox/internal/applets/shellutils/whoami"
	"github.com/nao1215/mimixbox/internal/applets/textutils/cat"
	"github.com/nao1215/mimixbox/internal/applets/textutils/dos2unix"
	"github.com/nao1215/mimixbox/internal/applets/textutils/expand"
	"github.com/nao1215/mimixbox/internal/applets/textutils/head"
	"github.com/nao1215/mimixbox/internal/applets/textutils/md5sum"
	"github.com/nao1215/mimixbox/internal/applets/textutils/nl"
	"github.com/nao1215/mimixbox/internal/applets/textutils/sha1sum"
	"github.com/nao1215/mimixbox/internal/applets/textutils/sha256sum"
	"github.com/nao1215/mimixbox/internal/applets/textutils/sha512sum"
	"github.com/nao1215/mimixbox/internal/applets/textutils/tac"
	"github.com/nao1215/mimixbox/internal/applets/textutils/tail"

	"github.com/nao1215/mimixbox/internal/applets/textutils/tr"
	"github.com/nao1215/mimixbox/internal/applets/textutils/unexpand"
	"github.com/nao1215/mimixbox/internal/applets/textutils/unix2dos"
	"github.com/nao1215/mimixbox/internal/applets/textutils/wc"
)

type EntryPoint func() (int, error)

type Applet struct {
	Ep   EntryPoint
	Desc string
}

var Applets map[string]Applet

// reg builds an Applet entry for a command that has been migrated to the
// internal/command framework. The command's own Synopsis becomes the listed
// description, so the two never drift apart.
func reg(c command.Command) Applet {
	return Applet{Ep: command.Adapt(c), Desc: c.Synopsis()}
}

func init() {
	Applets = map[string]Applet{
		"add-shell": {addShell.Run, "Add shell name to /etc/shells"},
		"base64":    reg(base64.New()),
		"basename":  reg(basename.New()),
		"cat":       reg(cat.New()),
		"cowsay":    {cowsay.Run, "Print message with cow's ASCII art"},
		"chgrp":     {chgrp.Run, "Change the group of each FILE to GROUP"},
		"chown":     {chown.Run, "Change the owner and/or group of each FILE to OWNER and/or GROUP"},
		"chroot":    {chroot.Run, "Run command or interactive shell with special root directory"},
		//"chsh":         {chsh.Run, "Cqhange login shell"},
		"clear":        reg(clear.New()),
		"cp":           reg(cp.New()),
		"dirname":      reg(dirname.New()),
		"dos2unix":     reg(dos2unix.New()),
		"echo":         reg(echo.New()),
		"expand":       reg(expand.New()),
		"fakemovie":    {fakemovie.Run, "Adds a video playback button to the image"},
		"false":        reg(boolfalse.New()),
		"ghrdc":        {ghrdc.Run, "GitHub Relase Download Counter"},
		"groups":       reg(groups.New()),
		"gzip":         {gzipCmd.Run, "Compress or uncompress FILEs (by default, compress FILES in-place)"},
		"halt":         {halt.Run, "Halt the system"},
		"head":         reg(head.New()),
		"hostid":       reg(hostid.New()),
		"id":           reg(id.New()),
		"ischroot":     reg(ischroot.New()),
		"kill":         {kill.Run, "Kill process or send signal to process"},
		"lifegame":     {lifegame.Run, "Life game (Conway's Game of Life)"},
		"ln":           reg(ln.New()),
		"mbsh":         {mbsh.Run, "Mimix Box Shell"},
		"md5sum":       reg(md5sum.New()),
		"mkdir":        reg(mkdir.New()),
		"mkfifo":       reg(mkfifo.New()),
		"mv":           reg(mv.New()),
		"nl":           reg(nl.New()),
		"path":         reg(path.New()),
		"poweroff":     {halt.Run, "Power off the system"},
		"printenv":     reg(printenv.New()),
		"printf":       reg(printf.New()),
		"pwd":          reg(pwd.New()),
		"remove-shell": {removeShell.Run, "Remove shell name from /etc/shells"},
		"reboot":       {halt.Run, "Reboot the system"},
		"reset":        reg(reset.New()),
		"rm":           reg(rm.New()),
		"rmdir":        reg(rmdir.New()),
		"sddf":         {sddf.Run, "Search & Delete Duplicated File"},
		"serial":       reg(serial.New()),
		"sha1sum":      reg(sha1sum.New()),
		"sha256sum":    reg(sha256sum.New()),
		"sha512sum":    reg(sha512sum.New()),
		"seq":          reg(seq.New()),
		"sl":           {sl.Run, "Cure your bad habit of mistyping"},
		"sleep":        reg(sleep.New()),
		"sync":         reg(sync.New()),
		"tac":          reg(tac.New()),
		"tail":         reg(tail.New()),
		"touch":        reg(touch.New()),
		"tr":           reg(tr.New()),
		"true":         reg(booltrue.New()),
		"unexpand":     reg(unexpand.New()),
		"unix2dos":     reg(unix2dos.New()),
		"uuidgen":      reg(uuidgen.New()),
		"valid-shell":  {validShell.Run, "Verify if /etc/shells is valid"},
		"wc":           reg(wc.New()),
		"wget":         {wget.Run, "The non-interactive network downloader"},
		"which":        reg(which.New()),
		"whoami":       reg(whoami.New()),
	}
}

func HasApplet(target string) bool {
	_, ok := Applets[target]
	return ok
}

func ListApplets() {
	format := "%" + strconv.Itoa(longestAppletLength()) + "s - %s\n"
	for _, key := range SortApplet() {
		fmt.Fprintf(os.Stdout, format, key, Applets[key].Desc)
	}
}

func ShowAppletsBySpaceSeparated() {
	const wrap = 60
	var b strings.Builder
	lineLen := 0
	for i, key := range SortApplet() {
		if i > 0 {
			if lineLen+1+len(key) > wrap {
				b.WriteByte('\n')
				lineLen = 0
			} else {
				b.WriteByte(' ')
				lineLen++
			}
		}
		b.WriteString(key)
		lineLen += len(key)
	}
	b.WriteByte('\n')
	fmt.Fprint(os.Stdout, b.String())
}

func SortApplet() []string {
	var keys []string
	for applet := range Applets {
		keys = append(keys, applet)
	}
	sort.Strings(keys)
	return keys
}

func longestAppletLength() int {
	var max int = 0
	for _, key := range SortApplet() {
		if max < len(key) {
			max = len(key)
		}
	}
	return max
}
