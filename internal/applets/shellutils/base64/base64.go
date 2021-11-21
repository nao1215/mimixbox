//
// mimixbox/internal/applets/shellutils/base64/base64.go
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
package base64

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	mb "github.com/nao1215/mimixbox/internal/lib"

	"github.com/jessevdk/go-flags"
)

const cmdName string = "base64"

const version = "1.0.0"

var osExit = os.Exit

// Exit code
const (
	ExitSuccess int = iota // 0
	ExitFailuer
)

type options struct {
	Decode  bool `short:"d" long:"decode" description:"Decode data (Default is encode)"`
	Wrap    int  `short:"w" long:"wrap" default:"76" description:"Line break at the Nth character. If N=0, not line break"`
	Version bool `short:"v" long:"version" description:"Show base64 command version"`
}

func Run() (int, error) {
	var opts options
	var args []string
	var err error
	var resultStr string
	var resultByte []byte

	if args, err = parseArgs(&opts); err != nil {
		return ExitFailuer, nil
	}

	input, err := inputByte(args)
	if err != nil {
		return ExitFailuer, err
	}
	if opts.Decode {
		resultByte, err = base64.StdEncoding.DecodeString(string(input))
		if err != nil {
			return ExitFailuer, err
		}
		fmt.Println(mb.WrapString(string(resultByte), opts.Wrap))
	} else {
		resultStr = base64.StdEncoding.EncodeToString(input)
		fmt.Println(mb.WrapString(resultStr, opts.Wrap))
	}
	return ExitSuccess, nil
}

func inputByte(args []string) ([]byte, error) {
	var byteList []byte
	var err error

	if len(args) == 0 || (len(args) == 1 && args[0] == "-") {
		inputStrList := inputFronSTDIN()
		byteList = []byte(strings.Join(inputStrList, ""))
	} else {
		byteList, err = inputFronFile(args[0])
		if err != nil {
			return nil, err
		}
	}
	return byteList, nil
}

func inputFronSTDIN() []string {
	var inputs []string
	for {
		input, next := mb.Input()
		if !next {
			break
		}
		inputs = append(inputs, input)
	}
	return inputs
}

func inputFronFile(file string) ([]byte, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func parseArgs(opts *options) ([]string, error) {
	p := initParser(opts)

	args, err := p.Parse()
	if err != nil {
		return nil, err
	}

	if opts.Version {
		showVersion()
		osExit(ExitSuccess)
	}

	if !isValidArgNr(args) {
		showHelp(p)
		osExit(ExitFailuer)
	}

	if opts.Wrap < 0 {
		return nil, errors.New("Invalid wrap size: " + strconv.Itoa(opts.Wrap))
	}
	return args, nil
}

func initParser(opts *options) *flags.Parser {
	parser := flags.NewParser(opts, flags.Default)
	parser.Name = cmdName
	parser.Usage = "[OPTIONS] FILE_PATH"

	return parser
}

func isValidArgNr(args []string) bool {
	return len(args) <= 1
}

func showVersion() {
	description := cmdName + " version " + version + " (under Apache License verison 2.0)\n"
	fmt.Print(description)
}

func showHelp(p *flags.Parser) {
	p.WriteHelp(os.Stdout)
}
