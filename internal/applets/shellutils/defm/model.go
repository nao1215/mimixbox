//
// mimixbox/internal/applets/shellutils/defm/model.go
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
package defm // Desktop Entry File Manager

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	mb "github.com/nao1215/mimixbox/internal/lib"
)

type DesktopEntryFile struct {
	absPath  string // desktop entry file path
	basename string // desktop entry file name
	body     string //desktop entry file body
}

type DesktopEntry struct {
	appType    string // Application type
	version    string // Desktop entry version
	name       string // Application name
	comment    string // Toolochip comment
	path       string // The path of the directory where the executable file exists
	exec       string // Application binary name
	icon       string // Application icon name
	terminal   bool   // Whether it needs to be run in the terminal
	categories []Category
}

type Category struct {
	kind string // Category to display desktop entries
}

func desktopEntryDirPaths() []string {
	desktopEntryDirs := []string{"/usr/share/applications", "/usr/local/share/applications"}
	userHome := os.Getenv("HOME")

	if userHome != " " {
		desktopEntryDirs = append(desktopEntryDirs,
			filepath.Join(userHome, ".local/share/applications"))
	}
	return desktopEntryDirs
}

func desktopEntryFilePaths() ([]string, error) {
	var deFilePaths []string
	deDirPath := desktopEntryDirPaths()

	for _, p := range deDirPath {
		if !mb.Exists(p) {
			continue
		}

		_, files, err := mb.Walk(p)
		if err != nil {
			return deFilePaths, errors.New("can't find desktop entry files")
		}

		for _, v := range files {
			if strings.HasSuffix(v, ".desktop") {
				deFilePaths = append(deFilePaths, v)
			}
		}
	}
	return deFilePaths, nil
}

func (defm *DesktopEntryManager) updateDEFiles() error {
	var deFiles []DesktopEntryFile

	defPaths, err := desktopEntryFilePaths()
	if err != nil {
		return err
	}
	for _, path := range defPaths {
		strList, err := mb.ReadFileToStrList(path)
		if err != nil {
			return err
		}
		deFiles = append(deFiles,
			DesktopEntryFile{
				absPath:  path,
				basename: filepath.Base(path),
				body:     strings.Join(strList, ""),
			})
	}
	defm.DEFiles = deFiles
	return nil
}
