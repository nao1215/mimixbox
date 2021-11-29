//
// mimixbox/internal/lib/string.go
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
	"strings"
)

func ReplaceAll(lines []string, target string, after string) []string {
	var replacedLines []string
	for _, line := range lines {
		l := strings.ReplaceAll(line, target, after)
		replacedLines = append(replacedLines, l)
	}
	return replacedLines
}

func Remove(strings []string, target string) []string {
	result := []string{}
	for _, v := range strings {
		if v != target {
			result = append(result, v)
		}
	}
	return result
}

func AddLineFeed(lines []string) []string {
	var newLines []string
	for _, v := range lines {
		newLines = append(newLines, v+"\n")
	}
	return newLines
}
