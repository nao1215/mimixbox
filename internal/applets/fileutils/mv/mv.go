//
// mimixbox/internal/applets/fileutils/mv/mv.go
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

// Package mv implements the mv applet: rename SOURCE to DESTINATION, or move
// SOURCE(s) to DIRECTORY, with the common GNU options.
package mv

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
	mb "github.com/nao1215/mimixbox/internal/lib"
)

// Command is the mv applet.
type Command struct{}

// New returns a mv command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "mv" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string {
	return "Rename SOURCE to DESTINATION, or move SOURCE(s) to DIRECTORY"
}

type options struct {
	backup      bool
	force       bool
	interactive bool
	noClobber   bool
	verbose     bool
	update      bool
	targetDir   string
	noTargetDir bool
}

// Run executes mv.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]... SOURCE... DEST", stdio.Err).WithHelp(command.Help{
		Description: "Rename SOURCE to DEST, or move one or more SOURCEs into the directory DEST. Moves across file systems fall back to a copy followed by a delete. Use -i to prompt before overwriting, -n to never overwrite, -f to overwrite without prompting, and -b to back up an existing destination.",
		Examples: []command.Example{
			{Command: "mv old.txt new.txt", Explain: "Rename old.txt to new.txt."},
			{Command: "mv a.txt b.txt dir/", Explain: "Move a.txt and b.txt into the directory 'dir'."},
		},
		ExitStatus: "0  every move succeeded.\n1  a source was missing or could not be moved.",
	})
	backup := fs.BoolP("backup", "b", false, "make a backup of each existing destination file")
	force := fs.BoolP("force", "f", false, "do not prompt before overwriting")
	interactive := fs.BoolP("interactive", "i", false, "prompt before overwrite")
	noClobber := fs.BoolP("no-clobber", "n", false, "do not overwrite an existing file")
	verbose := fs.BoolP("verbose", "v", false, "explain what is being done")
	update := fs.BoolP("update", "u", false, "move only when the SOURCE is newer than the destination, or when the destination is missing")
	targetDir := fs.StringP("target-directory", "t", "", "move all SOURCE arguments into DIRECTORY")
	noTargetDir := fs.BoolP("no-target-directory", "T", false, "treat DEST as a normal file always")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	opts := options{
		backup:      *backup,
		force:       *force,
		interactive: *interactive,
		noClobber:   *noClobber,
		verbose:     *verbose,
		update:      *update,
		targetDir:   *targetDir,
		noTargetDir: *noTargetDir,
	}

	operands := fs.Args()

	// -t DIR: every operand is a source moved into DIR. The destination is the
	// target directory rather than the last operand.
	if opts.targetDir != "" {
		if len(operands) == 0 {
			_, _ = fmt.Fprintln(stdio.Err, "mv: missing file operand")
			return command.SilentFailure()
		}
		srcPaths := make([]string, 0, len(operands))
		for _, arg := range operands {
			abs, err := filepath.Abs(os.ExpandEnv(arg))
			if err != nil {
				return command.Failure(err)
			}
			srcPaths = append(srcPaths, abs)
		}
		destPath, err := filepath.Abs(os.ExpandEnv(opts.targetDir))
		if err != nil {
			return command.Failure(err)
		}
		if err := validArgs(opts); err != nil {
			return command.Failure(err)
		}
		return c.move(stdio, srcPaths, destPath, opts)
	}

	if len(operands) == 0 {
		_, _ = fmt.Fprintln(stdio.Err, "mv: missing file operand")
		return command.SilentFailure()
	}
	if len(operands) == 1 {
		_, _ = fmt.Fprintf(stdio.Err, "mv: missing destination file operand after '%s'\n", operands[0])
		return command.SilentFailure()
	}

	// -T: the destination is always a normal file; reject the multi-source
	// "mv SOURCE... DIRECTORY" form.
	if opts.noTargetDir && len(operands) > 2 {
		_, _ = fmt.Fprintf(stdio.Err, "mv: extra operand '%s'\n", operands[2])
		return command.SilentFailure()
	}

	srcPaths, err := getSrcAbsPaths(operands)
	if err != nil {
		return command.Failure(err)
	}
	destPath, err := getDestAbsPath(operands)
	if err != nil {
		return command.Failure(err)
	}

	if err := validArgs(opts); err != nil {
		return command.Failure(err)
	}

	return c.move(stdio, srcPaths, destPath, opts)
}

func validArgs(opts options) error {
	if opts.noClobber && opts.backup {
		return errors.New("--noclobber and --backup can't be used at the same time")
	}
	if opts.noClobber && opts.force {
		return errors.New("--noclobber and --force can't be used at the same time")
	}
	if opts.force && opts.interactive {
		return errors.New("--force and --intractive can't be used at the same time")
	}
	if opts.noClobber && opts.interactive {
		return errors.New("--noclobber and --interactive can't be used at the same time")
	}
	return nil
}

func (c *Command) move(stdio command.IO, srcPaths []string, dest string, opts options) error {
	var failed bool
	for _, src := range srcPaths {
		if !mb.Exists(src) {
			_, _ = fmt.Fprintln(stdio.Err, "mv: "+src+" doesn't exist")
			failed = true
			continue
		}

		// If SRC and DEST are the same, the option(-f, -b, -i) is ignored.
		if isSameFilePath(src, dest) {
			_, _ = fmt.Fprintln(stdio.Err, "mv: source '"+src+"' and destination '"+dest+"' is same")
			failed = true
			continue
		}

		// -u: skip the move when the destination exists and is at least as new
		// as the source. A missing destination always moves.
		if opts.update && !shouldUpdate(src, decideDestAbsPath(src, dest, opts)) {
			continue
		}

		if opts.noClobber {
			if err := noclobberMove(src, dest); err != nil {
				_, _ = fmt.Fprintln(stdio.Err, "mv: "+err.Error())
				failed = true
				continue
			}
			c.report(stdio, src, dest, opts)
			continue
		}

		if opts.force || (opts.backup && opts.interactive) {
			if err := forceMove(src, dest, opts); err != nil {
				_, _ = fmt.Fprintln(stdio.Err, "mv: "+err.Error())
				failed = true
				continue
			}
			c.report(stdio, src, dest, opts)
			continue
		}

		if opts.interactive {
			if err := interactiveMove(stdio, src, dest, opts); err != nil {
				_, _ = fmt.Fprintln(stdio.Err, "mv: "+err.Error())
				failed = true
				continue
			}
			c.report(stdio, src, dest, opts)
			continue
		}

		destPath := decideDestAbsPath(src, dest, opts)
		if err := rename(src, destPath); err != nil {
			_, _ = fmt.Fprintln(stdio.Err, "mv: "+err.Error())
			failed = true
			continue
		}
		c.report(stdio, src, dest, opts)
	}
	if failed {
		return command.SilentFailure()
	}
	return nil
}

func (c *Command) report(stdio command.IO, src, dest string, opts options) {
	if opts.verbose {
		_, _ = fmt.Fprintf(stdio.Out, "renamed '%s' -> '%s'\n", src, decideDestAbsPath(src, dest, opts))
	}
}

// rename moves src to dest, falling back to copy+remove when the rename crosses
// a device boundary (os.Rename fails with EXDEV in that case). The fallback
// preserves mode and timestamps and handles directories recursively, so a
// cross-filesystem move behaves like an in-filesystem one.
// osRename is os.Rename, indirected so tests can force the cross-device
// fallback path that real filesystems only take across mount points.
var osRename = os.Rename

func rename(src, dest string) error {
	if err := osRename(src, dest); err != nil {
		if !isCrossDevice(err) {
			return err
		}
		if cerr := mb.CopyTree(src, dest); cerr != nil {
			return cerr
		}
		return os.RemoveAll(src)
	}
	return nil
}

func isCrossDevice(err error) bool {
	var le *os.LinkError
	if errors.As(err, &le) {
		return strings.Contains(le.Err.Error(), "cross-device")
	}
	return strings.Contains(err.Error(), "cross-device")
}

func noclobberMove(src string, dest string) error {
	if isSameNameFileOrDir(src, dest) {
		return nil // Nothing to do. Say nothing.
	}
	if mb.IsFile(src) && mb.IsFile(dest) {
		if filepath.Base(src) == filepath.Base(dest) {
			return nil // Nothing to do. Say nothing.
		}
	}
	return rename(src, dest)
}

func isSameNameFileOrDir(src string, dest string) bool {
	if mb.IsDir(src) && mb.IsDir(dest) {
		if filepath.Base(src) == filepath.Base(dest) {
			return true
		}
	}
	if mb.IsFile(src) && mb.IsFile(dest) {
		if filepath.Base(src) == filepath.Base(dest) {
			return true
		}
	} else if mb.IsFile(src) && mb.IsDir(dest) {
		destPath := filepath.Join(dest, filepath.Base(src))
		if mb.Exists(destPath) {
			return true
		}
	}
	return false
}

func forceMove(src string, dest string, opts options) error {
	destPath := decideDestAbsPath(src, dest, opts)
	return rename(src, destPath)
}

func interactiveMove(stdio command.IO, src string, dest string, opts options) error {
	if isSameNameFileOrDir(src, dest) {
		if !question(stdio, "Overwrite "+filepath.Base(src)) {
			return nil
		}
	}

	opts.backup = false
	destPath := decideDestAbsPath(src, dest, opts)
	return rename(src, destPath)
}

// question prompts on stdio.Out and reads the answer from stdio.In, returning
// true only when the user types a "yes" answer.
func question(stdio command.IO, ask string) bool {
	_, _ = fmt.Fprintf(stdio.Out, "%s [Y/n] ", ask)
	r := bufio.NewReader(stdio.In)
	line, err := r.ReadString('\n')
	if err != nil && line == "" {
		return false
	}
	switch strings.ToLower(strings.TrimSpace(line)) {
	case "y", "yes":
		return true
	default:
		return false
	}
}

func decideDestAbsPath(src string, dest string, opts options) string {
	destPath := os.ExpandEnv(dest)
	srcPath := os.ExpandEnv(src)
	// -T: never descend into an existing destination directory; the
	// destination path is used verbatim (with optional backup).
	if opts.noTargetDir {
		if mb.Exists(destPath) && opts.backup {
			destPath = decideBackupFileName(destPath)
		}
		return destPath
	}
	if mb.IsDir(srcPath) && mb.IsDir(destPath) {
		destPath = filepath.Join(dest, filepath.Base(srcPath))
		if filepath.Base(srcPath) == filepath.Base(destPath) && opts.backup {
			destPath = decideBackupFileName(destPath)
		}
	} else if mb.IsFile(srcPath) && mb.IsFile(dest) && opts.backup {
		destPath = decideBackupFileName(destPath)
	} else if mb.IsFile(srcPath) && mb.IsDir(dest) {
		destPath = filepath.Join(dest, filepath.Base(srcPath))
		if mb.IsFile(destPath) && opts.backup {
			destPath = decideBackupFileName(destPath)
		}
	}
	return destPath
}

func decideBackupFileName(path string) string {
	var backupPath string
	if mb.Exists(path) {
		backupPath = path + mb.SimpleBackupSuffix()
	}
	if mb.Exists(backupPath) {
		return decideBackupFileName(backupPath)
	}
	return backupPath
}

func isSameFilePath(src string, dest string) bool {
	return src == dest
}

// shouldUpdate reports whether -u should move src to destPath. It returns true
// when the destination is missing or when the source's modification time is
// strictly newer than the destination's; an equal-or-older source is skipped.
func shouldUpdate(src, destPath string) bool {
	destInfo, err := os.Stat(destPath)
	if err != nil {
		// Missing (or unstattable) destination: always move.
		return true
	}
	srcInfo, err := os.Stat(src)
	if err != nil {
		// Let the later move attempt surface the source error.
		return true
	}
	return srcInfo.ModTime().After(destInfo.ModTime())
}

// getSrcAbsPaths returns the absolute paths of every operand except the last
// (which is the destination). operands does not include the program name.
func getSrcAbsPaths(operands []string) ([]string, error) {
	var srcPaths []string
	for _, arg := range operands {
		abs, err := filepath.Abs(os.ExpandEnv(arg))
		if err != nil {
			return nil, err
		}
		srcPaths = append(srcPaths, abs)
	}
	// Exclude only destination path
	return srcPaths[0 : len(operands)-1], nil
}

// getDestAbsPath returns the absolute path of the destination operand (the last
// operand). operands does not include the program name.
func getDestAbsPath(operands []string) (string, error) {
	return filepath.Abs(os.ExpandEnv(operands[len(operands)-1]))
}
