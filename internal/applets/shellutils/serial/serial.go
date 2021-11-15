//
//  mimixbox/internal/applets/shellutils/serial/serial.go
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
package serial

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"mimixbox/pkg/fileutils"

	"github.com/jessevdk/go-flags"
)

const cmdName string = "serial"

var osExit = os.Exit

const version = "1.0.1"

// Exit code
const (
	ExitSuccess int = iota // 0
	ExitFailuer
)

type options struct {
	DryRun  bool   `short:"d" long:"dry-run" description:"Output the file renaming result to standard output (do not update the file)"`
	Force   bool   `short:"f" long:"force" description:"Forcibly overwrite and save even if a file with the same name exists"`
	Keep    bool   `short:"k" long:"keep" description:"Keep the file before renaming"`
	Name    string `short:"n" long:"name" value-name:"<name>" description:"Base file name with/without directory path (assign a serial number to this file name)"`
	Prefix  bool   `short:"p" long:"prefix" description:"Add a serial number to the beginning of the file name(default)"`
	Suffix  bool   `short:"s" long:"suffix" description:"Add a serial number to the end of the file name"`
	Version bool   `short:"v" long:"version" description:"Show serial command version"`
}

func Run() (int, error) {
	var opts options
	var args = parseArgs(&opts)
	var dirPath = args[0]

	if !fileutils.Exists(dirPath) {
		err := fmt.Errorf("%s doesn't exist.", dirPath)
		return ExitFailuer, err
	}

	var files = getFilePathsInDir(dirPath)
	if len(files) == 0 {
		err := fmt.Errorf("No files in %s directory.", dirPath)
		return ExitFailuer, err
	}

	newFileNames := newNames(opts, files)
	dieIfExistSameNameFile(opts.Force, newFileNames)
	makeDirIfNeeded(newFileNames[files[0]])

	if opts.Keep {
		copy(newFileNames, opts.DryRun)
	} else {
		rename(newFileNames, opts.DryRun)
	}
	return ExitSuccess, nil
}

func rename(newFileNames map[string]string, dryRun bool) {
	keys := make([]string, 0, len(newFileNames))
	for k := range newFileNames {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, org := range keys {
		fmt.Printf("Rename %s to %s\n", org, newFileNames[org])
		if dryRun {
			continue
		}
		if err := os.Rename(org, newFileNames[org]); err != nil {
			fmt.Fprintf(os.Stderr, "Can't rename %s to %s\n", org, newFileNames[org])
			osExit(1)
		}
	}
}

func copy(newFileNames map[string]string, dryRun bool) {
	var dest string
	keys := make([]string, 0, len(newFileNames))

	for k := range newFileNames {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, org := range keys {
		dest = newFileNames[org]
		fmt.Printf("Copy %s to %s\n", org, dest)
		if dryRun {
			continue
		}
		// In the case of renaming, even the same file name can be overwritten.
		// On the other hand, in the case of copying, an error will occur
		// if serial command try to overwrite with the same file name.
		if org == dest {
			continue
		}

		// If this function is running, it will force the file to be overwritten.
		// If there is the file with the same name in the copy destination,
		// delete it before copy the file.
		if fileutils.Exists(dest) {
			if err := os.Remove(dest); err != nil {
				fmt.Fprintf(os.Stderr, "Can't copy %s to %s\n", org, dest)
				osExit(ExitFailuer)
			}
		}

		if err := os.Link(org, dest); err != nil {
			fmt.Fprintf(os.Stderr, "Can't copy %s to %s\n", org, dest)
			osExit(ExitFailuer)
		}
	}
}

func parseArgs(opts *options) []string {
	p := initParser(opts)

	args, err := p.Parse()
	if err != nil {
		osExit(ExitFailuer)
	}

	if opts.Version {
		showVersion()
		osExit(ExitSuccess)
	}

	if len(opts.Name) != 0 && !existFilenameInPath(opts.Name) {
		showHelp(p)
		osExit(ExitFailuer)
	}

	if !isValidArgNr(args) {
		showHelp(p)
		osExit(ExitFailuer)
	}

	return args
}

func initParser(opts *options) *flags.Parser {
	parser := flags.NewParser(opts, flags.Default)
	parser.Name = cmdName
	parser.Usage = "[OPTIONS] DIRECTORY_PATH"

	return parser
}

func isValidArgNr(args []string) bool {
	return len(args) == 1
}

func showVersion() {
	fmt.Printf("serial version %s\n", version)
}

func showHelp(p *flags.Parser) {
	fmt.Printf("serial command rename the file name to the name with a serial number.\n\n")
	p.WriteHelp(os.Stdout)
}

func existFilenameInPath(path string) bool {

	return !strings.HasSuffix(path, "/")
}

// getFilePathsInDir returns the paths of the file in the directory.
func getFilePathsInDir(dir string) []string {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Can't get file list.")
		osExit(ExitFailuer)
	}

	var path string
	var paths []string
	for _, file := range files {
		path = filepath.Join(dir, file.Name())
		if fileutils.IsFile(path) && !fileutils.IsHiddenFile(path) {
			paths = append(paths, filepath.Clean(path))
		}
	}
	sort.Strings(paths)
	return paths
}

func newNames(opts options, path []string) map[string]string {
	newNames := make(map[string]string)
	destDir := filepath.Dir(opts.Name)

	var fileName string
	var format string
	// TODO: Refactor for a simpler implementation
	for i, file := range path {
		ext := filepath.Ext(file)

		if len(opts.Name) == 0 {
			format = fileNameFormat(opts.Prefix, opts.Suffix, fileutils.BaseNameWithoutExt(file), len(path))
		} else {
			format = fileNameFormat(opts.Prefix, opts.Suffix, opts.Name, len(path))
		}

		if opts.Prefix && !opts.Suffix {
			fileName = fmt.Sprintf(format, i, ext)
		} else {
			fileName = fmt.Sprintf(format, i, ext)
		}

		if destDir == "." {
			newNames[file] = filepath.Clean(fileName)
		} else {
			newNames[file] = filepath.Clean(destDir + "/" + fileName)
		}
	}
	return newNames
}

func fileNameFormat(prefix bool, suffix bool, name string, totalFileNr int) string {
	baseName := filepath.Base(name)
	serial := "%0" + strconv.Itoa(len(strconv.Itoa(totalFileNr))) + "d"
	ext := "%s"

	// Default format（e.x.：%03d%s%s → 001_test.txt）
	format := serial + "_" + baseName + ext
	if !prefix && suffix {
		format = baseName + "_" + serial + ext
	}
	return format
}

func dieIfExistSameNameFile(force bool, fileNames map[string]string) {
	if force {
		return
	}

	for _, file := range fileNames {
		if fileutils.Exists(file) {
			fmt.Fprintf(os.Stderr, "%s (file name which is after renaming) is already exists.\n", file)
			fmt.Fprintf(os.Stderr, "Renaming may erase the contents of the file. ")
			fmt.Fprintf(os.Stderr, "So, nothing to do.\n")
			osExit(ExitFailuer)
		}
	}
}

func makeDirIfNeeded(filePath string) {
	dirPath := filepath.Dir(filePath)

	if fileutils.Exists(dirPath) {
		return
	}

	if err := os.MkdirAll(dirPath, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Can't make %s directory\n", dirPath)
		osExit(ExitFailuer)
	}
}
