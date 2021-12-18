//
// mimixbox/internal/lib/shadow.go
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
)

const ShellsFilePath = "/etc/shells"

func TmpShellsFile() string {
	return ShellsFilePath + ".tmp"
}

func IsRootUser() bool {
	return os.Geteuid() == 0 && os.Getuid() == 0
}

/* TODO: See chsh command.
func HasPamModule() bool {
	return IsFile("/etc/pam.conf")
}

func AuthByPasswordWithPam(userName string) error {
	// Stopping logging is not security measures.
	// There is a problem that the log during authentication cannot be stopped,
	// and the log is temporarily disabled in the workaround.
	// log.SetOutput(ioutil.Discard)

	fmt.Fprintf(os.Stdout, "Enter password: ")
	passwd, err := terminal.ReadPassword(syscall.Stdin)
	if err != nil {
		return err
	}
	fmt.Fprintln(os.Stdout, "")

	reader := bytes.NewReader(passwd)
	key, err := crypto.NewKeyFromReader(reader)
	if err != nil {
		return err
	}
	defer key.Wipe()

	err = pam.IsUserLoginToken(userName, key, true)
	return err
}
*/
