package panels

import (
	"fmt"
	"strings"

	"github.com/danielriddell21/unum/internal/image/optimize"
)

type SettingsView struct {
	Source   optimize.Source
	Format   optimize.Format
	Quality  int
	Scale    float64
	Width    int
	Height   int
	Result   *optimize.Result
	Steps    []optimize.Step
	Encoding bool
	Err      error
	Saved    string
}

type SettingsPanel struct {
	width  int
	height int
}

func NewSettingsPanel(w, h int) SettingsPanel {
	return SettingsPanel{width: w, height: h}
}

func (p *SettingsPanel) Resize(w, h int) {
	p.width = w
	p.height = h
}

func (p *SettingsPanel) View(v SettingsView) string {
	sep := styleHint.Render(strings.Repeat("─", max(p.width, 1)))

	lines := []string{""}
	lines = append(lines, row("source", v.Source.Name))
	lines = append(lines, row("original", fmt.Sprintf("%s  %d × %d  %s",
		v.Source.Format, v.Source.Width, v.Source.Height, optimize.HumanBytes(v.Source.Bytes))))
	lines = append(lines, sep)
	lines = append(lines, row("format", v.Format.String()))
	lines = append(lines, row("quality", fmt.Sprintf("%d", v.Quality)))
	lines = append(lines, row("scale", fmt.Sprintf("%.0f%%", v.Scale*100)))
	lines = append(lines, row("output", fmt.Sprintf("%d × %d", v.Width, v.Height)))
	lines = append(lines, sep)
	lines = append(lines, resultLine(v))

	if len(v.Steps) > 0 {
		lines = append(lines, "", styleHint.Render("  quality    size         saving"))
		for _, step := range v.Steps {
			lines = append(lines, ladderRow(v, step))
		}
	}

	if v.Saved != "" {
		lines = append(lines, "", styleWin.Render("  saved → "+v.Saved))
	}

	return strings.Join(lines, "\n")
}

func resultLine(v SettingsView) string {
	switch {
	case v.Err != nil:
		return styleErr.Render("  " + v.Err.Error())
	case v.Encoding || v.Result == nil:
		return styleHint.Render("  encoding…")
	default:
		saving := optimize.Saving(v.Source.Bytes, v.Result.Bytes)
		style := styleValue
		if saving > 0 {
			style = styleWin
		}
		return fmt.Sprintf("%s  %s  %s",
			styleLabel.Render(fmt.Sprintf("  %-9s", "result")),
			styleValue.Render(optimize.HumanBytes(v.Result.Bytes)),
			style.Render(fmt.Sprintf("%+.0f%%", -saving*100)))
	}
}

func ladderRow(v SettingsView, step optimize.Step) string {
	saving := optimize.Saving(v.Source.Bytes, step.Bytes)
	text := fmt.Sprintf("  %-10d %-12s %+.0f%%", step.Quality, optimize.HumanBytes(step.Bytes), -saving*100)
	if step.Quality == v.Quality {
		return styleActive.Render(text)
	}
	return styleValue.Render(text)
}

func row(label, value string) string {
	return fmt.Sprintf("%s  %s",
		styleLabel.Render(fmt.Sprintf("  %-9s", label)),
		styleValue.Render(value))
}
