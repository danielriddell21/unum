package cli

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/spf13/cobra"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	ometric "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"

	"github.com/danielriddell21/unum/internal/gui"
	"github.com/danielriddell21/unum/internal/telemetry"
)

func windowCmd(version string, tel *telemetry.Telemetry) *cobra.Command {
	return &cobra.Command{
		Use:   "window [file-a] [file-b]",
		Short: "Open the native window app with every tool",
		Long: `Open one native window that encompasses all the tools.

A menu lists every tool; select one to open its view:
  json   Pretty-printed view of file-a (scrollable)
  diff   Coloured diff of file-a vs file-b (scrollable)
  hash   Live derivation table — type to derive

Navigation follows the family conventions: arrows or WASD to move,
enter or space to select, esc back to the menu.

The window build is opt-in: it needs a display and OpenGL, so build
with -tags ebiten (just window) to enable it.`,
		Args:         cobra.MaximumNArgs(2),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, span := tel.Tracer().Start(context.Background(), "window.execute",
				trace.WithAttributes(
					attribute.String("tool", "window"),
					attribute.String("mode", "window"),
					attribute.String("os", runtime.GOOS),
					attribute.String("arch", runtime.GOARCH),
				),
			)
			defer span.End()

			cfg := gui.Config{Version: version}
			if len(args) > 0 {
				cfg.FileA = args[0]
			}
			if len(args) > 1 {
				cfg.FileB = args[1]
			}

			start := time.Now()
			if err := gui.Run(cfg); err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, "window")
				tel.M.Errors.Add(ctx, 1, ometric.WithAttributes(
					attribute.String("tool", "window"),
					attribute.String("error_type", "window"),
				))
				return fmt.Errorf("window: %w", err)
			}
			tel.M.Invocations.Add(ctx, 1, ometric.WithAttributes(
				attribute.String("tool", "window"),
				attribute.String("mode", "window"),
				attribute.String("os", runtime.GOOS),
				attribute.String("arch", runtime.GOARCH),
				attribute.String("version", version),
			))
			tel.M.Duration.Record(ctx, time.Since(start).Seconds(), ometric.WithAttributes(
				attribute.String("tool", "window"),
				attribute.String("mode", "window"),
			))
			return nil
		},
	}
}
