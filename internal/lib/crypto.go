// mimixbox/internal/lib/crypto.go
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
	"fmt"
	"hash"
	"io"
	"os"
	"strings"
)

// CompareChecksum verifies the files listed in each checksum file against the
// digest produced by h, writing a "<file>: OK" / "<file>: Fail" line per entry
// to out. out is injected so the result can be captured in tests instead of
// leaking to the process stdout.
//
// e.g. hash is md5.New(), md5.New(), sha512.New(), etc...
func CompareChecksum(out io.Writer, h hash.Hash, paths []string) error {
	for _, path := range paths {
		f, err := os.Open(os.ExpandEnv(path))
		if err != nil {
			return err
		}
		defer f.Close()

		reader := bufio.NewReaderSize(f, 4096)
		for {
			line, _, err := reader.ReadLine()
			if err == io.EOF {
				break
			}

			data := strings.Split(string(line), "  ")
			if len(data) != 2 {
				return fmt.Errorf("wrong checksum format")
			}

			r, err := os.Open(data[1])
			if err != nil {
				return err
			}
			defer r.Close()

			s, err := checksum(h, r)
			if err != nil {
				return err
			}
			if s == data[0] {
				fmt.Fprintf(out, "%s: OK\n", data[1])
			} else {
				fmt.Fprintf(out, "%s: Fail\n", data[1])
			}
			h.Reset()
		}
	}
	return nil
}

// ChecksumOutput writes the "<digest>  <path>" line for r to out. out is
// injected so callers (and tests) control the destination instead of the
// process stdout.
func ChecksumOutput(out io.Writer, hash hash.Hash, r io.Reader, path string) error {
	s, err := checksum(hash, r)
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "%s  %s\n", s, path)
	return nil
}

func checksum(hash hash.Hash, fp io.Reader) (string, error) {
	if _, err := io.Copy(hash, fp); err != nil {
		return "", fmt.Errorf("checksum: %s", err.Error())
	}
	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

// PrintChecksums writes "<digest>  <path>" lines for each readable path to out
// and per-path failure diagnostics to errw, returning a non-zero status if any
// path could not be hashed. Both writers are injected so the command output and
// diagnostics can be captured in tests rather than the process std streams.
func PrintChecksums(out io.Writer, errw io.Writer, cmdName string, hash hash.Hash, paths []string) (int, error) {
	status := 0
	for _, path := range paths {
		p := os.ExpandEnv(path)
		if !Exists(p) {
			fmt.Fprint(errw, cmdName+": "+p+": No such file or directory")
			status = 1
			continue
		}

		if IsDir(p) {
			fmt.Fprint(errw, cmdName+": "+p+": It is directory")
			status = 1
			continue
		}

		r, err := os.Open(p)
		if err != nil {
			fmt.Fprint(errw, cmdName+": "+err.Error())
			status = 1
			continue
		}
		defer r.Close()

		if err := ChecksumOutput(out, hash, r, p); err != nil {
			fmt.Fprint(errw, cmdName+": "+err.Error())
			status = 1
			continue
		}
		hash.Reset()
	}
	return status, nil
}

func CalcChecksum(hash hash.Hash, path string) (string, error) {
	r, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer r.Close()

	s, err := checksum(hash, r)
	if err != nil {
		return "", err
	}
	return s, nil
}
