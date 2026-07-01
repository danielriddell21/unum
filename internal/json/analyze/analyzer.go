package analyze

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/danielriddell21/unum/internal/json/node"
)

type Analyzer interface {
	Name() string
	Enabled(opts Options) bool
	Run(ctx context.Context, root *node.Node, opts Options) error
}

type Options struct {
	RunStats  bool
	RunMerkle bool
}

type Suite struct {
	analyzers []Analyzer
}

func NewSuite(analyzers ...Analyzer) *Suite {
	return &Suite{analyzers: analyzers}
}

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

func Build(opts Options, all []Analyzer) *Suite {
	var enabled []Analyzer
	for _, a := range all {
		if a.Enabled(opts) {
			enabled = append(enabled, a)
		}
	}
	return NewSuite(enabled...)
}
