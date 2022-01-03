//
// mimixbox/internal/applets/textutils/tr/tr.go
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
package tr

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/jessevdk/go-flags"
	mb "github.com/nao1215/mimixbox/internal/lib"
)

const cmdName string = "tr"
const version = "1.0.0"

var osExit = os.Exit

type options struct {
	Delete  string `short:"d" long:"delete" description:"delete characters in option argument. Do not translate"`
	Version bool   `short:"v" long:"version" description:"Show tr command version"`
}

type convert struct {
	before string
	after  string
}

func Run() (int, error) {
	var opts options
	var err error
	var args []string

	if args, err = parseArgs(&opts); err != nil {
		return mb.ExitSuccess, nil
	}

	return tr(args, opts)
}

func tr(args []string, opts options) (int, error) {
	if !mb.HasPipeData() {
		return interactiveTranslate(args, opts)
	}
	return fileterTranslate(args, opts)
}

func fileterTranslate(args []string, opts options) (int, error) {
	cnv, err := parseConvertCharSet(args, opts)
	if err != nil {
		return mb.ExitFailure, err
	}

	data := args[len(args)-1]

	if opts.Delete != "" {
		fmt.Fprintln(os.Stdout, strings.TrimRight(delete(data, opts.Delete), "\n"))
	} else {
		fmt.Fprintln(os.Stdout, strings.TrimRight(translate(data, cnv), "\n"))
	}
	return mb.ExitSuccess, nil
}

func interactiveTranslate(args []string, opts options) (int, error) {
	cnv, err := parseConvertCharSet(args, opts)
	if err != nil {
		return mb.ExitFailure, err
	}

	for {
		input, ok := mb.Input()
		if !ok {
			break
		}

		if opts.Delete != "" {
			fmt.Fprintln(os.Stdout, strings.TrimRight(delete(input, opts.Delete), "\n"))
		} else {
			fmt.Fprintln(os.Stdout, strings.TrimRight(translate(input, cnv), "\n"))
		}
	}
	return mb.ExitSuccess, nil
}

func delete(s string, del string) string {
	before := strings.Split(del, "")
	len := len(before)

	replace := s
	for i := 0; i < len; i++ {
		replace = strings.Replace(replace, before[i], "", -1)
	}
	return replace
}

func translate(s string, cnv convert) string {
	before := strings.Split(cnv.before, "")
	after := strings.Split(cnv.after, "")
	len := len(before)

	src := strings.Split(s, "")
	replace := ""
	for _, v := range src {
		org := v
		for i := 0; i < len; i++ {
			v = strings.Replace(v, before[i], after[i], -1)
			if v != org {
				break
			}
		}
		replace = replace + v
	}
	return replace
}

func parseConvertCharSet(args []string, opts options) (convert, error) {
	var cnv convert = convert{}

	if opts.Delete != "" {
		if len(args) > 1 && mb.HasPipeData() {
			return cnv, errors.New("extra operand " + mb.WithSingleCoat(args[1]))
		} else if len(args) > 0 && !mb.HasPipeData() {
			return cnv, errors.New("extra operand " + mb.WithSingleCoat(args[0]))
		}
		return cnv, nil
	} else {
		if len(args) == 0 || (mb.HasPipeData() && len(args) == 1) {
			return cnv, errors.New("missing operand")
		}
	}

	if mb.HasPipeData() { // expect: 0=before, 1=after, 2=data from pipe
		if len(args) == 2 {
			return cnv, errors.New("missing operand after " + mb.WithSingleCoat(args[0]))
		} else if len(args) >= 4 {
			return cnv, errors.New("extra operand " + mb.WithSingleCoat(args[3]))
		}
	} else {
		if len(args) == 1 { // expect: 0=before, 1=after
			return cnv, errors.New("missing operand after " + mb.WithSingleCoat(args[0]))
		} else if len(args) >= 3 {
			return cnv, errors.New("extra operand " + mb.WithSingleCoat(args[2]))
		}
	}

	if len(args[0]) != len(args[1]) {
		return cnv, errors.New(mb.WithSingleCoat(args[0]) + " and " + mb.BaseNameWithoutExt(args[1]) +
			" must be same length")
	}
	cnv.before = args[0]
	cnv.after = args[1]
	return cnv, nil
}

func parseArgs(opts *options) ([]string, error) {
	p := initParser(opts)

	args, err := p.Parse()
	if err != nil {
		return nil, err
	}

	if opts.Version {
		mb.ShowVersion(cmdName, version)
		osExit(mb.ExitSuccess)
	}

	if mb.HasPipeData() {
		stdin, err := mb.FromPIPE()
		if err != nil {
			return nil, err
		}
		if strings.Count(stdin, "\n") == 1 {
			stdin = strings.Trim(stdin, "\n")
		}
		args = append(args, stdin)
	}
	return args, nil
}

func initParser(opts *options) *flags.Parser {
	parser := flags.NewParser(opts, flags.Default)
	parser.Name = cmdName
	parser.Usage = "[OPTIONS] STR_SET [STR_SET2]"

	return parser
}
