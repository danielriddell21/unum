package analyze

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/danielriddell21/unum/internal/json/node"
)

// Analyzer is implemented by every annotation lens.
// Each lens walks the node tree and stamps its findings as Annotations.
// Implementations must be safe to call concurrently with other Analyzers.
type Analyzer interface {
	Name() string // "stats", "merkle", "decode" — drives TUI labels and CLI flag names
	Run(ctx context.Context, root *node.Node, opts Options) error
}

// Options carries all per-run configuration for annotation lenses.
type Options struct {
	RunStats  bool
	RunMerkle bool
	RunDecode bool
}

// Suite holds a set of analyzers and runs them concurrently.
type Suite struct {
	analyzers []Analyzer
}

// NewSuite creates a Suite from the given analyzers.
func NewSuite(analyzers ...Analyzer) *Suite {
	return &Suite{analyzers: analyzers}
}

// Run executes all analyzers concurrently and waits for all to finish.
// Each analyzer writes only its own annotation keys, so concurrent writes are safe.
func (s *Suite) Run(ctx context.Context, root *node.Node, opts Options) error {
	if len(s.analyzers) == 0 {
		return nil
	}

	var (
		wg   sync.WaitGroup
		mu   sync.Mutex
		errs []error
	)

	for _, a := range s.analyzers {
		a := a
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := a.Run(ctx, root, opts); err != nil {
				mu.Lock()
				errs = append(errs, fmt.Errorf("%s lens: %w", a.Name(), err))
				mu.Unlock()
			}
		}()
	}

	wg.Wait()
	return errors.Join(errs...)
}

// Build constructs a Suite from Options, adding only the enabled analyzers.
// Lenses are registered here; adding a new lens means adding one line.
func Build(opts Options, all []Analyzer) *Suite {
	var enabled []Analyzer
	for _, a := range all {
		switch a.Name() {
		case "stats":
			if opts.RunStats {
				enabled = append(enabled, a)
			}
		case "merkle":
			if opts.RunMerkle {
				enabled = append(enabled, a)
			}
		case "decode":
			if opts.RunDecode {
				enabled = append(enabled, a)
			}
		}
	}
	return NewSuite(enabled...)
}
