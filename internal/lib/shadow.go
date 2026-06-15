// mimixbox/internal/lib/shadow.go
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
	"os"
)

const ShellsFilePath = "/etc/shells"

func TmpShellsFile() string {
	return ShellsFilePath + ".tmp"
}

func IsRootUser() bool {
	return os.Geteuid() == 0 && os.Getuid() == 0
}

// Authentication model: MimixBox does not link against PAM. Applets that
// modify the user databases (for example chsh and chpasswd) authorize the
// caller through filesystem permissions on /etc/passwd and /etc/shadow rather
// than by prompting for a password. On a typical Linux host that means such
// operations require privilege (root or the file owner); on platforms without
// these databases the applets fail explicitly. The previously commented-out
// PAM helper was removed because that path was never implemented.
