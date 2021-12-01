//
// mimixbox/internal/lib/file.go
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

// Wark return 1）directory List, 2) file list, 3) error
func Walk(dir string) ([]string, []string, error) {
	fileList := []string{}
	dirList := []string{}

	err := filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
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

// Copy file(src) to dest
func Copy(src string, dest string) error {
	s, err := os.Open(src)
	if err != nil {
		return err
	}
	defer s.Close()

	d, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer d.Close()

	_, err = io.Copy(d, s)
	if err != nil {
		return err
	}
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
	dirs, files, err := Walk(dir)
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
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		strList = append(strList, scanner.Text()+"\n")
	}

	if len(strList) >= 1 {
		strList[len(strList)-1] = strings.TrimRight(strList[len(strList)-1], "\n")
	}
	return strList, nil
}

func ListToFile(filepath string, lines []string) error {
	fp, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer fp.Close()

	writer := bufio.NewWriter(fp)
	for _, line := range lines {
		writer.WriteString(line)
	}
	return writer.Flush()
}
