//
// mimixbox/internal/applets/jokeutils/fakemovie/fakemovie.go
//
// Copyright © 2021 by Yasuhiro Matsumoto (a.k.a. mattn)
// Changes since September 2021 copyright © by Naohiro CHIKAMATSU
// This file is forked from https://github.com/mattn/fakemovie at 2021/09/18.
//
//-----------------------[Original license begin]------------------------------------
// MIT License
//
// Copyright 2021 mattn
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
//-----------------------[Original license end]------------------------------------
//
// The code that did not exist in the original fakemovie is shown below.
//   - Semantic Versioning
//   - run()         : Move the main process from main() to run()
//   - parseArgs()   : Argument parsing function
//   - initParser()  : Parser initialization function
//   - isValidArgNr(): Function to check if the number of arguments is correct
//   - isValidExt()  : Function to check if image file extension is correct
//   - showVersion() : Function to show fakemovie version
//   - showHelp()    : Function to show help message
//   - openImage()   : Function to open image file
//   - writeImage()  : Function to write image file
//	 - addBlueBtn()  : Rename Fake() to addBlueBtn()
//   - addOrangeBtn(): Function to add orange button to the image
//   - decideOutputFileName()     : Function that make string with "original file name" + "_fake" + "extension"
//   - extractFileNameWithoutExt(): Function that return filename without extension.
//
// The above code is the copyright below.
//
// Copyright 2021 Naohiro CHIKAMATSU.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package fakemovie

import (
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/fogleman/gg"
	"github.com/jessevdk/go-flags"
)

const cmdName string = "fakemovie"

var osExit = os.Exit

const version = "1.0.1"

// Exit code
const (
	ExitSuccess int = iota // 0
	ExitFailuer
)

type options struct {
	Output  string `short:"o" long:"output" value-name:"<output-file-name>" description:"Output file name(default: Added suffix \"_fake\" to original name)"`
	Phub    bool   `short:"p" long:"phub" description:"Put p-hub button(default: Color similar to twitter button)"`
	Radius  int    `short:"r" long:"radius" value-name:"<number(integer)>" description:"Radius of button(default: Auto caluculate)"`
	Version bool   `short:"v" long:"version" description:"Show fakemovie command version"`
}

func Run() error {
	var opts options
	var args = parseArgs(&opts)
	var inputFileName string = args[0]
	var outputFileName string = opts.Output
	var radius int = opts.Radius

	if !isValidExt(inputFileName) {
		return errors.New("fakemovie command only support jpg or png.")
	}

	img, err := openImage(inputFileName)
	if err != nil {
		return err
	}

	if radius <= 0 {
		radius = calcButtonRadius(img)
	}

	if outputFileName == "" {
		outputFileName = decideOutputFileName(inputFileName)
	}

	if opts.Phub {
		img = addOrangeBtn(img, radius)
	} else {
		img = addBlueBtn(img, radius)
	}
	err = writeImage(img, outputFileName)
	if err != nil {
		return err
	}

	return nil
}

func openImage(imageFileName string) (image.Image, error) {
	f, err := os.Open(imageFileName)
	if err != nil {
		return nil, err
	}

	img, _, err := image.Decode(f)
	if err != nil {
		return nil, err
	}
	f.Close()

	return img, nil
}

func isValidExt(imageFileName string) bool {
	targets := []string{"jpg", "jpeg", "png"}
	var ext string = filepath.Ext(imageFileName)

	for _, target := range targets {
		if strings.Contains(ext, target) {
			return true
		}
	}
	return false
}

func calcButtonRadius(img image.Image) int {
	const ImgDivisionNum = 14 // Manually adjusted value.
	rect := img.Bounds()

	// This calculation algorithm is not meaningful.
	// Extremely different aspect ratios of images give strange radius values.
	xRadiusSize := rect.Max.X / ImgDivisionNum
	yRadiusSize := rect.Max.Y / ImgDivisionNum
	return (xRadiusSize + yRadiusSize) / 2
}

func decideOutputFileName(inputFileName string) string {
	return extractFileNameWithoutExt(inputFileName) + "_fake" + filepath.Ext(inputFileName)
}

func extractFileNameWithoutExt(path string) string {
	return filepath.Base(path[:len(path)-len(filepath.Ext(path))])
}

func writeImage(img image.Image, outputFileName string) error {
	f, err := os.Create(outputFileName)
	if err != nil {
		return err
	}

	if filepath.Ext(outputFileName) == "png" {
		err = png.Encode(f, img)
	} else {
		err = jpeg.Encode(f, img, nil) // nil means default quality (=75, max=100)
	}
	if err != nil {
		return err
	}
	f.Close()

	return nil
}

func addBlueBtn(img image.Image, r int) image.Image {
	bounds := img.Bounds()
	dx, dy := bounds.Dx(), bounds.Dy()

	dc := gg.NewContext(dx, dy)
	dc.DrawImage(img, 0, 0)                                      // draw original image.
	dc.DrawCircle(float64(dx/2), float64(dy/2), float64(r)*1.25) // draw outside the circle
	dc.SetRGB(1.0, 1.0, 1.0)                                     // White
	dc.Fill()

	dc.DrawCircle(float64(dx/2), float64(dy/2), float64(r))
	dc.SetRGB(0.3, 0.7, 1.0) // Blue like twitter
	dc.Fill()

	rr := float64(r) / 4
	dc.MoveTo(float64(dx/2)-rr, float64(dy/2)-rr*2)
	dc.LineTo(float64(dx/2)+rr*2, float64(dy/2))
	dc.LineTo(float64(dx/2)-rr, float64(dy/2)+rr*2)
	dc.ClosePath()
	dc.SetRGB(1.0, 1.0, 1.0) // White
	dc.Fill()

	return dc.Image()
}

func addOrangeBtn(img image.Image, r int) image.Image {
	bounds := img.Bounds()
	dx, dy := bounds.Dx(), bounds.Dy()

	dc := gg.NewContext(dx, dy)
	dc.DrawImage(img, 0, 0) // draw original image.

	dc.DrawCircle(float64(dx/2), float64(dy/2), float64(r)*1.12) // draw outside the circle
	dc.SetRGB(0.905, 0.588, 0.2)                                 // Orange like p-hub
	dc.Fill()

	dc.DrawCircle(float64(dx/2), float64(dy/2), float64(r))
	dc.SetRGB(0, 0, 0) // Black
	dc.Fill()

	rr := float64(r) / 3.5
	dc.MoveTo(float64(dx/2)-rr, float64(dy/2)-rr*2)
	dc.LineTo(float64(dx/2)+rr*2, float64(dy/2))
	dc.LineTo(float64(dx/2)-rr, float64(dy/2)+rr*2)
	dc.ClosePath()
	dc.SetRGB(0.905, 0.588, 0.2) // Orange like p-hub
	dc.Fill()

	return dc.Image()
}

func parseArgs(opts *options) []string {
	p := initParser(opts)

	args, err := p.Parse()
	if err != nil {
		osExit(ExitFailuer)
	}

	if opts.Version {
		showVersion()
		osExit(ExitSuccess)
	}

	if !isValidArgNr(args) {
		showHelp(p)
		osExit(ExitFailuer)
	}
	return args
}

func initParser(opts *options) *flags.Parser {
	parser := flags.NewParser(opts, flags.Default)
	parser.Name = cmdName
	parser.Usage = "[OPTIONS] IMAGE_FILE_NAME"

	return parser
}

func isValidArgNr(args []string) bool {
	return len(args) == 1
}

func showVersion() {
	fmt.Printf("%s version %s\n", cmdName, version)
}

func showHelp(p *flags.Parser) {
	fmt.Printf("fakemovie adds fake-movie button to the image.\n\n")
	p.WriteHelp(os.Stdout)
}
