//
// mimixbox/internal/applets/textutils/wc/wc.go
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
package wc

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	mb "github.com/nao1215/mimixbox/internal/lib"

	"github.com/jessevdk/go-flags"
)

const cmdName string = "wc"
const version = "1.0.1"

var osExit = os.Exit

type options struct {
	Bytes      bool `short:"c" long:"bytes" description:"Print the byte counts"`
	Lines      bool `short:"l" long:"lines" description:"Print the newline counts"`
	MaxLineLen bool `short:"L" long:"max-line-length" description:"Print the maximum display width"`
	Words      bool `short:"w" long:"words" description:"Print the word counts"`
	Version    bool `short:"v" long:"version" description:"Show wc command version"`
}

type wordCount struct {
	lines     int
	words     int
	bytes     int
	maxLength int
	filePath  string
}

// Exit code
const (
	ExitSuccess int = iota // 0
	ExitFailuer
)

func Run() (int, error) {
	var opts options
	var args []string
	var err error

	if args, err = parseArgs(&opts); err != nil {
		return ExitFailuer, nil
	}

	if mb.HasPipeData() {
		result, err := wc(strings.Split(args[0], ""), "-", opts)
		if err != nil {
			return ExitFailuer, nil
		}
		printWordCountData([]wordCount{result}, opts)
		return ExitSuccess, nil
	}

	if len(args) == 0 || mb.Contains(args, "-") {
		var lines []string
		for {
			input, next := mb.Input()
			if !next {
				break
			}
			if input != "" {
				lines = append(lines, input)
			}
		}
		result, err := wc(lines, "-", opts)
		if err != nil {
			return ExitFailuer, nil
		}
		printWordCountData([]wordCount{result}, opts)
		return ExitSuccess, nil
	}

	return wcAll(args, opts)
}

func wcAll(args []string, opts options) (int, error) {
	status := ExitSuccess
	var results []wordCount
	for _, file := range args {
		target := os.ExpandEnv(file)

		if mb.IsDir(target) {
			fmt.Fprintf(os.Stderr, file+": this path is directory")
			status = ExitFailuer
			continue
		}
		if !mb.IsFile(target) {
			fmt.Fprintf(os.Stderr, file+": no such File")
			status = ExitFailuer
			continue
		}

		lines, err := mb.ReadFileToStrList(target)
		if err != nil {
			fmt.Fprintln(os.Stderr, target+": can't read file")
			status = ExitFailuer
			continue
		}

		result, err := wc(lines, target, opts)
		if err != nil {
			fmt.Fprintln(os.Stderr, target+": can't read file")
			status = ExitFailuer
			continue
		}
		results = append(results, result)
	}

	if len(args) > 1 {
		results = append(results, total(results))
	}
	printWordCountData(results, opts)

	return status, nil
}

func wc(lines []string, path string, opts options) (wordCount, error) {
	var result wordCount = wordCount{0, 0, 0, 0, ""}

	result.filePath = path
	for _, line := range lines {
		result.lines++
		result.words += countWord(line)
		result.bytes += len([]byte(line))
	}
	// In Coreutils, it looks like the terminator is also counted as a Byte count.
	result.bytes = result.bytes + 1
	return result, nil
}

func printWordCountData(counts []wordCount, opts options) {
	digit := maxDigit(counts, opts)
	oneContent := "%" + strconv.Itoa(digit) + "d"
	formatForOneContent := oneContent + " %s\n"
	formatForAll := oneContent + " " + oneContent + " " + oneContent + " %s\n"

	//TODO: Not supported when multiple options are specified.
	for _, v := range counts {
		if opts.Bytes {
			fmt.Fprintf(os.Stdout, formatForOneContent, v.bytes, v.filePath)
			return
		} else if opts.Lines {
			fmt.Fprintf(os.Stdout, formatForOneContent, v.lines, v.filePath)
			return
		} else if opts.MaxLineLen {
			fmt.Fprintf(os.Stdout, formatForOneContent, v.maxLength, v.filePath)
			return
		}
		fmt.Fprintf(os.Stdout, formatForAll, v.lines, v.words, v.bytes, v.filePath)
	}
}

func maxDigit(counts []wordCount, opts options) int {
	var maxDigit int = 0
	for _, v := range counts {
		bytes := len(strconv.Itoa(v.bytes))
		lines := len(strconv.Itoa(v.lines))
		length := len(strconv.Itoa(v.maxLength))
		words := len(strconv.Itoa(v.words))

		if opts.Bytes && maxDigit < bytes {
			maxDigit = bytes
		} else if opts.Lines && maxDigit < lines {
			maxDigit = lines
		} else if opts.MaxLineLen && maxDigit < length {
			maxDigit = length
		} else if opts.Words && maxDigit < words {
			maxDigit = words
		} else {
			if maxDigit < bytes {
				maxDigit = bytes
			}
			if maxDigit < lines {
				maxDigit = lines
			}
			if maxDigit < length {
				maxDigit = length
			}
			if maxDigit < words {
				maxDigit = words
			}
		}
	}
	return maxDigit
}

func countWord(line string) int {
	reg := " |\t"
	split := regexp.MustCompile(reg).Split(line, -1)

	split = mb.Remove(split, "")
	split = mb.Remove(split, " ")
	split = mb.Remove(split, "\t")
	split = mb.Remove(split, "\n")

	return len(split)
}

func total(wordCounts []wordCount) wordCount {
	var total = wordCount{0, 0, 0, 0, "total"}
	for _, w := range wordCounts {
		total.bytes += w.bytes
		total.lines += w.lines
		total.maxLength += w.maxLength
		total.words += w.words
	}
	return total
}

func parseArgs(opts *options) ([]string, error) {
	p := initParser(opts)

	args, err := p.Parse()
	if err != nil {
		return nil, err
	}

	if mb.HasPipeData() {
		stdin, err := mb.FromPIPE()
		if err != nil {
			return nil, err
		}
		return []string{stdin}, nil
	}

	if opts.Version {
		mb.ShowVersion(cmdName, version)
		osExit(ExitSuccess)
	}

	return args, nil
}

func initParser(opts *options) *flags.Parser {
	parser := flags.NewParser(opts, flags.Default)
	parser.Name = cmdName
	parser.Usage = "[OPTIONS] FILE_PATH"

	return parser
}
