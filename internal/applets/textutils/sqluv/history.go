//
// mimixbox/internal/applets/textutils/sqluv/history.go
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

package sqluv

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// historyPath returns the file the query history should be appended to. An
// explicit path wins; otherwise a deterministic temp-directory path is used so
// the applet never writes to the real home directory unless asked. The original
// sqluv stored history under the user's config dir, but for a MimixBox applet
// that must stay hermetic in CI, defaulting to a temp file is the safe choice.
func historyPath(explicit string) string {
	if explicit != "" {
		return explicit
	}
	return filepath.Join(os.TempDir(), "sqluv-history.log")
}

// recordHistory appends a single timestamped query line to the history file. A
// missing parent directory is created. Errors are returned so the caller can
// decide how loud to be (the headless path treats them as warnings).
func recordHistory(explicit, query string) error {
	path := historyPath(explicit)
	if dir := filepath.Dir(path); dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil { //nolint:gosec // history dir is user-visible by design
			return err
		}
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	line := fmt.Sprintf("%s\t%s\n", time.Now().UTC().Format(time.RFC3339), strings.ReplaceAll(query, "\n", " "))
	_, err = f.WriteString(line)
	return err
}
