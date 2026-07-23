//go:build !ebiten

package gui

import "errors"

func Run(Config) error {
	return errors.New("native window unavailable in this build: rebuild with `-tags ebiten` (e.g. `just window`), or use --ui / --web on each tool")
}
