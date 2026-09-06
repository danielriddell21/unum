package imagetool

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/danielriddell21/unum/internal/image/optimize"
	"github.com/danielriddell21/unum/internal/image/render/static"
	imageTUI "github.com/danielriddell21/unum/internal/image/render/tui"
	imageWeb "github.com/danielriddell21/unum/internal/image/render/web"
	"github.com/danielriddell21/unum/internal/tui/panels"
)

func runTUI(f *flags, src optimize.Source, set settings) error {
	static.Boot(os.Stderr, src.Name, static.Options{Theme: static.ResolveTheme(f.theme), Quiet: f.quiet})
	imageTUI.ApplyPalette(panels.ResolvePalette(f.theme))

	output := f.output
	if output == "" || output == "-" {
		output = defaultOutputName(src.Name, set.opts.Format)
	}

	model := imageTUI.NewModel(src, set.opts, output, f.version, imageTUI.DefaultDeps())
	if _, err := tea.NewProgram(model, tea.WithAltScreen()).Run(); err != nil {
		return fmt.Errorf("image TUI: %w", err)
	}
	return nil
}

func runWeb(f *flags, name string, data []byte, format optimize.Format) error {
	if err := imageWeb.Start(imageWeb.Options{
		Port:       f.webPort,
		Quiet:      f.quiet,
		DarkTheme:  f.theme,
		LightTheme: f.lightTheme,
		Version:    f.version,
		Tel:        f.tel,
		SourceName: name,
		SourceData: data,
		SourceMIME: format.MIME(),
	}); err != nil {
		return fmt.Errorf("image web: %w", err)
	}
	return nil
}

func defaultOutputName(name string, format optimize.Format) string {
	ext := filepath.Ext(name)
	return strings.TrimSuffix(name, ext) + "-small" + format.Ext()
}
