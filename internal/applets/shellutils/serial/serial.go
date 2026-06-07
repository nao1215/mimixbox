// Package serial implements the serial applet: rename (or copy) the files in a
// directory to names that carry a zero-padded serial number. serial is a
// MimixBox-original command; it adds the serial number as a prefix (the
// default) or a suffix, and can optionally replace the base file name.
package serial

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the serial applet.
type Command struct{}

// New returns a serial command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "serial" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Rename the file to the name with a serial number" }

type options struct {
	dryRun bool
	force  bool
	keep   bool
	name   string
	prefix bool
	suffix bool
}

// Run executes serial.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]... DIRECTORY", stdio.Err)
	dryRun := fs.BoolP("dry-run", "d", false, "Output the file renaming result to standard output (do not update the file)")
	force := fs.BoolP("force", "f", false, "Forcibly overwrite and save even if a file with the same name exists")
	keep := fs.BoolP("keep", "k", false, "Keep the file before renaming")
	name := fs.StringP("name", "n", "", "Base file name with/without directory path (assign a serial number to this file name)")
	prefix := fs.BoolP("prefix", "p", false, "Add a serial number to the beginning of the file name(default)")
	suffix := fs.BoolP("suffix", "s", false, "Add a serial number to the end of the file name")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	opts := options{
		dryRun: *dryRun,
		force:  *force,
		keep:   *keep,
		name:   *name,
		prefix: *prefix,
		suffix: *suffix,
	}

	operands := fs.Args()
	if len(operands) == 0 {
		fmt.Fprintln(stdio.Err, "serial: missing operand")
		return command.SilentFailure()
	}
	if len(operands) != 1 {
		fmt.Fprintf(stdio.Err, "serial: extra operand '%s'\n", operands[1])
		return command.SilentFailure()
	}
	if opts.name != "" && strings.HasSuffix(opts.name, "/") {
		fmt.Fprintf(stdio.Err, "serial: invalid --name '%s' (must include a file name)\n", opts.name)
		return command.SilentFailure()
	}

	dirPath := operands[0]
	if !exists(dirPath) {
		return command.Failuref("%s doesn't exist.", dirPath)
	}

	files, err := getFilePathsInDir(dirPath)
	if err != nil {
		return command.Failure(err)
	}
	if len(files) == 0 {
		return command.Failuref("No files in %s directory.", dirPath)
	}

	newFileNames := newNames(opts, files)
	if err := dieIfExistSameNameFile(opts.force, newFileNames); err != nil {
		return command.Failure(err)
	}
	if err := makeDirIfNeeded(newFileNames[files[0]]); err != nil {
		return command.Failure(err)
	}

	if opts.keep {
		return copyFiles(stdio, newFileNames, opts.dryRun)
	}
	return rename(stdio, newFileNames, opts.dryRun)
}

func rename(stdio command.IO, newFileNames map[string]string, dryRun bool) error {
	for _, org := range sortedKeys(newFileNames) {
		fmt.Fprintf(stdio.Out, "Rename %s to %s\n", org, newFileNames[org])
		if dryRun {
			continue
		}
		if err := os.Rename(org, newFileNames[org]); err != nil {
			fmt.Fprintf(stdio.Err, "Can't rename %s to %s\n", org, newFileNames[org])
			return command.SilentFailure()
		}
	}
	return nil
}

func copyFiles(stdio command.IO, newFileNames map[string]string, dryRun bool) error {
	for _, org := range sortedKeys(newFileNames) {
		dest := newFileNames[org]
		fmt.Fprintf(stdio.Out, "Copy %s to %s\n", org, dest)
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
		if exists(dest) {
			if err := os.Remove(dest); err != nil {
				fmt.Fprintf(stdio.Err, "Can't copy %s to %s\n", org, dest)
				return command.SilentFailure()
			}
		}

		if err := os.Link(org, dest); err != nil {
			fmt.Fprintf(stdio.Err, "Can't copy %s to %s\n", org, dest)
			return command.SilentFailure()
		}
	}
	return nil
}

func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// getFilePathsInDir returns the cleaned, sorted paths of the regular,
// non-hidden files directly under dir.
func getFilePathsInDir(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("can't get file list of %s", dir)
	}

	var paths []string
	for _, entry := range entries {
		path := filepath.Join(dir, entry.Name())
		if isFile(path) && !isHiddenFile(path) {
			paths = append(paths, filepath.Clean(path))
		}
	}
	sort.Strings(paths)
	return paths, nil
}

func newNames(opts options, path []string) map[string]string {
	newNames := make(map[string]string)
	// When --name carries a directory, files are placed there; otherwise each
	// renamed file stays in the directory it came from.
	nameDir := filepath.Dir(opts.name)

	for i, file := range path {
		ext := filepath.Ext(file)

		var format string
		if opts.name == "" {
			format = fileNameFormat(opts.prefix, opts.suffix, baseNameWithoutExt(file), len(path))
		} else {
			format = fileNameFormat(opts.prefix, opts.suffix, opts.name, len(path))
		}

		fileName := fmt.Sprintf(format, i, ext)

		destDir := nameDir
		if opts.name == "" || nameDir == "." {
			destDir = filepath.Dir(file)
		}
		newNames[file] = filepath.Clean(filepath.Join(destDir, fileName))
	}
	return newNames
}

func fileNameFormat(prefix bool, suffix bool, name string, totalFileNr int) string {
	baseName := filepath.Base(name)
	serial := "%0" + strconv.Itoa(len(strconv.Itoa(totalFileNr))) + "d"
	ext := "%s"

	// Default format (e.g. %01d_test%s -> 0_test.txt). The serial number is a
	// prefix unless --suffix is requested without --prefix.
	format := serial + "_" + baseName + ext
	if !prefix && suffix {
		format = baseName + "_" + serial + ext
	}
	return format
}

func dieIfExistSameNameFile(force bool, fileNames map[string]string) error {
	if force {
		return nil
	}
	for _, file := range fileNames {
		if exists(file) {
			return fmt.Errorf("%s (file name which is after renaming) is already exists. "+
				"Renaming may erase the contents of the file. So, nothing to do", file)
		}
	}
	return nil
}

func makeDirIfNeeded(filePath string) error {
	dirPath := filepath.Dir(filePath)
	if exists(dirPath) {
		return nil
	}
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return fmt.Errorf("can't make %s directory", dirPath)
	}
	return nil
}

// exists reports whether path exists.
func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// isFile reports whether path exists and is a regular file.
func isFile(path string) bool {
	stat, err := os.Stat(path)
	return err == nil && !stat.IsDir()
}

// isHiddenFile reports whether path is a file whose name starts with a dot.
func isHiddenFile(path string) bool {
	_, file := filepath.Split(path)
	return isFile(path) && strings.HasPrefix(file, ".")
}

// baseNameWithoutExt returns the file name without its directory or extension.
func baseNameWithoutExt(path string) string {
	_, file := filepath.Split(path)
	ext := filepath.Ext(path)
	if ext == "" {
		return file
	}
	return file[:len(file)-len(ext)]
}
