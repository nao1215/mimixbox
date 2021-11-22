//
// mimixbox/internal/applets/shellutils/defm/view.go
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
package defm // Desktop Entry File Manager

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const (
	InputPanel int = iota + 1
	DesktopEntryNamePanel
	DesktopEntryBodyPanel
)

type Gui struct {
	FilterInput          *tview.InputField
	DesktopEntryMgr      *DesktopEntryManager
	DesktopEntryBodyView *DesktopEntryBodyView
	App                  *tview.Application
	Pages                *tview.Pages
	Panels
}

type DesktopEntryManager struct {
	*tview.Table
	DEFiles    []DesktopEntryFile
	FilterWord string
}
type DesktopEntryBodyView struct {
	*tview.TextView
}

type Panels struct {
	Current int
	Panels  []tview.Primitive
	Kinds   []int
}

func NewAllView() *Gui {
	filter := tview.NewInputField().SetLabel("app name:")
	desktopEntryMgr := NewDesktopEntryManager()
	desktopEntryBodyView := NewDesktopEntryBodyView()

	gui := &Gui{
		FilterInput:          filter,
		DesktopEntryMgr:      desktopEntryMgr,
		DesktopEntryBodyView: desktopEntryBodyView,
		App:                  tview.NewApplication(),
	}

	gui.Panels = Panels{
		Panels: []tview.Primitive{
			filter,
			desktopEntryMgr,
		},
		Kinds: []int{
			InputPanel,
			DesktopEntryNamePanel,
			DesktopEntryBodyPanel,
		},
	}
	return gui
}

func (g *Gui) Run() error {
	g.SetKeybinds()
	if err := g.DesktopEntryMgr.UpdateView(); err != nil {
		return err
	}

	g.DesktopEntryMgr.Select(1, 0)
	g.UpdateViews()

	infoGrid := tview.NewGrid().
		SetRows(0, 0, 0, 0).
		SetColumns(0, 0).
		AddItem(g.DesktopEntryMgr, 1, 0, 1, 1, 0, 0, true).
		AddItem(g.DesktopEntryBodyView, 1, 1, 1, 1, 0, 0, true)

	grid := tview.NewGrid().
		SetSize(2, 2, 0, 0).
		AddItem(g.FilterInput, 0, 0, 1, 1, 0, 0, true).
		AddItem(infoGrid, 1, 1, 1, 2, 0, 0, true)

	g.Pages = tview.NewPages().AddAndSwitchToPage("main", grid, true)

	if err := g.App.SetRoot(g.Pages, true).Run(); err != nil {
		g.App.Stop()
		return err
	}

	return nil
}

func NewDesktopEntryManager() *DesktopEntryManager {
	dem := &DesktopEntryManager{
		Table: tview.NewTable().Select(0, 0).SetFixed(1, 1).SetSelectable(true, false),
	}
	dem.SetBorder(true).SetTitle("Desktop Entry").SetTitleAlign(tview.AlignLeft)
	return dem
}

func NewDesktopEntryBodyView() *DesktopEntryBodyView {
	p := &DesktopEntryBodyView{
		TextView: tview.NewTextView().SetTextAlign(tview.AlignLeft).SetDynamicColors(true),
	}
	p.SetTitleAlign(tview.AlignLeft).SetBorder(true)
	p.SetWrap(false)
	return p
}

func (dem *DesktopEntryManager) UpdateView() error {
	if err := dem.updateDEFiles(); err != nil {
		return err
	}

	// Set desktop entry filename
	table := dem.Clear()
	for i, def := range dem.DEFiles {
		table.SetCell(i+1, 0, tview.NewTableCell(def.basename))
		i++
	}
	return nil
}

func (dem *DesktopEntryManager) Selected() *DesktopEntryFile {
	if len(dem.DEFiles) == 0 {
		return nil
	}

	row, _ := dem.GetSelection()
	if row < 0 || len(dem.DEFiles) < row {
		return nil
	}

	return &dem.DEFiles[row-1]
}

func (g *Gui) SetKeybinds() {
	g.FilterInputKeybinds()
}

func (g *Gui) FilterInputKeybinds() {
	g.FilterInput.SetDoneFunc(func(key tcell.Key) {
		switch key {
		case tcell.KeyEscape:
			g.App.Stop()
		case tcell.KeyEnter:
			g.nextPanel()
		}
	}).SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		g.GrobalKeybind(event)
		return event
	})

	g.FilterInput.SetChangedFunc(func(text string) {
		g.DesktopEntryMgr.FilterWord = text
		g.DesktopEntryMgr.UpdateView()
	})
}

func (g *Gui) GrobalKeybind(event *tcell.EventKey) {
	switch event.Key() {
	case tcell.KeyTab:
		g.nextPanel()
	case tcell.KeyBacktab:
		g.prePanel()
	}

	//g.NaviView.UpdateView(g)
}

func (g *Gui) nextPanel() {
	idx := (g.Panels.Current + 1) % len(g.Panels.Panels)
	g.Panels.Current = idx
	g.SwitchPanel(g.Panels.Panels[g.Panels.Current])
}

func (g *Gui) prePanel() {
	g.Panels.Current--

	if g.Panels.Current < 0 {
		g.Current = len(g.Panels.Panels) - 1
	} else {
		idx := (g.Panels.Current) % len(g.Panels.Panels)
		g.Panels.Current = idx
	}
	g.SwitchPanel(g.Panels.Panels[g.Panels.Current])
}

func (g *Gui) SwitchPanel(p tview.Primitive) *tview.Application {
	g.UpdateViews()
	return g.App.SetFocus(p)
}

func (g *Gui) CurrentPanelKind() int {
	return g.Panels.Kinds[g.Panels.Current]
}

func (g *Gui) UpdateViews() {
	g.DesktopEntryBodyView.Update(g)
}

func (deBdoyView *DesktopEntryBodyView) Update(g *Gui) {
	deFile := g.DesktopEntryMgr.Selected()
	if deFile != nil {
		g.DesktopEntryBodyView.SetText(deFile.body)
	}
}
