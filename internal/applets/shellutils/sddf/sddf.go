//
// mimixbox/internal/applets/shellutils/sddf/sddf.go
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

// Package sddf implements the sddf applet: a MimixBox-original "Search & Delete
// Duplicated File" command. It walks one or more directories, finds files whose
// byte content is identical (compared by md5 checksum for speed), and either
// writes the duplicate groups to a *.sddf report or deletes the duplicates
// in place, keeping the newest copy of each group.
package sddf

import (
	"bufio"
	"context"
	"crypto/md5" //nolint:gosec // md5 is used as a fast content fingerprint, not for security
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

const ext string = ".sddf"

// importantPrefixes are absolute paths excluded from the duplicate scan so that
// the command never touches system files.
var importantPrefixes = []string{
	"/boot", "/dev", "/etc", "/lib", "/lib32", "/lib64", "/libx32", "/lost+found",
	"/proc", "/root", "/run", "/sys", "/bin", "/sbin",
}

// Command is the sddf applet.
type Command struct{}

// New returns an sddf command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "sddf" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Search & Delete Duplicated File" }

type options struct {
	output      string
	delete      bool
	interactive bool
	dryRun      bool
}

// Run executes sddf.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]... DIRECTORY...", stdio.Err).WithHelp(command.Help{
		Description: "Search for and delete duplicated files (a MimixBox-original command). It walks " +
			"each DIRECTORY, groups files whose byte content is identical, and by default writes " +
			"the duplicate groups to a *.sddf report. With -d the duplicates are deleted in place, " +
			"keeping the newest copy of each group; -n reports what would be deleted without " +
			"removing anything, and -i prompts before each deletion. A previously written *.sddf " +
			"report may be passed as an operand to delete the duplicates it records. System paths " +
			"such as /bin and /etc are always excluded from the scan.",
		Examples: []command.Example{
			{Command: "sddf ~/Pictures", Explain: "Scan ~/Pictures and write a *.sddf report of duplicates."},
			{Command: "sddf -d ~/Downloads", Explain: "Delete duplicates in ~/Downloads, keeping the newest copy."},
			{Command: "sddf -n ~/Downloads", Explain: "Show which duplicates would be deleted, without removing them."},
			{Command: "sddf duplicated-file.sddf", Explain: "Delete the duplicates recorded in a saved report."},
		},
		ExitStatus: "0  the scan or deletion completed successfully.\n1  a directory could not be scanned or a file could not be deleted, or no operand was given.",
	})
	output := fs.StringP("output", "o", "duplicated-file", "Change output file-name without extension")
	del := fs.BoolP("delete", "d", false, "delete duplicated files in place, keeping the newest copy")
	interactive := fs.BoolP("interactive", "i", false, "prompt before each deletion (with --delete)")
	dryRun := fs.BoolP("dry-run", "n", false, "report what would be deleted without removing anything")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	opts := options{
		output:      *output,
		delete:      *del,
		interactive: *interactive,
		dryRun:      *dryRun,
	}

	dirs := fs.Args()
	if len(dirs) == 0 {
		_, _ = fmt.Fprintf(stdio.Err, "%s: missing directory operand\n", c.Name())
		return command.SilentFailure()
	}

	return c.run(stdio, dirs, opts)
}

// run dispatches each operand. A lone *.sddf file restores a previously written
// report and deletes from it; otherwise the operand is treated as a directory to
// scan.
func (c *Command) run(stdio command.IO, dirs []string, opts options) error {
	var failed bool
	in := bufio.NewReader(stdio.In)

	for _, dir := range dirs {
		path := os.ExpandEnv(dir)

		info, err := os.Stat(path)
		if err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "%s: %s\n", c.Name(), command.FileError(path, err))
			failed = true
			continue
		}

		if !info.IsDir() {
			// A *.sddf report: restore the groups and delete from them.
			if err := c.restoreAndDelete(stdio, in, path, opts); err != nil {
				_, _ = fmt.Fprintf(stdio.Err, "%s: %s\n", c.Name(), err.Error())
				failed = true
			}
			continue
		}

		if err := c.scan(stdio, in, path, opts); err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "%s: %s\n", c.Name(), err.Error())
			failed = true
		}
	}

	if failed {
		return command.SilentFailure()
	}
	return nil
}

// scan walks a directory, finds duplicate groups, and either deletes them
// (--delete) or writes them to a *.sddf report.
func (c *Command) scan(stdio command.IO, in *bufio.Reader, dir string, opts options) error {
	_, _ = fmt.Fprintln(stdio.Out, "Get all file path at "+dir)
	files := collectFiles(dir)
	if len(files) == 0 {
		_, _ = fmt.Fprintln(stdio.Out, dir+" has no file")
		return nil
	}

	_, _ = fmt.Fprintln(stdio.Out, "Find the same file on a file content basis")
	groups := findDuplicates(files)

	if opts.delete || opts.dryRun {
		return c.deleteGroups(stdio, in, groups, opts)
	}
	return dumpToFile(stdio, groups, decideOutputFileName(opts.output))
}

// fileGroup is one set of byte-identical files, identified by their shared
// content checksum.
type fileGroup struct {
	Checksum string
	Paths    []string
}

// findDuplicates is the pure, testable core of sddf: it hashes every file by
// content and returns only the groups that contain more than one file (i.e. the
// duplicates). Files that cannot be read are skipped. The returned groups are
// independent of any IO or deletion side effect; the paths inside each group are
// kept in the order the files were supplied.
func findDuplicates(files []string) []fileGroup {
	order := []string{} // checksums in first-seen order, for deterministic output
	byChecksum := map[string][]string{}

	for _, path := range files {
		checksum, err := hashFile(path)
		if err != nil {
			continue
		}
		if _, ok := byChecksum[checksum]; !ok {
			order = append(order, checksum)
		}
		byChecksum[checksum] = append(byChecksum[checksum], path)
	}

	groups := []fileGroup{}
	for _, checksum := range order {
		paths := byChecksum[checksum]
		if len(paths) <= 1 {
			continue
		}
		groups = append(groups, fileGroup{Checksum: checksum, Paths: paths})
	}
	return groups
}

// hashFile returns the md5 hex digest of the file's content.
func hashFile(path string) (string, error) {
	f, err := os.Open(path) //nolint:gosec // operating on a user-named file is the point
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }()

	h := md5.New() //nolint:gosec // content fingerprint only
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

// collectFiles returns every regular file under dir, excluding system paths and
// named pipes (whose checksum read would block).
func collectFiles(dir string) []string {
	files := []string{}
	_ = walkFiles(dir, func(path string, info os.FileInfo) {
		if info.IsDir() {
			return
		}
		if hasImportantPrefix(path) {
			return
		}
		if isNamedPipe(info) {
			return
		}
		files = append(files, path)
	})
	return files
}

// walkFiles visits every entry under dir, ignoring read errors so an
// unreadable subtree does not abort the whole scan.
func walkFiles(dir string, fn func(path string, info os.FileInfo)) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		path := dir + string(os.PathSeparator) + e.Name()
		info, err := e.Info()
		if err != nil {
			continue
		}
		if info.IsDir() {
			fn(path, info)
			_ = walkFiles(path, fn)
			continue
		}
		fn(path, info)
	}
	return nil
}

func isNamedPipe(info os.FileInfo) bool {
	return info.Mode()&os.ModeNamedPipe != 0
}

func hasImportantPrefix(path string) bool {
	for _, p := range importantPrefixes {
		if path == p || strings.HasPrefix(path, p+"/") {
			return true
		}
	}
	return false
}

// deleteGroups removes the duplicates in each group, keeping the newest file.
// With --dry-run nothing is removed; with --interactive each removal is
// confirmed by reading "y" from stdio.In.
func (c *Command) deleteGroups(stdio command.IO, in *bufio.Reader, groups []fileGroup, opts options) error {
	_, _ = fmt.Fprintln(stdio.Out, "Decide delete target files")
	targets := []string{}
	for _, g := range groups {
		targets = append(targets, deleteTargets(g.Paths)...)
	}

	_, _ = fmt.Fprintln(stdio.Out, "Start deleting files")
	var sum int64
	var failed bool
	for _, path := range targets {
		size := fileSize(path)
		if opts.dryRun {
			_, _ = fmt.Fprintln(stdio.Out, "Delete(DryRun): "+path)
			sum += size
			continue
		}
		if !confirm(stdio, in, path, opts) {
			continue
		}
		if err := os.Remove(path); err != nil {
			_, _ = fmt.Fprintln(stdio.Out, "Delete(Failure): "+path)
			failed = true
			continue
		}
		_, _ = fmt.Fprintln(stdio.Out, "Delete(Success): "+path+": "+strconv.FormatInt(size, 10)+"Byte")
		sum += size
	}
	_, _ = fmt.Fprintln(stdio.Out, "End deleting files. Size="+strconv.FormatInt(sum, 10)+"Byte")

	if failed {
		return fmt.Errorf("failed to delete one or more files")
	}
	return nil
}

// confirm asks the user before removing path when -i is set. The prompt is
// written to stdio.Err and the answer is read from stdio.In (never os.Stdin).
// Answers starting with "y" (case-insensitive) approve; anything else, EOF
// included, keeps the file.
func confirm(stdio command.IO, in *bufio.Reader, path string, opts options) bool {
	if !opts.interactive {
		return true
	}
	_, _ = fmt.Fprintf(stdio.Err, "sddf: remove '%s'? ", path)
	line, err := in.ReadString('\n')
	answer := strings.ToLower(strings.TrimSpace(line))
	if err != nil && answer == "" {
		return false
	}
	return strings.HasPrefix(answer, "y")
}

// deleteTargets returns every path in a duplicate group except the newest one,
// which is kept.
func deleteTargets(paths []string) []string {
	if len(paths) <= 1 {
		return nil
	}
	var newest string
	var newestUnix int64 = -1
	for _, p := range paths {
		info, err := os.Stat(p)
		if err != nil {
			continue
		}
		if mod := info.ModTime().Unix(); mod > newestUnix {
			newestUnix = mod
			newest = p
		}
	}

	targets := []string{}
	for _, p := range paths {
		if p == newest {
			continue
		}
		targets = append(targets, p)
	}
	return targets
}

func fileSize(path string) int64 {
	info, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return info.Size()
}

// dumpToFile writes the duplicate groups to a *.sddf report so they can later be
// deleted with "sddf <report>".
func dumpToFile(stdio command.IO, groups []fileGroup, output string) (err error) {
	_, _ = fmt.Fprintln(stdio.Out, "Write down duplicated file list to "+output)
	f, err := os.Create(output) //nolint:gosec // user-named output file
	if err != nil {
		return err
	}
	defer func() {
		if cerr := f.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	var b strings.Builder
	for _, g := range groups {
		b.WriteByte('[')
		b.WriteString(g.Checksum)
		b.WriteString("]\n")
		for _, p := range g.Paths {
			b.WriteString(p)
			b.WriteByte('\n')
		}
		b.WriteByte('\n')
	}
	if _, werr := f.WriteString(b.String()); werr != nil {
		return werr
	}

	_, _ = fmt.Fprintln(stdio.Out, "See duplicated file list: "+output)
	_, _ = fmt.Fprintln(stdio.Out, "If you delete files, execute the following command.")
	_, _ = fmt.Fprintln(stdio.Out, "$ sddf "+output)
	return nil
}

// restoreAndDelete reads a *.sddf report and deletes its duplicate groups.
func (c *Command) restoreAndDelete(stdio command.IO, in *bufio.Reader, path string, opts options) error {
	if !strings.HasSuffix(path, ext) {
		return fmt.Errorf("%s: file format is not *.sddf", path)
	}
	groups, err := restore(path)
	if err != nil {
		return err
	}
	// A restored report always deletes; honor --dry-run / --interactive.
	opts.delete = true
	_, _ = fmt.Fprintln(stdio.Out, "Restore data from "+path)
	return c.deleteGroups(stdio, in, groups, opts)
}

// restore parses a *.sddf report back into duplicate groups.
func restore(path string) ([]fileGroup, error) {
	f, err := os.Open(path) //nolint:gosec // user-named report file
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	groups := []fileGroup{}
	var cur *fileGroup
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimRight(scanner.Text(), "\r")
		switch {
		case isChecksumLine(line):
			groups = append(groups, fileGroup{Checksum: strings.Trim(line, "[]")})
			cur = &groups[len(groups)-1]
		case line == "":
			cur = nil
		default:
			if cur != nil {
				cur.Paths = append(cur.Paths, line)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return groups, nil
}

func isChecksumLine(line string) bool {
	return strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") && len(line) == 34
}

func decideOutputFileName(output string) string {
	if strings.HasSuffix(output, ext) {
		return output
	}
	return output + ext
}
