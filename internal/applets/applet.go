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

	"github.com/nao1215/mimixbox/internal/applets/archival/ar"
	"github.com/nao1215/mimixbox/internal/applets/archival/bunzip2"
	compressCmd "github.com/nao1215/mimixbox/internal/applets/archival/compress"
	"github.com/nao1215/mimixbox/internal/applets/archival/cpio"
	"github.com/nao1215/mimixbox/internal/applets/archival/gunzip"
	gzipCmd "github.com/nao1215/mimixbox/internal/applets/archival/gzip"
	"github.com/nao1215/mimixbox/internal/applets/archival/rpm"
	"github.com/nao1215/mimixbox/internal/applets/archival/rpm2cpio"
	tarCmd "github.com/nao1215/mimixbox/internal/applets/archival/tar"
	"github.com/nao1215/mimixbox/internal/applets/archival/uncompress"
	"github.com/nao1215/mimixbox/internal/applets/archival/unzip"
	zipCmd "github.com/nao1215/mimixbox/internal/applets/archival/zip"
	"github.com/nao1215/mimixbox/internal/applets/console-tools/clear"
	"github.com/nao1215/mimixbox/internal/applets/console-tools/reset"
	"github.com/nao1215/mimixbox/internal/applets/console-tools/resize"
	addShell "github.com/nao1215/mimixbox/internal/applets/debianutils/add-shell"
	"github.com/nao1215/mimixbox/internal/applets/debianutils/ischroot"
	"github.com/nao1215/mimixbox/internal/applets/debianutils/mktemp"
	removeShell "github.com/nao1215/mimixbox/internal/applets/debianutils/remove-shell"
	"github.com/nao1215/mimixbox/internal/applets/debianutils/which"
	"github.com/nao1215/mimixbox/internal/applets/editors/awk"
	"github.com/nao1215/mimixbox/internal/applets/editors/diff"
	"github.com/nao1215/mimixbox/internal/applets/editors/patch"
	"github.com/nao1215/mimixbox/internal/applets/editors/sed"
	"github.com/nao1215/mimixbox/internal/applets/editors/vi"
	"github.com/nao1215/mimixbox/internal/applets/fileutils/chgrp"
	"github.com/nao1215/mimixbox/internal/applets/fileutils/chown"
	"github.com/nao1215/mimixbox/internal/applets/fileutils/cp"
	"github.com/nao1215/mimixbox/internal/applets/fileutils/link"
	"github.com/nao1215/mimixbox/internal/applets/fileutils/ln"
	"github.com/nao1215/mimixbox/internal/applets/fileutils/mkdir"
	"github.com/nao1215/mimixbox/internal/applets/fileutils/mkfifo"
	"github.com/nao1215/mimixbox/internal/applets/fileutils/mountpoint"
	"github.com/nao1215/mimixbox/internal/applets/fileutils/mv"
	"github.com/nao1215/mimixbox/internal/applets/fileutils/readlink"
	"github.com/nao1215/mimixbox/internal/applets/fileutils/rm"
	"github.com/nao1215/mimixbox/internal/applets/fileutils/rmdir"
	"github.com/nao1215/mimixbox/internal/applets/fileutils/shred"
	statCmd "github.com/nao1215/mimixbox/internal/applets/fileutils/stat"
	"github.com/nao1215/mimixbox/internal/applets/fileutils/touch"
	"github.com/nao1215/mimixbox/internal/applets/fileutils/truncate"
	"github.com/nao1215/mimixbox/internal/applets/fileutils/unlink"
	"github.com/nao1215/mimixbox/internal/applets/findutils/find"
	"github.com/nao1215/mimixbox/internal/applets/findutils/grep"
	"github.com/nao1215/mimixbox/internal/applets/findutils/xargs"
	"github.com/nao1215/mimixbox/internal/applets/games/lifegame"
	"github.com/nao1215/mimixbox/internal/applets/jokeutils/cowsay"
	"github.com/nao1215/mimixbox/internal/applets/jokeutils/fakemovie"
	"github.com/nao1215/mimixbox/internal/applets/jokeutils/sl"

	//"github.com/nao1215/mimixbox/internal/applets/loginutils/chsh"
	validShell "github.com/nao1215/mimixbox/internal/applets/debianutils/valid-shell"
	"github.com/nao1215/mimixbox/internal/applets/pmutils/halt"
	"github.com/nao1215/mimixbox/internal/applets/shellutils/arch"
	"github.com/nao1215/mimixbox/internal/applets/shellutils/base64"
	"github.com/nao1215/mimixbox/internal/applets/shellutils/basename"
	"github.com/nao1215/mimixbox/internal/applets/shellutils/cal"
	"github.com/nao1215/mimixbox/internal/applets/shellutils/chmod"
	"github.com/nao1215/mimixbox/internal/applets/shellutils/chroot"
	"github.com/nao1215/mimixbox/internal/applets/shellutils/cmp"
	"github.com/nao1215/mimixbox/internal/applets/shellutils/cut"
	"github.com/nao1215/mimixbox/internal/applets/shellutils/date"
	"github.com/nao1215/mimixbox/internal/applets/shellutils/dd"
	"github.com/nao1215/mimixbox/internal/applets/shellutils/df"
	"github.com/nao1215/mimixbox/internal/applets/shellutils/dirname"
	"github.com/nao1215/mimixbox/internal/applets/shellutils/du"
	"github.com/nao1215/mimixbox/internal/applets/shellutils/echo"
	"github.com/nao1215/mimixbox/internal/applets/shellutils/env"
	"github.com/nao1215/mimixbox/internal/applets/shellutils/expr"
	"github.com/nao1215/mimixbox/internal/applets/shellutils/false"
	"github.com/nao1215/mimixbox/internal/applets/shellutils/free"
	"github.com/nao1215/mimixbox/internal/applets/shellutils/ghrdc"
	"github.com/nao1215/mimixbox/internal/applets/shellutils/groups"
	"github.com/nao1215/mimixbox/internal/applets/shellutils/hostid"
	"github.com/nao1215/mimixbox/internal/applets/shellutils/hostname"
	"github.com/nao1215/mimixbox/internal/applets/shellutils/id"
	"github.com/nao1215/mimixbox/internal/applets/shellutils/logname"
	"github.com/nao1215/mimixbox/internal/applets/shellutils/install"
	"github.com/nao1215/mimixbox/internal/applets/shellutils/kill"
	"github.com/nao1215/mimixbox/internal/applets/shellutils/killall"
	"github.com/nao1215/mimixbox/internal/applets/shellutils/mbsh"
	"github.com/nao1215/mimixbox/internal/applets/shellutils/mknod"
	"github.com/nao1215/mimixbox/internal/applets/shellutils/nohup"
	"github.com/nao1215/mimixbox/internal/applets/shellutils/nproc"
	"github.com/nao1215/mimixbox/internal/applets/shellutils/od"
	"github.com/nao1215/mimixbox/internal/applets/shellutils/path"
	"github.com/nao1215/mimixbox/internal/applets/shellutils/printenv"
	"github.com/nao1215/mimixbox/internal/applets/shellutils/printf"
	"github.com/nao1215/mimixbox/internal/applets/shellutils/pwd"
	"github.com/nao1215/mimixbox/internal/applets/shellutils/realpath"
	"github.com/nao1215/mimixbox/internal/applets/shellutils/sddf"
	"github.com/nao1215/mimixbox/internal/applets/shellutils/seq"
	"github.com/nao1215/mimixbox/internal/applets/shellutils/serial"
	"github.com/nao1215/mimixbox/internal/applets/shellutils/sleep"
	sortcmd "github.com/nao1215/mimixbox/internal/applets/shellutils/sort"
	"github.com/nao1215/mimixbox/internal/applets/shellutils/sync"
	"github.com/nao1215/mimixbox/internal/applets/shellutils/tee"
	testcmd "github.com/nao1215/mimixbox/internal/applets/shellutils/test"
	timeoutCmd "github.com/nao1215/mimixbox/internal/applets/shellutils/timeout"
	"github.com/nao1215/mimixbox/internal/applets/shellutils/true"
	"github.com/nao1215/mimixbox/internal/applets/shellutils/tty"
	"github.com/nao1215/mimixbox/internal/applets/shellutils/uname"
	"github.com/nao1215/mimixbox/internal/applets/shellutils/watch"
	"github.com/nao1215/mimixbox/internal/applets/shellutils/uniq"
	"github.com/nao1215/mimixbox/internal/applets/shellutils/uuidgen"
	"github.com/nao1215/mimixbox/internal/applets/shellutils/wget"
	"github.com/nao1215/mimixbox/internal/applets/shellutils/who"
	"github.com/nao1215/mimixbox/internal/applets/shellutils/whoami"
	"github.com/nao1215/mimixbox/internal/applets/shellutils/yes"
	"github.com/nao1215/mimixbox/internal/applets/textutils/base32"
	"github.com/nao1215/mimixbox/internal/applets/textutils/cat"
	"github.com/nao1215/mimixbox/internal/applets/textutils/cksum"
	"github.com/nao1215/mimixbox/internal/applets/textutils/comm"
	"github.com/nao1215/mimixbox/internal/applets/textutils/dos2unix"
	"github.com/nao1215/mimixbox/internal/applets/textutils/expand"
	fmtCmd "github.com/nao1215/mimixbox/internal/applets/textutils/fmt"
	"github.com/nao1215/mimixbox/internal/applets/textutils/fold"
	"github.com/nao1215/mimixbox/internal/applets/textutils/head"
	"github.com/nao1215/mimixbox/internal/applets/textutils/md5sum"
	"github.com/nao1215/mimixbox/internal/applets/textutils/nl"
	"github.com/nao1215/mimixbox/internal/applets/textutils/paste"
	"github.com/nao1215/mimixbox/internal/applets/textutils/rev"
	"github.com/nao1215/mimixbox/internal/applets/textutils/sha1sum"
	"github.com/nao1215/mimixbox/internal/applets/textutils/sha256sum"
	"github.com/nao1215/mimixbox/internal/applets/textutils/sha512sum"
	"github.com/nao1215/mimixbox/internal/applets/textutils/shuf"
	"github.com/nao1215/mimixbox/internal/applets/textutils/split"
	stringsCmd "github.com/nao1215/mimixbox/internal/applets/textutils/strings"
	"github.com/nao1215/mimixbox/internal/applets/textutils/tac"
	"github.com/nao1215/mimixbox/internal/applets/textutils/tail"

	"github.com/nao1215/mimixbox/internal/applets/textutils/tr"
	"github.com/nao1215/mimixbox/internal/applets/textutils/unexpand"
	"github.com/nao1215/mimixbox/internal/applets/textutils/unix2dos"
	"github.com/nao1215/mimixbox/internal/applets/textutils/wc"
	"github.com/nao1215/mimixbox/internal/applets/textutils/xxd"
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
		"add-shell": reg(addShell.New()),
		"ar":        reg(ar.New()),
		"arch":      reg(arch.New()),
		"awk":       reg(awk.New()),
		"base32":    reg(base32.New()),
		"base64":    reg(base64.New()),
		"basename":  reg(basename.New()),
		"bunzip2":   reg(bunzip2.New()),
		"cal":       reg(cal.New()),
		"cat":       reg(cat.New()),
		"cksum":     reg(cksum.New()),
		"chmod":     reg(chmod.New()),
		"comm":      reg(comm.New()),
		"cowsay":    reg(cowsay.New()),
		"chgrp":     reg(chgrp.New()),
		"chown":     reg(chown.New()),
		"chroot":    reg(chroot.New()),
		//"chsh":         {chsh.Run, "Cqhange login shell"},
		"clear":        reg(clear.New()),
		"cmp":          reg(cmp.New()),
		"compress":     reg(compressCmd.New()),
		"cp":           reg(cp.New()),
		"cpio":         reg(cpio.New()),
		"cut":          reg(cut.New()),
		"date":         reg(date.New()),
		"dd":           reg(dd.New()),
		"df":           reg(df.New()),
		"diff":         reg(diff.New()),
		"dirname":      reg(dirname.New()),
		"du":           reg(du.New()),
		"dos2unix":     reg(dos2unix.New()),
		"echo":         reg(echo.New()),
		"env":          reg(env.New()),
		"expand":       reg(expand.New()),
		"expr":         reg(expr.New()),
		"fakemovie":    reg(fakemovie.New()),
		"false":        reg(boolfalse.New()),
		"find":         reg(find.New()),
		"fmt":          reg(fmtCmd.New()),
		"fold":         reg(fold.New()),
		"free":         reg(free.New()),
		"ghrdc":        reg(ghrdc.New()),
		"grep":         reg(grep.New()),
		"groups":       reg(groups.New()),
		"gunzip":       reg(gunzip.New()),
		"gzip":         reg(gzipCmd.New()),
		"halt":         reg(halt.NewHalt()),
		"head":         reg(head.New()),
		"hostid":       reg(hostid.New()),
		"hostname":     reg(hostname.New()),
		"id":           reg(id.New()),
		"install":      reg(install.New()),
		"ischroot":     reg(ischroot.New()),
		"kill":         reg(kill.New()),
		"killall":      reg(killall.New()),
		"lifegame":     reg(lifegame.New()),
		"link":         reg(link.New()),
		"ln":           reg(ln.New()),
		"logname":      reg(logname.New()),
		"mbsh":         reg(mbsh.New()),
		"md5sum":       reg(md5sum.New()),
		"mkdir":        reg(mkdir.New()),
		"mkfifo":       reg(mkfifo.New()),
		"mknod":        reg(mknod.New()),
		"mktemp":       reg(mktemp.New()),
		"mountpoint":   reg(mountpoint.New()),
		"mv":           reg(mv.New()),
		"nl":           reg(nl.New()),
		"nohup":        reg(nohup.New()),
		"nproc":        reg(nproc.New()),
		"od":           reg(od.New()),
		"paste":        reg(paste.New()),
		"patch":        reg(patch.New()),
		"path":         reg(path.New()),
		"poweroff":     reg(halt.NewPoweroff()),
		"printenv":     reg(printenv.New()),
		"printf":       reg(printf.New()),
		"pwd":          reg(pwd.New()),
		"readlink":     reg(readlink.New()),
		"realpath":     reg(realpath.New()),
		"remove-shell": reg(removeShell.New()),
		"rev":          reg(rev.New()),
		"reboot":       reg(halt.NewReboot()),
		"rpm":          reg(rpm.New()),
		"rpm2cpio":     reg(rpm2cpio.New()),
		"reset":        reg(reset.New()),
		"resize":       reg(resize.New()),
		"rm":           reg(rm.New()),
		"rmdir":        reg(rmdir.New()),
		"sddf":         reg(sddf.New()),
		"serial":       reg(serial.New()),
		"sha1sum":      reg(sha1sum.New()),
		"sha256sum":    reg(sha256sum.New()),
		"sha512sum":    reg(sha512sum.New()),
		"shred":        reg(shred.New()),
		"sed":          reg(sed.New()),
		"seq":          reg(seq.New()),
		"shuf":         reg(shuf.New()),
		"sl":           reg(sl.New()),
		"sleep":        reg(sleep.New()),
		"sort":         reg(sortcmd.New()),
		"split":        reg(split.New()),
		"stat":         reg(statCmd.New()),
		"strings":      reg(stringsCmd.New()),
		"sync":         reg(sync.New()),
		"tac":          reg(tac.New()),
		"tail":         reg(tail.New()),
		"tar":          reg(tarCmd.New()),
		"tee":          reg(tee.New()),
		"test":         reg(testcmd.New()),
		"timeout":      reg(timeoutCmd.New()),
		"touch":        reg(touch.New()),
		"tr":           reg(tr.New()),
		"true":         reg(booltrue.New()),
		"truncate":     reg(truncate.New()),
		"tty":          reg(tty.New()),
		"uname":        reg(uname.New()),
		"uncompress":   reg(uncompress.New()),
		"unexpand":     reg(unexpand.New()),
		"uniq":         reg(uniq.New()),
		"unix2dos":     reg(unix2dos.New()),
		"unlink":       reg(unlink.New()),
		"unzip":        reg(unzip.New()),
		"vi":           reg(vi.New()),
		"uuidgen":      reg(uuidgen.New()),
		"valid-shell":  reg(validShell.New()),
		"watch":        reg(watch.New()),
		"wc":           reg(wc.New()),
		"wget":         reg(wget.New()),
		"which":        reg(which.New()),
		"xargs":        reg(xargs.New()),
		"xxd":          reg(xxd.New()),
		"zip":          reg(zipCmd.New()),
		"who":          reg(who.New()),
		"whoami":       reg(whoami.New()),
		"yes":          reg(yes.New()),
	}
}

func HasApplet(target string) bool {
	_, ok := Applets[target]
	return ok
}

func ListApplets() {
	format := "%" + strconv.Itoa(longestAppletLength()) + "s - %s\n"
	for _, key := range SortApplet() {
		_, _ = fmt.Fprintf(os.Stdout, format, key, Applets[key].Desc)
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
	_, _ = fmt.Fprint(os.Stdout, b.String())
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
	max := 0
	for _, key := range SortApplet() {
		if max < len(key) {
			max = len(key)
		}
	}
	return max
}
