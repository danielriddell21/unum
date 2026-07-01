package analyze_test

import (
	"context"
	"errors"
	"testing"

	"github.com/danielriddell21/unum/internal/json/analyze"
	"github.com/danielriddell21/unum/internal/json/node"
	"github.com/danielriddell21/unum/internal/json/parse"
)

type okAnalyzer struct {
	name string
	on   bool
}

func (a *okAnalyzer) Name() string                   { return a.name }
func (a *okAnalyzer) Enabled(_ analyze.Options) bool { return a.on }
func (a *okAnalyzer) Run(context.Context, *node.Node, analyze.Options) error {
	return nil
}

type errAnalyzer struct {
	name string
	on   bool
	err  error
}

func (a *errAnalyzer) Name() string                   { return a.name }
func (a *errAnalyzer) Enabled(_ analyze.Options) bool { return a.on }
func (a *errAnalyzer) Run(context.Context, *node.Node, analyze.Options) error {
	return a.err
}

func mustParse(t *testing.T, src string) *node.Node {
	t.Helper()
	n, err := parse.Parse([]byte(src))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	return n
}

func TestSuiteRun_Empty(t *testing.T) {
	s := analyze.NewSuite()
	root := mustParse(t, `{}`)
	if err := s.Run(context.Background(), root, analyze.Options{}); err != nil {
		t.Errorf("empty suite should return nil, got %v", err)
	}
}

func TestSuiteRun_HappyPath(t *testing.T) {
	s := analyze.NewSuite(&okAnalyzer{name: "a"}, &okAnalyzer{name: "b"})
	root := mustParse(t, `{"x": 1}`)
	if err := s.Run(context.Background(), root, analyze.Options{}); err != nil {
		t.Errorf("all-ok suite: unexpected error: %v", err)
	}
}

func TestSuiteRun_OneError(t *testing.T) {
	boom := errors.New("boom")
	s := analyze.NewSuite(&okAnalyzer{name: "ok"}, &errAnalyzer{name: "stats", err: boom})
	root := mustParse(t, `{}`)
	err := s.Run(context.Background(), root, analyze.Options{})
	if err == nil {
		t.Fatal("expected error from failing analyzer")
	}
	if !errors.Is(err, boom) {
		t.Errorf("error should wrap original: %v", err)
	}
}

func TestSuiteRun_MultipleErrors(t *testing.T) {
	e1 := errors.New("e1")
	e2 := errors.New("e2")
	s := analyze.NewSuite(&errAnalyzer{name: "stats", err: e1}, &errAnalyzer{name: "merkle", err: e2})
	root := mustParse(t, `{}`)
	err := s.Run(context.Background(), root, analyze.Options{})
	if err == nil {
		t.Fatal("expected errors from all failing analyzers")
	}
	if !errors.Is(err, e1) || !errors.Is(err, e2) {
		t.Errorf("error should include both: %v", err)
	}
}

func TestSuiteRun_Concurrent(t *testing.T) {
	// Run many analyzers to verify no data races (run with -race).
	analyzers := make([]analyze.Analyzer, 20)
	for i := range analyzers {
		analyzers[i] = &okAnalyzer{name: "a"}
	}
	s := analyze.NewSuite(analyzers...)
	root := mustParse(t, `{"k": 1}`)
	if err := s.Run(context.Background(), root, analyze.Options{}); err != nil {
		t.Errorf("concurrent run: %v", err)
	}
}

func TestBuild_EnabledByOption(t *testing.T) {
	boom := errors.New("boom")
	// The disabled analyzer would fail if run; Build must exclude it.
	all := []analyze.Analyzer{
		&okAnalyzer{name: "stats", on: true},
		&errAnalyzer{name: "merkle", on: false, err: boom},
	}
	opts := analyze.Options{RunStats: true}
	s := analyze.Build(opts, all)
	root := mustParse(t, `{}`)
	if err := s.Run(context.Background(), root, opts); err != nil {
		t.Errorf("Build should have excluded the disabled analyzer, got %v", err)
	}
}

func TestBuild_NoneEnabled(t *testing.T) {
	boom := errors.New("boom")
	all := []analyze.Analyzer{
		&errAnalyzer{name: "stats", on: false, err: boom},
		&errAnalyzer{name: "merkle", on: false, err: boom},
	}
	s := analyze.Build(analyze.Options{}, all)
	root := mustParse(t, `{}`)
	if err := s.Run(context.Background(), root, analyze.Options{}); err != nil {
		t.Errorf("Build with nothing enabled: %v", err)
	}
}
