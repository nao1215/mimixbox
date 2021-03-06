//
// mimixbox/internal/applets/shellutils/wget/wget.go
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
package wget

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"

	"github.com/jessevdk/go-flags"
	mb "github.com/nao1215/mimixbox/internal/lib"
)

const cmdName string = "wget"

const version = "1.0.0"

var osExit = os.Exit

type options struct {
	Version bool `short:"v" long:"version" description:"Show wget command version"`
}

func Run() (int, error) {
	var opts options
	var args []string
	var err error

	if args, err = parseArgs(&opts); err != nil {
		return mb.ExitFailure, nil
	}
	return wget(args)
}

func wget(args []string) (int, error) {
	status := mb.ExitSuccess
	for _, v := range args {
		target, err := openTargetFile(v)
		if err != nil {
			fmt.Fprintln(os.Stderr, cmdName+": "+err.Error()+": v")
			status = mb.ExitFailure
			continue
		}
		defer target.Close()

		client := http.Client{}
		response, err := client.Get(v)
		if err != nil {
			fmt.Fprintln(os.Stderr, cmdName+": "+err.Error()+": v")
			status = mb.ExitFailure
			continue
		}
		defer response.Body.Close()

		_, err = io.Copy(target, response.Body)
		if err != nil {
			fmt.Fprintln(os.Stderr, cmdName+": "+err.Error()+": v")
			status = mb.ExitFailure
			continue
		}
	}
	return status, nil
}

func openTargetFile(urlStr string) (io.WriteCloser, error) {
	filename, err := targetFilename(urlStr)
	if err != nil {
		return nil, err
	}
	return os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0644)
}

func targetFilename(urlStr string) (string, error) {
	url, err := url.Parse(urlStr)
	if err != nil {
		return "", err
	}

	filename := path.Base(url.Path)
	if len(filename) == 0 || filename == "." {
		filename = "index.html"
	}
	return preventFilenameDup(filename), nil
}

func preventFilenameDup(filename string) string {
	if !mb.IsFile(filename) {
		return filename
	}

	i := 1
	for {
		newName := filename + "." + strconv.Itoa(i)
		if !mb.IsFile(newName) {
			return newName
		}
		i++
	}
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

	if !isValidArgNr(args) {
		fmt.Fprintln(os.Stderr, "wget: missing URL")
		osExit(mb.ExitFailure)
	}

	return args, nil
}

func initParser(opts *options) *flags.Parser {
	parser := flags.NewParser(opts, flags.Default)
	parser.Name = cmdName
	parser.Usage = "[OPTIONS] URL"

	return parser
}

func isValidArgNr(args []string) bool {
	return len(args) >= 1
}
