// mimixbox/internal/lib/file.go
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
	"bufio"
	"errors"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

const (
	// Unknown : Don't use this bits
	Unknown os.FileMode = 1 << (9 - iota)
	// Readable : readable bits
	Readable
	// Writable : writable bits
	Writable
	// Executable : executable bits
	Executable
)

// IsFile reports whether the path exists and is a file.
func IsFile(path string) bool {
	stat, err := os.Stat(path)
	return (err == nil) && (!stat.IsDir())
}

// Exists reports whether the path exists.
func Exists(path string) bool {
	_, err := os.Stat(path)
	return (err == nil)
}

// IsDir reports whether the path exists and is a directory.
func IsDir(path string) bool {
	stat, err := os.Stat(path)
	return (err == nil) && (stat.IsDir())
}

// IsSymlink reports whether the path exists and is a symbolic link.
func IsSymlink(path string) bool {
	stat, err := os.Lstat(path)
	if err != nil {
		return false
	}
	if stat.Mode()&os.ModeSymlink == os.ModeSymlink {
		return true
	}
	return false
}

// IsZero reports whether the path exists and is zero size.
func IsZero(path string) bool {
	stat, err := os.Stat(path)
	return (err == nil) && (stat.Size() == 0)
}

// IsReadable reports whether the path exists and is readable.
func IsReadable(path string) bool {
	stat, err := os.Stat(path)
	return (err == nil) && ((stat.Mode() & Readable) != 0)
}

// IsWritable reports whether the path exists and is writable.
func IsWritable(path string) bool {
	stat, err := os.Stat(path)
	return (err == nil) && ((stat.Mode() & Writable) != 0)
}

// IsExecutable reports whether the path exists and is executable.
func IsExecutable(path string) bool {
	stat, err := os.Stat(path)
	return (err == nil) && ((stat.Mode() & Executable) != 0)
}

// IsHiddenFile reports whether the path exists and is included hidden file.
func IsHiddenFile(filePath string) bool {
	_, file := path.Split(filePath)
	if IsFile(filePath) && strings.HasPrefix(file, ".") {
		return true
	}
	return false
}

// BaseNameWithoutExt return file name without extension.
func BaseNameWithoutExt(path string) string {
	_, file := filepath.Split(path)
	if filepath.Ext(path) == "" {
		return file
	}
	return file[:len(file)-len(filepath.Ext(path))]
}

func IsNamedPipe(path string) bool {
	stat, err := os.Stat(path)
	return (err == nil) && ((stat.Mode() & fs.ModeNamedPipe) != 0)
}

// Wark return 1）directory List, 2) file list, 3) error
func Walk(dir string, ignoreErr bool) ([]string, []string, error) {
	fileList := []string{}
	dirList := []string{}

	err := filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		if err != nil && !ignoreErr {
			return err
		}

		if IsDir(path) {
			dirList = append(dirList, path)
		} else {
			fileList = append(fileList, path)
		}
		return nil
	})
	return dirList, fileList, err
}

// IsSameFileName return true if src and dest is same name,
// false if src and dest is not same name
func IsSameFileName(src string, dest string) bool {
	return filepath.Base(src) == filepath.Base(dest)
}

// Copy copies the regular file src to dest, preserving the source's permission
// bits and modification time. Preserving the mode matters for callers such as
// mv's cross-filesystem fallback, where a plain os.Create would otherwise drop
// the execute bit and reset timestamps.
func Copy(src string, dest string) error {
	info, err := os.Lstat(src)
	if err != nil {
		return err
	}

	s, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = s.Close() }()

	d, err := os.OpenFile(dest, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, info.Mode().Perm())
	if err != nil {
		return err
	}

	if _, err := io.Copy(d, s); err != nil {
		_ = d.Close()
		return err
	}
	if err := d.Close(); err != nil {
		return err
	}

	// Restore the exact mode (umask may have masked it on creation) and mtime.
	_ = os.Chmod(dest, info.Mode().Perm())
	_ = os.Chtimes(dest, info.ModTime(), info.ModTime())
	return nil
}

// CopyTree recursively copies the file or directory tree rooted at src to dest,
// preserving each entry's permission bits and modification times. It is used by
// mv when a rename crosses a filesystem boundary (os.Rename returns EXDEV), a
// case a single-file Copy cannot handle.
func CopyTree(src string, dest string) error {
	info, err := os.Lstat(src)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return Copy(src, dest)
	}

	if err := os.MkdirAll(dest, info.Mode().Perm()); err != nil {
		return err
	}
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if err := CopyTree(filepath.Join(src, e.Name()), filepath.Join(dest, e.Name())); err != nil {
			return err
		}
	}
	_ = os.Chmod(dest, info.Mode().Perm())
	_ = os.Chtimes(dest, info.ModTime(), info.ModTime())
	return nil
}

func RemoveFile(path string, interactive bool) error {
	if !IsFile(path) {
		return errors.New(path + " is not file")
	}
	if interactive && !Question("Remove "+path+"?") {
		return nil // Skip this file
	}
	if err := os.Remove(path); err != nil {
		return err
	}
	return nil
}

func RemoveDir(dir string, interactive bool) error {
	if !interactive {
		if err := os.RemoveAll(dir); err != nil {
			return err
		}
		return nil
	}
	if err := interactiveRemoveDir(dir); err != nil {
		return err
	}
	return nil
}

func interactiveRemoveDir(dir string) error {
	dirs, files, err := Walk(dir, false)
	if err != nil {
		return err
	}

	// Start with the deepest directory or file
	sort.Sort(sort.Reverse(sort.StringSlice(dirs)))
	sort.Sort(sort.Reverse(sort.StringSlice(files)))

	for _, file := range files {
		if !Question("Remove " + file + "?") {
			continue
		}
		err := os.Remove(file)
		if err != nil {
			return err
		}
	}
	for _, dir := range dirs {
		if !Question("Remove " + dir + "?") {
			continue
		}
		err := os.Remove(dir)
		if err != nil {
			return err
		}
	}
	return nil
}

func ReadFileToStrList(path string) ([]string, error) {
	var strList []string
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	r := bufio.NewReader(f)
	for {
		line, err := r.ReadString('\n')
		if err != nil && err != io.EOF {
			return nil, err
		}
		if err == io.EOF && len(line) == 0 {
			break
		}
		strList = append(strList, line)
	}
	return strList, nil
}

// ListToFile writes lines to path atomically: it writes a temporary file in the
// same directory and renames it over path on success. This way an interrupted
// write or a full disk leaves the original file intact instead of a truncated
// one (the danger of a plain os.Create on the target). When path already exists
// its permission bits are preserved.
func ListToFile(path string, lines []string) error {
	mode := os.FileMode(0644)
	if info, err := os.Stat(path); err == nil {
		mode = info.Mode().Perm()
	}

	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, "."+filepath.Base(path)+".tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	// Clean up the temp file if anything below fails before the rename.
	defer func() { _ = os.Remove(tmpName) }()

	writer := bufio.NewWriter(tmp)
	for _, line := range lines {
		if _, err := writer.WriteString(line); err != nil {
			_ = tmp.Close()
			return err
		}
	}
	if err := writer.Flush(); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmpName, mode); err != nil {
		return err
	}
	return os.Rename(tmpName, path)
}

// Return file size(Byte)
func Size(path string) (int64, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return 0, err
	}

	return fileInfo.Size(), nil
}
