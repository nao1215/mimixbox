// mimixbox/internal/lib/lib_more_test.go
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
	"crypto/md5"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReplaceAll(t *testing.T) {
	t.Parallel()
	got := ReplaceAll([]string{"foo bar", "baz foo"}, "foo", "X")
	if strings.Join(got, "|") != "X bar|baz X" {
		t.Errorf("ReplaceAll = %v", got)
	}
}

func TestRemove(t *testing.T) {
	t.Parallel()
	got := Remove([]string{"a", "b", "a", "c"}, "a")
	if strings.Join(got, "") != "bc" {
		t.Errorf("Remove = %v", got)
	}
}

func TestAddLineFeed(t *testing.T) {
	t.Parallel()
	got := AddLineFeed([]string{"a", "b"})
	if got[0] != "a\n" || got[1] != "b\n" {
		t.Errorf("AddLineFeed = %v", got)
	}
}

func TestListDigit(t *testing.T) {
	t.Parallel()
	// 5 elements -> len("5")=1 -> "1".
	if got := ListDigit(make([]string, 5)); got != "1" {
		t.Errorf("ListDigit(5) = %q, want 1", got)
	}
	// 12 elements -> len("12")=2 -> "2".
	if got := ListDigit(make([]string, 12)); got != "2" {
		t.Errorf("ListDigit(12) = %q, want 2", got)
	}
}

func TestWithSingleCoat(t *testing.T) {
	t.Parallel()
	if got := WithSingleCoat("x"); got != "'x'" {
		t.Errorf("WithSingleCoat = %q", got)
	}
}

func TestContains(t *testing.T) {
	t.Parallel()
	if !Contains([]string{"a", "b"}, "b") {
		t.Error("Contains should find b")
	}
	if Contains([]string{"a", "b"}, "z") {
		t.Error("Contains should not find z")
	}
	if Contains("not a slice", "x") {
		t.Error("Contains on non-slice should be false")
	}
}

func TestSignalHelpers(t *testing.T) {
	t.Parallel()
	if !IsSignalNumber("9") {
		t.Error("9 is a signal number")
	}
	if IsSignalNumber("999") {
		t.Error("999 is not a signal number")
	}
	if !IsSignalName("SIGKILL") || !IsSignalName("KILL") {
		t.Error("SIGKILL/KILL should be recognized")
	}
	if IsSignalName("NOPE") {
		t.Error("NOPE is not a signal name")
	}
	if SignalAtoi("15") != 15 {
		t.Error("SignalAtoi(15) should be 15")
	}
	if SignalAtoi("xx") != -1 {
		t.Error("SignalAtoi(xx) should be -1")
	}
	if ConvSignalNameToNum("KILL") != 9 {
		t.Error("ConvSignalNameToNum(KILL) should be 9")
	}
	if ConvSignalNameToNum("NOPE") != -1 {
		t.Error("ConvSignalNameToNum(NOPE) should be -1")
	}
	// Smoke-test the printers (they write to stdout).
	PrintSignalList()
	PrintSignal("KILL")
	PrintSignal("9")
	PrintSignal("HUP")
}

func TestOptionHelpers(t *testing.T) {
	t.Parallel()
	if !HasVersionOpt([]string{"cmd", "--version"}) || !HasVersionOpt([]string{"cmd", "-v"}) {
		t.Error("HasVersionOpt should detect version flags")
	}
	if HasVersionOpt([]string{"cmd", "foo"}) {
		t.Error("HasVersionOpt false positive")
	}
	if !HasHelpOpt([]string{"cmd", "--help"}) || !HasHelpOpt([]string{"cmd", "-h"}) {
		t.Error("HasHelpOpt should detect help flags")
	}
	if HasHelpOpt([]string{"cmd", "foo"}) {
		t.Error("HasHelpOpt false positive")
	}
}

func TestSimpleBackupSuffix(t *testing.T) {
	t.Setenv("SIMPLE_BACKUP_SUFFIX", "")
	if got := SimpleBackupSuffix(); got != "~" {
		t.Errorf("default suffix = %q, want ~", got)
	}
	t.Setenv("SIMPLE_BACKUP_SUFFIX", ".bak")
	if got := SimpleBackupSuffix(); got != ".bak" {
		t.Errorf("env suffix = %q, want .bak", got)
	}
}

func TestPathHelpers(t *testing.T) {
	t.Parallel()
	if !IsSamePath("/tmp/x", "/tmp/x") {
		t.Error("identical paths should be the same")
	}
	if IsSamePath("/tmp/x", "/tmp/y") {
		t.Error("different paths should not be the same")
	}
	if got := TopDirName("usr/local/bin"); got != "usr" {
		t.Errorf("TopDirName = %q, want usr", got)
	}
	if got := TopDirName("nofslash"); got != "nofslash" {
		t.Errorf("TopDirName(no slash) = %q", got)
	}
}

func TestShellPureHelpers(t *testing.T) {
	t.Parallel()
	if WrapString("abcdef", 2) != "ab\ncd\nef" {
		t.Errorf("WrapString = %q", WrapString("abcdef", 2))
	}
	if WrapString("abc", 0) != "abc" {
		t.Error("WrapString with column 0 returns src")
	}
	if Chop("line\n") != "line" {
		t.Error("Chop should trim trailing newline")
	}
	if Chop("noeol") != "noeol" {
		t.Error("Chop without newline returns input")
	}
	got := ChopAll([]string{"a\n", "b"})
	if got[0] != "a" || got[1] != "b" {
		t.Errorf("ChopAll = %v", got)
	}
	if !IsRootDir("/") {
		t.Error("/ should be the root dir")
	}
	if IsRootDir("/tmp") {
		t.Error("/tmp is not root")
	}
	if !ExistCmd("go") {
		t.Error("go should be on PATH in the test environment")
	}
	if ExistCmd("definitely-not-a-real-command-xyz") {
		t.Error("nonexistent command should not be found")
	}
}

func TestHasOperand(t *testing.T) {
	t.Parallel()
	if !HasOperand([]string{"cat", "file.txt"}, "cat") {
		t.Error("file.txt is an operand")
	}
	if HasOperand([]string{"cat", "-n"}, "cat") {
		t.Error("-n is an option, not an operand")
	}
	if !HasNoOperand([]string{"cat", "-n"}, "cat") {
		t.Error("HasNoOperand should be true when only options are present")
	}
}

func TestPrintNumberLines(t *testing.T) {
	t.Parallel()
	// Smoke-test the stdout printers for coverage.
	PrintStrWithNumberLine(1, "%6d  %s", "hello")
	PrintStrListWithNumberLine([]string{"a", "\n", "b"}, false)
	PrintStrListWithNumberLine([]string{"a", "b"}, true)
	Dump([]string{"x\n"}, false)
	Dump([]string{"x"}, true)
}

func TestConcatenate(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	f1 := filepath.Join(dir, "a.txt")
	f2 := filepath.Join(dir, "b.txt")
	if err := os.WriteFile(f1, []byte("hello\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(f2, []byte("world\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	lines, err := Concatenate([]string{f1, f2})
	if err != nil {
		t.Fatalf("Concatenate error = %v", err)
	}
	if strings.Join(lines, "") != "hello\nworld\n" {
		t.Errorf("Concatenate = %q", strings.Join(lines, ""))
	}
	if _, err := Concatenate([]string{"/no/such/file"}); err == nil {
		t.Error("Concatenate should error on a missing file")
	}
}

func TestChecksumHelpers(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	data := filepath.Join(dir, "data.txt")
	if err := os.WriteFile(data, []byte("hello"), 0o600); err != nil {
		t.Fatal(err)
	}
	sum, err := CalcChecksum(md5.New(), data)
	if err != nil {
		t.Fatalf("CalcChecksum error = %v", err)
	}
	// md5("hello") = 5d41402abc4b2a76b9719d911017c592
	if sum != "5d41402abc4b2a76b9719d911017c592" {
		t.Errorf("CalcChecksum = %q", sum)
	}
	if _, err := CalcChecksum(md5.New(), "/no/such/file"); err == nil {
		t.Error("CalcChecksum should error on a missing file")
	}

	// PrintChecksums over a present and an absent path.
	status, err := PrintChecksums("md5sum", md5.New(), []string{data})
	if err != nil || status != 0 {
		t.Errorf("PrintChecksums = %d, %v", status, err)
	}
	status, _ = PrintChecksums("md5sum", md5.New(), []string{"/no/such/file"})
	if status != 1 {
		t.Errorf("PrintChecksums missing-file status = %d, want 1", status)
	}

	// CompareChecksum against a valid checksum file.
	sumFile := filepath.Join(dir, "sums.txt")
	if err := os.WriteFile(sumFile, []byte(sum+"  "+data+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := CompareChecksum(md5.New(), []string{sumFile}); err != nil {
		t.Errorf("CompareChecksum error = %v", err)
	}
}

func TestIp4(t *testing.T) {
	t.Parallel()
	// Just exercise it; the result is host-specific and may be empty.
	if _, err := Ip4(); err != nil {
		t.Errorf("Ip4 error = %v", err)
	}
}

func TestLookupIDs(t *testing.T) {
	t.Parallel()
	// root / gid 0 exist on every Linux system.
	if uid, err := LookupUid("0"); err != nil || uid != 0 {
		t.Errorf("LookupUid(0) = %d, %v", uid, err)
	}
	if gid, err := LookupGid("0"); err != nil || gid != 0 {
		t.Errorf("LookupGid(0) = %d, %v", gid, err)
	}
	if _, err := LookupUid("no-such-user-xyz"); err == nil {
		t.Error("LookupUid of a bogus user should error")
	}
	if _, err := LookupGid("no-such-group-xyz"); err == nil {
		t.Error("LookupGid of a bogus group should error")
	}
}

func TestShadowHelpers(t *testing.T) {
	t.Parallel()
	if ShellsFilePath != "/etc/shells" {
		t.Errorf("ShellsFilePath = %q", ShellsFilePath)
	}
	if TmpShellsFile() != "/etc/shells.tmp" {
		t.Errorf("TmpShellsFile = %q", TmpShellsFile())
	}
	// IsRootUser just needs to run without panicking; its value depends on who
	// runs the test.
	_ = IsRootUser()
}

func TestShowVersion(t *testing.T) {
	t.Parallel()
	// Smoke-test; writes to stdout.
	ShowVersion("mimixbox", "1.2.3")
	if ExitSuccess != 0 || ExitFailure != 1 {
		t.Error("exit code constants are wrong")
	}
}

func TestGroups(t *testing.T) {
	t.Parallel()
	// Look up the current user's groups; should not error for a real user.
	u := os.Getenv("USER")
	if u == "" {
		t.Skip("no USER in environment")
	}
	groups, err := Groups(u)
	if err != nil {
		t.Skipf("Groups(%s) error = %v (user may not be in the user database)", u, err)
	}
	DumpGroups(groups, true)
	DumpGroups(groups, false)
}
