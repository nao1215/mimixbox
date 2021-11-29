//
// mimixbox/internal/lib/file_test.go
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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsFile(t *testing.T) {
	assert.Equal(t, true, IsFile("/tmp/mimixbox/ut/Readable.txt"))
	assert.Equal(t, true, IsFile("/tmp/mimixbox/ut/symbolic.txt"))
	assert.Equal(t, false, IsFile("/tmp/mimixbox/ut"))
	assert.Equal(t, true, IsFile("/tmp/mimixbox/ut/AllZero.txt"))
	assert.Equal(t, false, IsFile("/tmp/mimixbox/ut/NoReadableDir"))
	assert.Equal(t, true, IsFile("/tmp/mimixbox/ut/.hidden.txt"))
	assert.Equal(t, false, IsFile("abcdef"))
}

func TestExists(t *testing.T) {
	assert.Equal(t, true, Exists("/tmp/mimixbox/ut/Readable.txt"))
	assert.Equal(t, true, Exists("/tmp/mimixbox/ut/symbolic.txt"))
	assert.Equal(t, true, Exists("/tmp/mimixbox/ut"))
	assert.Equal(t, true, Exists("/tmp/mimixbox/ut/AllZero.txt"))
	assert.Equal(t, true, Exists("/"))
	assert.Equal(t, true, Exists("/tmp/mimixbox"))
	assert.Equal(t, false, Exists("abcdef"))
}

func TestIsDir(t *testing.T) {
	assert.Equal(t, false, IsDir("/tmp/mimixbox/ut/Readable.txt"))
	assert.Equal(t, false, IsDir("/tmp/mimixbox/ut/symbolic.txt"))
	assert.Equal(t, true, IsDir("/tmp/mimixbox/ut"))
	assert.Equal(t, false, IsDir("/tmp/mimixbox/ut/AllZero.txt"))
	assert.Equal(t, true, IsDir("/"))
	assert.Equal(t, true, IsDir("/tmp/mimixbox"))
	assert.Equal(t, false, IsDir("abcdef"))
	assert.Equal(t, true, IsDir("/tmp/mimixbox/ut/NoWritableDir"))
}

func TestIsSymlink(t *testing.T) {
	assert.Equal(t, false, IsSymlink("/tmp/mimixbox/ut/Readable.txt"))
	assert.Equal(t, true, IsSymlink("/tmp/mimixbox/ut/symbolic.txt"))
	assert.Equal(t, false, IsSymlink("/tmp/mimixbox/ut"))
	assert.Equal(t, false, IsSymlink("/tmp/mimixbox/ut/AllZero.txt"))
	assert.Equal(t, false, IsSymlink("/"))
	assert.Equal(t, false, IsSymlink("/tmp/mimixbox"))
	assert.Equal(t, false, IsSymlink("abcdef"))
}

func TestIsZero(t *testing.T) {
	assert.Equal(t, true, IsZero("/tmp/mimixbox/ut/Readable.txt"))
	assert.Equal(t, true, IsZero("/tmp/mimixbox/ut/symbolic.txt"))
	assert.Equal(t, false, IsZero("/tmp/mimixbox/ut"))
	assert.Equal(t, false, IsZero("abcdef"))
	assert.Equal(t, true, IsZero("/tmp/mimixbox/ut/AllZero.txt"))
}

func TestIsReadable(t *testing.T) {
	assert.Equal(t, true, IsReadable("/tmp/mimixbox/ut/Readable.txt"))
	assert.Equal(t, true, IsReadable("/tmp/mimixbox/ut/symbolic.txt"))
	assert.Equal(t, true, IsReadable("/tmp/mimixbox/ut"))
	assert.Equal(t, true, IsReadable("/tmp/mimixbox"))
	assert.Equal(t, false, IsReadable("/tmp/mimixbox/ut/NonReadable.txt"))
	assert.Equal(t, false, IsReadable("abcdef"))
	assert.Equal(t, false, IsReadable("/tmp/mimixbox/ut/AllZero.txt"))
}

func TestIsWritable(t *testing.T) {
	assert.Equal(t, true, IsWritable("/tmp/mimixbox/ut/Writable.txt"))
	assert.Equal(t, true, IsWritable("/tmp/mimixbox/ut/symbolic.txt"))
	assert.Equal(t, true, IsWritable("/tmp/mimixbox/ut"))
	assert.Equal(t, true, IsWritable("/tmp/mimixbox"))
	assert.Equal(t, false, IsWritable("/tmp/mimixbox/ut/NonWritable.txt"))
	assert.Equal(t, false, IsWritable("abcdef"))
	assert.Equal(t, false, IsWritable("/tmp/mimixbox/ut/AllZero.txt"))
}
func TestIsExecutable(t *testing.T) {
	assert.Equal(t, true, IsExecutable("/tmp/mimixbox/ut/Executable.txt"))
	assert.Equal(t, true, IsExecutable("/tmp/mimixbox/ut/symbolic.txt"))
	assert.Equal(t, true, IsExecutable("/tmp/mimixbox/ut"))
	assert.Equal(t, true, IsExecutable("/tmp/mimixbox"))
	assert.Equal(t, false, IsExecutable("/tmp/mimixbox/ut/NonExecutable.txt"))
	assert.Equal(t, false, IsExecutable("abcdef"))
	assert.Equal(t, false, IsExecutable("/tmp/mimixbox/ut/AllZero.txt"))
}

func TestIsHiddenFile(t *testing.T) {
	assert.Equal(t, false, IsHiddenFile("/tmp/mimixbox/ut/Executable.txt"))
	assert.Equal(t, true, IsHiddenFile("/tmp/mimixbox/ut/.hidden.txt"))
	assert.Equal(t, false, IsHiddenFile("/tmp/mimixbox"))
	assert.Equal(t, false, IsHiddenFile("/tmp/mimixbox/ut/"))
	assert.Equal(t, false, IsHiddenFile("/tmp/mimixbox/ut"))
	assert.Equal(t, false, IsHiddenFile("abcdef"))
	assert.Equal(t, false, IsHiddenFile(".abcdef"))
	assert.Equal(t, false, IsHiddenFile(".HiddenDir"))
}

func TestBaseNameWithoutExt(t *testing.T) {
	assert.Equal(t, "Executable", BaseNameWithoutExt("/tmp/mimixbox/ut/Executable.txt"))
	assert.Equal(t, ".hidden", BaseNameWithoutExt("/tmp/mimixbox/ut/.hidden.txt"))
	assert.Equal(t, "file", BaseNameWithoutExt("./file.go"))
	assert.Equal(t, "mimixbox", BaseNameWithoutExt("/tmp/mimixbox"))
	assert.Equal(t, "", BaseNameWithoutExt("/tmp/mimixbox/ut/"))
	assert.Equal(t, "ut", BaseNameWithoutExt("/tmp/mimixbox/ut"))
	assert.Equal(t, "abcdef", BaseNameWithoutExt("abcdef"))
	assert.Equal(t, "", BaseNameWithoutExt(".HiddenDir"))
}
