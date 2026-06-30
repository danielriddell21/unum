package analyze_test

import (
	"context"
	"errors"
	"testing"

	"github.com/danielriddell21/unum/internal/json/analyze"
	"github.com/danielriddell21/unum/internal/json/node"
	"github.com/danielriddell21/unum/internal/json/parse"
)

type okAnalyzer struct{ name string }

func (a *okAnalyzer) Name() string { return a.name }
func (a *okAnalyzer) Run(_ context.Context, _ *node.Node, _ analyze.Options) error {
	return nil
}

type errAnalyzer struct {
	name string
	err  error
}

func (a *errAnalyzer) Name() string { return a.name }
func (a *errAnalyzer) Run(_ context.Context, _ *node.Node, _ analyze.Options) error {
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
	s := analyze.NewSuite(&okAnalyzer{"a"}, &okAnalyzer{"b"})
	root := mustParse(t, `{"x": 1}`)
	if err := s.Run(context.Background(), root, analyze.Options{}); err != nil {
		t.Errorf("all-ok suite: unexpected error: %v", err)
	}
}

func TestSuiteRun_OneError(t *testing.T) {
	boom := errors.New("boom")
	s := analyze.NewSuite(&okAnalyzer{"ok"}, &errAnalyzer{"stats", boom})
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
	s := analyze.NewSuite(&errAnalyzer{"stats", e1}, &errAnalyzer{"merkle", e2})
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
		analyzers[i] = &okAnalyzer{"a"}
	}
	s := analyze.NewSuite(analyzers...)
	root := mustParse(t, `{"k": 1}`)
	if err := s.Run(context.Background(), root, analyze.Options{}); err != nil {
		t.Errorf("concurrent run: %v", err)
	}
}

func TestBuild_EnabledByOption(t *testing.T) {
	all := []analyze.Analyzer{
		&okAnalyzer{"stats"},
		&okAnalyzer{"merkle"},
		&okAnalyzer{"decode"},
	}

	opts := analyze.Options{RunStats: true, RunMerkle: true}
	s := analyze.Build(opts, all)
	root := mustParse(t, `{}`)
	if err := s.Run(context.Background(), root, opts); err != nil {
		t.Errorf("Build with stats+merkle: %v", err)
	}
}

func TestBuild_NoneEnabled(t *testing.T) {
	all := []analyze.Analyzer{
		&okAnalyzer{"stats"},
		&okAnalyzer{"merkle"},
	}
	s := analyze.Build(analyze.Options{}, all)
	root := mustParse(t, `{}`)
	if err := s.Run(context.Background(), root, analyze.Options{}); err != nil {
		t.Errorf("Build with nothing enabled: %v", err)
	}
}

func TestBuild_DecodeEnabled(t *testing.T) {
	all := []analyze.Analyzer{&okAnalyzer{"decode"}}
	s := analyze.Build(analyze.Options{RunDecode: true}, all)
	root := mustParse(t, `{}`)
	if err := s.Run(context.Background(), root, analyze.Options{RunDecode: true}); err != nil {
		t.Errorf("Build with decode: %v", err)
	}
}

func TestBuild_UnknownAnalyzerSkipped(t *testing.T) {
	// An analyzer with an unknown name is simply not included.
	all := []analyze.Analyzer{&okAnalyzer{"custom"}}
	s := analyze.Build(analyze.Options{RunStats: true}, all)
	root := mustParse(t, `{}`)
	if err := s.Run(context.Background(), root, analyze.Options{}); err != nil {
		t.Errorf("Build with unknown analyzer: %v", err)
	}
}
