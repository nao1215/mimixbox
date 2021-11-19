//
// mimixbox/internal/lib/path.go
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
	"os"
	"path/filepath"
	"strings"
)

// IsSamePath return true if src and dest is same path,
// false if src and dest is not same path or error occur
func IsSamePath(src string, dest string) bool {
	s, err := filepath.Abs(src)
	if err != nil {
		return false
	}

	d, err := filepath.Abs(dest)
	if err != nil {
		return false
	}
	return s == d
}

// TopDirName return top directory name from path.
func TopDirName(path string) string {
	index := strings.Index(path, string(os.PathSeparator))
	if index == -1 {
		return path
	}

	byteList := []byte(path)
	return string(byteList[:index])
}
