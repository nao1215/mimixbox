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
package mv

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	mb "github.com/nao1215/mimixbox/internal/lib"

	"github.com/jessevdk/go-flags"
)

const cmdName string = "mv"

const version = "1.0.2"

var osExit = os.Exit

type options struct {
	Backup      bool `short:"b" long:"backup" description:"Backup file if same name file already exists"`
	Force       bool `short:"f" long:"force" description:"Forcibly overwrite if same name file already exists"`
	Interactive bool `short:"i" long:"interactive" description:"Check whether overwrite file or not if same name file already exists"`
	NoClobber   bool `short:"n" long:"no-clobber" description:"Don't overwrite if same name file/directory already exists"`
	Version     bool `short:"v" long:"version" description:"Show mv command version"`
}

func Run() (int, error) {
	var opts options
	var args []string
	var err error

	if args, err = parseArgs(&opts); err != nil {
		return mb.ExitFailure, nil
	}

	srcPaths, err := getSrcAbsPaths(args)
	if err != nil {
		return mb.ExitFailure, err
	}

	destPath, err := getDestAbsPath(args)
	if err != nil {
		return mb.ExitFailure, err
	}

	if err := validArgs(srcPaths, destPath, opts); err != nil {
		return mb.ExitFailure, err
	}

	return move(srcPaths, destPath, opts)
}

func validArgs(srcPaths []string, destPath string, opts options) error {
	if opts.NoClobber && opts.Backup {
		return errors.New("--noclobber and --backup can't be used at the same time")
	}

	if opts.NoClobber && opts.Force {
		return errors.New("--noclobber and --force can't be used at the same time")
	}

	if opts.Force && opts.Interactive {
		return errors.New("--force and --intractive can't be used at the same time")
	}

	if opts.NoClobber && opts.Interactive {
		return errors.New("--noclobber and --interactive can't be used at the same time")
	}
	return nil
}

func move(srcPaths []string, dest string, opts options) (int, error) {
	status := mb.ExitSuccess
	for _, src := range srcPaths {
		if !mb.Exists(src) {
			fmt.Fprintln(os.Stderr, cmdName+": "+src+" doesn't exist")
			status = mb.ExitFailure
			continue
		}

		// If SRC and DEST are the same, the option(-f, -b, -i) is ignored.
		if isSameFilePath(src, dest) {
			fmt.Fprintln(os.Stderr, cmdName+": source '"+src+"' and destination '"+dest+"' is same")
			status = mb.ExitFailure
			continue
		}

		if opts.NoClobber {
			if err := noclobberMove(src, dest); err != nil {
				fmt.Fprintln(os.Stderr, cmdName+": "+err.Error())
				status = mb.ExitFailure
			}
			continue
		}

		if opts.Force || (opts.Backup && opts.Interactive) {
			if err := forceMove(src, dest, opts); err != nil {
				fmt.Fprintln(os.Stderr, cmdName+": "+err.Error())
				status = mb.ExitFailure
			}
			continue
		}

		if opts.Interactive {
			if err := interactiveMove(src, dest, opts); err != nil {
				fmt.Fprintln(os.Stderr, cmdName+": "+err.Error())
				status = mb.ExitFailure
			}
			continue
		}

		destPath := decideDestAbsPath(src, dest, opts)
		if err := os.Rename(src, destPath); err != nil {
			fmt.Fprintln(os.Stderr, cmdName+": "+err.Error())
			status = mb.ExitFailure
		}
	}
	return status, nil
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
	if err := os.Rename(src, dest); err != nil {
		return err
	}
	return nil
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
	if err := os.Rename(src, destPath); err != nil {
		return err
	}
	return nil
}

func interactiveMove(src string, dest string, opts options) error {
	if isSameNameFileOrDir(src, dest) {
		if !mb.Question("Overwrite " + filepath.Base(src)) {
			return nil
		}
	}

	opts.Backup = false
	destPath := decideDestAbsPath(src, dest, opts)
	if err := os.Rename(src, destPath); err != nil {
		return err
	}
	return nil
}

func decideDestAbsPath(src string, dest string, opts options) string {
	destPath := os.ExpandEnv(dest)
	srcPath := os.ExpandEnv(src)
	if mb.IsDir(srcPath) && mb.IsDir(destPath) {
		destPath = filepath.Join(dest, filepath.Base(srcPath))
		if filepath.Base(srcPath) == filepath.Base(destPath) && opts.Backup {
			destPath = decideBackupFileName(destPath)
		}
	} else if mb.IsFile(srcPath) && mb.IsFile(dest) && opts.Backup {
		destPath = decideBackupFileName(destPath)
	} else if mb.IsFile(srcPath) && mb.IsDir(dest) {
		destPath = filepath.Join(dest, filepath.Base(srcPath))
		if mb.IsFile(destPath) && opts.Backup {
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

// args don't have program name(= mv).
func getSrcAbsPaths(args []string) ([]string, error) {
	var srcPaths []string
	for _, arg := range args {
		arg, err := filepath.Abs(os.ExpandEnv(arg))
		if err != nil {
			return nil, err
		}
		srcPaths = append(srcPaths, arg)
	}
	// Exclude only destination path
	return srcPaths[0 : len(args)-1], nil
}

// args don't have program name(= mv).
func getDestAbsPath(args []string) (string, error) {
	destPath, err := filepath.Abs(os.ExpandEnv(args[len(args)-1]))
	if err != nil {
		return "", err
	}
	return destPath, nil
}

func parseArgs(opts *options) ([]string, error) {
	p := initParser(opts)

	args, err := p.Parse()
	if err != nil {
		return nil, err
	}

	if opts.Version {
		mb.ShowVersion(cmdName, version)
		osExit(mb.ExitSuccess)
	}

	if !isValidArgNr(args) {
		showHelp(p)
		osExit(mb.ExitFailure)
	}
	return args, nil
}

func initParser(opts *options) *flags.Parser {
	parser := flags.NewParser(opts, flags.Default)
	parser.Name = cmdName
	parser.Usage = "[OPTIONS] SOURCE_PATH DESTINATION_PATH"

	return parser
}

func isValidArgNr(args []string) bool {
	return len(args) >= 2
}

func showHelp(p *flags.Parser) {
	p.WriteHelp(os.Stdout)
}
