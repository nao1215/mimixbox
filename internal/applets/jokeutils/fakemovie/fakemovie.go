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

// Package fakemovie implements the fakemovie applet: it draws a fake video
// playback button onto an image so the result looks like a movie thumbnail.
package fakemovie

import (
	"context"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/fogleman/gg"
	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the fakemovie applet.
type Command struct{}

// New returns a fakemovie command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "fakemovie" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Adds a video playback button to the image" }

type options struct {
	output string
	phub   bool
	radius int
}

// Run executes fakemovie.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]... IMAGE_FILE...", stdio.Err).WithHelp(command.Help{
		Description: "Draw a fake video playback button onto each IMAGE_FILE so the result looks like a " +
			"movie thumbnail, writing a new image alongside the original.",
		Examples: []command.Example{
			{Command: "fakemovie photo.png", Explain: "Write photo_fake.png with a play button drawn on it."},
			{Command: "fakemovie -o out.png photo.png", Explain: "Write the result to out.png."},
			{Command: "fakemovie -p photo.png", Explain: "Use a p-hub style button instead of the default."},
		},
		ExitStatus: "0  success.\n1  an error occurred (e.g. the image could not be read or written).",
	})
	output := fs.StringP("output", "o", "", "output file name (default: add the suffix \"_fake\" to the original name)")
	phub := fs.BoolP("phub", "p", false, "put a p-hub style button (default: a twitter-like button)")
	radius := fs.IntP("radius", "r", 0, "radius of the button (default: auto calculate)")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	files := fs.Args()
	if len(files) == 0 {
		_, _ = fmt.Fprintf(stdio.Err, "fakemovie: missing operand\n")
		return command.SilentFailure()
	}

	opts := options{
		output: *output,
		phub:   *phub,
		radius: *radius,
	}

	var firstErr error
	for _, name := range files {
		if err := processFile(name, opts); err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "fakemovie: %s\n", command.FileError(name, err))
			firstErr = keep(firstErr)
		}
	}
	return firstErr
}

// processFile reads one input image, draws the play-button overlay onto it and
// writes the result back out. The output path and button style come from opts.
func processFile(name string, opts options) error {
	input := os.ExpandEnv(name)
	if !isValidExt(input) {
		return errors.New("fakemovie command only supports jpg or png")
	}

	img, err := openImage(input)
	if err != nil {
		return err
	}

	radius := opts.radius
	if radius <= 0 {
		radius = calcButtonRadius(img)
	}

	output := os.ExpandEnv(opts.output)
	if output == "" {
		output = decideOutputFileName(input)
	}

	img = AddButton(img, radius, opts.phub)

	return writeImage(img, output)
}

// AddButton returns a copy of img with a fake video playback button drawn at its
// center. radius controls the button size; when phub is true an orange p-hub
// style button is drawn, otherwise a blue twitter-like button. AddButton is a
// pure function (no file IO) so it can be exercised directly by tests.
func AddButton(img image.Image, radius int, phub bool) image.Image {
	if phub {
		return addOrangeBtn(img, radius)
	}
	return addBlueBtn(img, radius)
}

func openImage(imageFileName string) (image.Image, error) {
	f, err := os.Open(imageFileName) //nolint:gosec // operating on a user-named file is the whole point
	if err != nil {
		return nil, err
	}
	defer f.Close() //nolint:errcheck // read-only handle; a close error here is harmless

	img, _, err := image.Decode(f)
	if err != nil {
		return nil, err
	}
	return img, nil
}

func isValidExt(imageFileName string) bool {
	targets := []string{"jpg", "jpeg", "png"}
	ext := strings.ToLower(filepath.Ext(imageFileName))

	for _, target := range targets {
		if strings.Contains(ext, target) {
			return true
		}
	}
	return false
}

func calcButtonRadius(img image.Image) int {
	const imgDivisionNum = 14 // Manually adjusted value.
	rect := img.Bounds()

	// This calculation algorithm is not meaningful.
	// Extremely different aspect ratios of images give strange radius values.
	xRadiusSize := rect.Max.X / imgDivisionNum
	yRadiusSize := rect.Max.Y / imgDivisionNum
	return (xRadiusSize + yRadiusSize) / 2
}

func decideOutputFileName(inputFileName string) string {
	return extractFileNameWithoutExt(inputFileName) + "_fake" + filepath.Ext(inputFileName)
}

func extractFileNameWithoutExt(path string) string {
	return filepath.Base(path[:len(path)-len(filepath.Ext(path))])
}

// writeImage encodes img to outputFileName. A close error on the write path is
// returned, because a failed close can mean the file was left corrupt.
func writeImage(img image.Image, outputFileName string) (err error) {
	f, err := os.Create(outputFileName) //nolint:gosec // operating on a user-named file is the whole point
	if err != nil {
		return err
	}
	defer func() {
		if cerr := f.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	if strings.EqualFold(filepath.Ext(outputFileName), ".png") {
		err = png.Encode(f, img)
	} else {
		err = jpeg.Encode(f, img, nil) // nil means default quality (=75, max=100)
	}
	return err
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

func keep(existing error) error {
	if existing != nil {
		return existing
	}
	return command.SilentFailure()
}
