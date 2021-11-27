//
// mimixbox/internal/lib/crypto.go
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
	"fmt"
	"hash"
	"io"
	"os"
	"strings"
)

// e.g. hash is md5.New(), md5.New(), sha512.New(), etc...
func CompareChecksum(h hash.Hash, paths []string) error {
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
				fmt.Fprintf(os.Stdout, "%s: OK\n", data[1])
			} else {
				fmt.Fprintf(os.Stdout, "%s: Fail\n", data[1])
			}
			h.Reset()
		}
	}
	return nil
}

func ChecksumOutput(hash hash.Hash, r io.Reader, path string) error {
	s, err := checksum(hash, r)
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stdout, "%s  %s\n", s, path)
	return nil
}

func checksum(hash hash.Hash, fp io.Reader) (string, error) {
	if _, err := io.Copy(hash, fp); err != nil {
		return "", fmt.Errorf("checksum: %s", err.Error())
	}
	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}
