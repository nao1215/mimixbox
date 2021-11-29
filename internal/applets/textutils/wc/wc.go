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
	"strings"

	mb "github.com/nao1215/mimixbox/internal/lib"

	"github.com/jessevdk/go-flags"
)

const cmdName string = "wc"
const version = "1.0.0"

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
		_, err := wc(strings.Split(args[0], ""), "-", opts)
		if err != nil {
			return ExitFailuer, nil
		}
		return ExitSuccess, nil
	}

	if len(args) == 0 || mb.Contains(args, "-") {
		var data []string
		for {
			input, next := mb.Input()
			if !next {
				break
			}
			if input != "" {
				data = append(data, input)
			}
		}
		_, err := wc(data, "-", opts)
		if err != nil {
			return ExitFailuer, nil
		}
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

	if len(results) >= 2 {
		printWordCountData(sum(results), "Total", opts)
	}

	return status, nil
}

func wc(lines []string, path string, opts options) (wordCount, error) {
	// Gnu Coreutils wc does not count the first line as the number of lines.
	// So, initial value of line number is -1.
	var result wordCount = wordCount{-1, 0, 0, 0}

	for _, line := range lines {
		result.lines++
		result.words += countWord(line)
		result.bytes += len([]byte(line))
	}
	printWordCountData(result, path, opts)
	return result, nil
}

func printWordCountData(wc wordCount, path string, opts options) {
	//TODO: Not supported when multiple options are specified
	if opts.Bytes {
		fmt.Fprintf(os.Stdout, "%d %s\n", wc.bytes, path)
		return
	} else if opts.Lines {
		fmt.Fprintf(os.Stdout, "%d %s\n", wc.lines, path)
		return
	} else if opts.MaxLineLen {
		fmt.Fprintf(os.Stdout, "%d %s\n", wc.maxLength, path)
		return
	}
	fmt.Fprintf(os.Stdout, "%d %d %d %s\n", wc.lines, wc.words, wc.bytes, path)
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

func sum(wordCounts []wordCount) wordCount {
	var total = wordCount{0, 0, 0, 0}
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
