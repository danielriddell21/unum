//go:build ebiten

package gui

import (
	"fmt"
	"image/color"
	"path/filepath"
	"strings"

	eb "github.com/hajimehoshi/ebiten/v2"
	ebinput "github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/danielriddell21/crucible/canvas"
	"github.com/danielriddell21/crucible/keymap"
	"github.com/danielriddell21/crucible/menu"
	"github.com/danielriddell21/crucible/window"
)

const (
	screenW    = 1024
	screenH    = 768
	lineH      = 16
	pad        = 10
	contentTop = 96
)

type screenID int

const (
	screenMenu screenID = iota
	screenJSON
	screenDiff
	screenHash
)

type app struct {
	cfg Config
	cv  *canvas.Canvas
	m   menu.Menu
	th  menu.Theme

	scr    screenID
	lines  []Line
	scroll int
	input  []rune
	quit   bool
}

func Run(cfg Config) error {
	a := &app{cfg: cfg, cv: canvas.New(screenW, screenH), th: menu.DefaultTheme()}
	a.m = menu.Menu{
		Title:    "unum",
		Subtitle: []string{"all tools · one window · v" + cfg.Version},
		Items: []menu.Item{
			{Label: func() string { return "json  " + baseOrNone(cfg.FileA) }, Action: func() { a.open(screenJSON) }},
			{Label: func() string { return "diff  " + baseOrNone(cfg.FileA) + " vs " + baseOrNone(cfg.FileB) }, Action: func() { a.open(screenDiff) }},
			{Label: func() string { return "hash  type to derive" }, Action: func() { a.open(screenHash) }},
			{Label: func() string { return "quit" }, Action: func() { a.quit = true }},
		},
	}
	window.Configure(window.Options{Title: "unum", Width: screenW, Height: screenH, MinWidth: screenW / 2, MinHeight: screenH / 2})
	if err := eb.RunGame(a); err != nil {
		return fmt.Errorf("run window: %w", err)
	}
	return nil
}

func (a *app) open(s screenID) {
	a.scr = s
	a.scroll = 0
	a.refresh()
}

func (a *app) refresh() {
	switch a.scr {
	case screenJSON:
		if a.cfg.FileA == "" {
			a.lines = usageLines("unum window <file>")
			return
		}
		a.lines = JSONLines(a.cfg.FileA)
	case screenDiff:
		if a.cfg.FileA == "" || a.cfg.FileB == "" {
			a.lines = usageLines("unum window <file-a> <file-b>")
			return
		}
		a.lines = DiffLines(a.cfg.FileA, a.cfg.FileB)
	case screenHash:
		a.lines = HashLines(string(a.input))
	}
}

func usageLines(usage string) []Line {
	return []Line{{Text: "no input file · run: " + usage, Kind: KindDim}}
}

func (a *app) Update() error {
	if a.quit {
		return eb.Termination
	}
	if a.scr == screenMenu {
		a.m.Update(menu.Poll())
		if ebinput.IsKeyJustPressed(eb.KeyEscape) || ebinput.IsKeyJustPressed(eb.KeyQ) {
			return eb.Termination
		}
		return nil
	}
	if ebinput.IsKeyJustPressed(eb.KeyEscape) {
		a.scr = screenMenu
		return nil
	}
	if a.scr == screenHash {
		a.typeInput()
	}
	a.scrollInput()
	return nil
}

func (a *app) typeInput() {
	changed := false
	for _, r := range eb.AppendInputChars(nil) {
		if r >= ' ' {
			a.input = append(a.input, r)
			changed = true
		}
	}
	if keyRepeat(eb.KeyBackspace) && len(a.input) > 0 {
		a.input = a.input[:len(a.input)-1]
		changed = true
	}
	if changed {
		a.refresh()
	}
}

func (a *app) scrollInput() {
	page := visibleRows()
	switch {
	case keyRepeat(eb.KeyDown):
		a.scroll++
	case keyRepeat(eb.KeyUp):
		a.scroll--
	case ebinput.IsKeyJustPressed(eb.KeyPageDown):
		a.scroll += page
	case ebinput.IsKeyJustPressed(eb.KeyPageUp):
		a.scroll -= page
	case ebinput.IsKeyJustPressed(eb.KeyHome):
		a.scroll = 0
	case ebinput.IsKeyJustPressed(eb.KeyEnd):
		a.scroll = len(a.lines)
	}
	a.scroll = max(0, min(a.scroll, max(0, len(a.lines)-page)))
}

func keyRepeat(k eb.Key) bool {
	d := ebinput.KeyPressDuration(k)
	return d == 1 || (d >= 30 && d%3 == 0)
}

func (a *app) Draw(screen *eb.Image) {
	a.cv.Fill(color.RGBA{R: 18, G: 20, B: 26, A: 255})
	if a.scr == screenMenu {
		a.m.Draw(a.cv, a.th)
	} else {
		a.drawTool()
	}
	a.drawKeymap()
	screen.WritePixels(a.cv.Pixels())
}

func (a *app) drawTool() {
	a.cv.Text(pad, 26, truncate(a.toolTitle()), a.th.Title)
	if a.scr == screenHash {
		a.cv.Text(pad, 56, truncate("> "+string(a.input)+"_"), a.th.Selected)
	}
	page := visibleRows()
	y := contentTop
	for _, ln := range a.lines[a.scroll:min(a.scroll+page, len(a.lines))] {
		a.cv.Text(pad, y, truncate(expandTabs(ln.Text)), a.kindColor(ln.Kind))
		y += lineH
	}
	if len(a.lines) > page {
		pos := fmt.Sprintf("%d-%d/%d", a.scroll+1, min(a.scroll+page, len(a.lines)), len(a.lines))
		a.cv.Text(screenW-pad-len(pos)*canvas.GlyphWidth, 26, pos, a.th.Dim)
	}
}

func (a *app) toolTitle() string {
	switch a.scr {
	case screenDiff:
		return "diff  " + a.cfg.FileA + " vs " + a.cfg.FileB
	case screenHash:
		return "hash"
	default:
		return "json  " + a.cfg.FileA
	}
}

func (a *app) kindColor(k LineKind) color.RGBA {
	switch k {
	case KindDim:
		return a.th.Dim
	case KindAdded:
		return a.th.Selected
	case KindRemoved:
		return a.th.Title
	case KindModified:
		return color.RGBA{R: 230, G: 200, B: 90, A: 255}
	default:
		return a.th.Text
	}
}

func (a *app) drawKeymap() {
	f := keymap.Face{LineHeight: lineH, Measure: func(s string) int { return len(s) * canvas.GlyphWidth }}
	for _, ln := range keymap.BottomBar(a.bindings(), screenW, screenH, pad, f) {
		// +11: BottomBar returns top-left origins; Text draws from the
		// baseline, and the 7x13 face's ascent is 11.
		a.cv.Text(ln.X, ln.Y+11, ln.Text, a.th.Dim)
	}
}

func (a *app) bindings() []keymap.Binding {
	switch a.scr {
	case screenMenu:
		return []keymap.Binding{
			{Key: "arrows", Action: "navigate"},
			{Key: "enter", Action: "select"},
			{Key: "q/esc", Action: "quit"},
		}
	case screenHash:
		return []keymap.Binding{
			{Key: "type", Action: "derive"},
			{Key: "backspace", Action: "delete"},
			{Key: "esc", Action: "menu"},
		}
	default:
		return []keymap.Binding{
			{Key: "up/down", Action: "scroll"},
			{Key: "pgup/pgdn", Action: "page"},
			{Key: "home/end", Action: "jump"},
			{Key: "esc", Action: "menu"},
		}
	}
}

func visibleRows() int {
	return (screenH - contentTop - 3*lineH) / lineH
}

func truncate(s string) string {
	maxCols := (screenW - 2*pad) / canvas.GlyphWidth
	r := []rune(s)
	if len(r) <= maxCols {
		return s
	}
	return string(r[:maxCols-3]) + "..."
}

func expandTabs(s string) string {
	return strings.ReplaceAll(s, "\t", "    ")
}

func baseOrNone(path string) string {
	if path == "" {
		return "(no file)"
	}
	return filepath.Base(path)
}

func (a *app) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenW, screenH
}
