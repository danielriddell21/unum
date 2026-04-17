package panels

import (
	"testing"

	"github.com/danielriddell21/unum/internal/json/parse"
)

func mustParseLens(t *testing.T, src string) *LensPanel {
	t.Helper()
	n, err := parse.Parse([]byte(src))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	p := NewLensPanel(n, 80, 20)
	return &p
}

func TestLensPanel_ViewNoPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("View panicked: %v", r)
		}
	}()
	p := mustParseLens(t, `{"name": "alice"}`)
	_ = p.View()
}

func TestLensPanel_DefaultActiveLens(t *testing.T) {
	p := mustParseLens(t, `{"x": 1}`)
	if p.ActiveLens() != LensTypeGen {
		t.Errorf("default lens=%v, want LensTypeGen", p.ActiveLens())
	}
}

func TestLensPanel_HandleKeyDigitsSwitchLens(t *testing.T) {
	p := mustParseLens(t, `{"x": 1}`)
	keys := []struct {
		key  string
		want LensID
	}{
		{"1", LensTypeGen},
		{"2", LensSchema},
		{"3", LensYAML},
		{"4", LensMerkle},
		{"5", LensStats},
		{"6", LensQuery},
	}
	for _, tc := range keys {
		p.HandleKey(tc.key)
		if p.ActiveLens() != tc.want {
			t.Errorf("HandleKey(%q): active=%v, want %v", tc.key, p.ActiveLens(), tc.want)
		}
	}
}

func TestLensPanel_HandleKeyViewNoPanic(t *testing.T) {
	p := mustParseLens(t, `{"a": 1}`)
	for _, k := range []string{"1", "2", "3", "4", "5", "6"} {
		p.HandleKey(k)
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("View panicked after key %q: %v", k, r)
			}
		}()
		_ = p.View()
	}
}

func TestLensPanel_SetCursorNode(t *testing.T) {
	root, _ := parse.Parse([]byte(`{"child": {"k": 1}}`))
	p := NewLensPanel(root, 80, 20)
	p.SetCursorNode(root.Children[0])
	_ = p.View()
}

func TestLensPanel_SetCursorNil(t *testing.T) {
	p := mustParseLens(t, `{"k": 1}`)
	p.SetCursorNode(nil)
	_ = p.View()
}

func TestLensPanel_SetFocused(t *testing.T) {
	p := mustParseLens(t, `{"k": 1}`)
	p.SetFocused(true)
	_ = p.View()
	p.SetFocused(false)
	_ = p.View()
}

func TestLensPanel_Resize(t *testing.T) {
	p := mustParseLens(t, `{"k": 1}`)
	p.Resize(100, 30)
	_ = p.View()
}

func TestLensPanel_PrettyNode(t *testing.T) {
	n, _ := parse.Parse([]byte(`{"k": "v"}`))
	out := PrettyNode(n)
	if out == "" {
		t.Error("PrettyNode returned empty string for object node")
	}
}

func TestLensPanel_QuerySubMode(t *testing.T) {
	p := mustParseLens(t, `{"k": 1}`)
	// Switch to Query lens
	p.HandleKey("6")
	// Enter query expression sub-mode (if supported by HandleKey)
	p.HandleKey("enter")
	_ = p.View()
}
