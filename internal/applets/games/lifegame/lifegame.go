//
// mimixbox/internal/applets/games/lifegame/lifegame.go
//
// Copyright 2021 Naohiro CHIKAMATSU, polynomialspace
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
package lifegame

import (
	"math/rand"
	"os"
	"time"

	mb "github.com/nao1215/mimixbox/internal/lib"
	"github.com/nsf/termbox-go"

	"github.com/jessevdk/go-flags"
)

var cmdName string = "lifegame"

const version = "1.0.0"

var osExit = os.Exit

// Exit code
const (
	ExitSuccess int = iota // 0
	ExitFailuer
)

type options struct {
	Version bool `short:"v" long:"version" description:"Show lifegame command version"`
}

const (
	BackGround = termbox.ColorBlack
	Alive      = termbox.ColorWhite
	Interval   = 100 * time.Millisecond
)

type Window struct {
	width  int
	height int
}

type Field struct {
	matrix [][]bool //[row][column]
}

type Game struct {
	win   Window
	field Field
	queue chan termbox.Event
}

func Run() (int, error) {
	var opts options
	var args []string
	var err error

	if args, err = parseArgs(&opts); err != nil {
		return ExitFailuer, nil
	}
	return lifegame(args, opts)
}

func lifegame(args []string, opts options) (int, error) {
	if err := termbox.Init(); err != nil {
		return ExitFailuer, err
	}
	defer termbox.Close()

	var game Game
	g := game.New(termbox.Size())
	return g.start()
}

func (g *Game) New(w, h int) *Game {
	var field Field

	return &Game{
		win:   Window{w, h},
		field: field.New(w, h),
		queue: pollEvent(),
	}
}

func (f *Field) New(w, h int) Field {
	row := make([][]bool, h)
	for i := 0; i < h; i++ {
		row[i] = make([]bool, w)
	}
	return Field{row}
}

func (g *Game) start() (int, error) {
	g.field.random(g.win.width, g.win.height)

	for {
		select {
		case ev := <-g.queue:
			if isGameEnd(ev.Key) {
				return ExitSuccess, nil
			}
		default:
			time.Sleep(Interval)
			g.field.render(g.win.width, g.win.height)
			g.field.update(g.win.width, g.win.height)
		}
	}
}

func isGameEnd(key termbox.Key) bool {
	return key == termbox.KeyEsc || key == termbox.KeyCtrlD || key == termbox.KeyCtrlC
}

func (f *Field) random(w, h int) {
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < h; i++ {
		for j := 0; j < w; j++ {
			if rand.Intn(20) == 0 {
				f.matrix[i][j] = true
			} else {
				f.matrix[i][j] = false
			}
		}
	}
}

func pollEvent() chan termbox.Event {
	q := make(chan termbox.Event)
	go func() {
		for {
			q <- termbox.PollEvent()
		}
	}()
	return q
}

func (f *Field) exist(row, column, width, height int) bool {
	if row < 0 || column < 0 || row >= height || column >= width {
		return false
	}
	if f.matrix[row][column] {
		return true
	}
	return false
}

func (f *Field) countNeighborhood(row, column, width, height int) int {
	var sum = 0
	for i := -1; i < 2; i++ {
		for j := -1; j < 2; j++ {
			if i == 0 && j == 0 {
				continue
			}
			if f.exist(row+i, column+j, width, height) {
				sum += 1
			}
		}
	}
	return sum
}

//
// [Birth] 3 alive cell near 1 dead cell
// ■■□
// ■□□
// □□□
//
// [Alive(maintenance)] 2 or 3 alive cell near 1 alive cell
// □□□□
// □■■□
// □■■□
// □□□□
//
// [Death (depopulation)] 0 or 1 alive cell near 1 alive cell
// □□□
// □■■
// □□□
//
// [Death (overcrowding)] over 4 alive cell near 1 alive cell
// ■■■
// ■■□
// □□□
//
func (f *Field) update(w, h int) {
	var field Field
	newFields := field.New(w, h)

	var count = 0
	for i := 0; i < h; i++ {
		for j := 0; j < w; j++ {
			count = f.countNeighborhood(i, j, w, h)
			if !f.isAlive(i, j) {
				f.matrix[i][j] = (count == 3) // Birth
			} else {
				f.matrix[i][j] = (count == 2 || count == 3) // Dead or Alive
			}
		}
	}
	f = &newFields
}

func (f *Field) isAlive(row, column int) bool {
	return f.matrix[row][column]
}

func (f *Field) render(w, h int) {
	termbox.Clear(BackGround, BackGround)
	for i := 0; i < h; i++ {
		for j := 0; j < w; j++ {
			if f.matrix[i][j] {
				termbox.SetCell(j, i, '█', Alive, BackGround)
			}
		}
	}
	termbox.Flush()
}

func parseArgs(opts *options) ([]string, error) {
	p := initParser(opts)

	args, err := p.Parse()
	if err != nil {
		return nil, err
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
	parser.Usage = "[OPTIONS] "

	return parser
}
