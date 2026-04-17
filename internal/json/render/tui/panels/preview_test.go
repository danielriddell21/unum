package panels

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/danielriddell21/unum/internal/json/parse"
)

func TestPreviewPanel_ViewNoPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("View panicked: %v", r)
		}
	}()
	p := NewPreviewPanel(80, 20)
	_ = p.View()
}

func TestPreviewPanel_SetNilNode(t *testing.T) {
	p := NewPreviewPanel(80, 20)
	p.SetNode(nil)
	_ = p.View()
}

func TestPreviewPanel_SetObjectNode(t *testing.T) {
	n, err := parse.Parse([]byte(`{"name": "alice", "age": 30}`))
	if err != nil {
		t.Fatal(err)
	}
	p := NewPreviewPanel(80, 20)
	p.SetNode(n)
	_ = p.View()
}

func TestPreviewPanel_SetLeafNodes(t *testing.T) {
	for _, src := range []string{`"hello"`, `42`, `true`, `null`} {
		n, err := parse.Parse([]byte(src))
		if err != nil {
			t.Fatalf("parse(%q): %v", src, err)
		}
		p := NewPreviewPanel(80, 20)
		p.SetNode(n)
		_ = p.View()
	}
}

func TestPreviewPanel_SetArrayNode(t *testing.T) {
	n, err := parse.Parse([]byte(`[1, "two", null]`))
	if err != nil {
		t.Fatal(err)
	}
	p := NewPreviewPanel(80, 20)
	p.SetNode(n)
	_ = p.View()
}

func TestPreviewPanel_HandleKeyScroll(t *testing.T) {
	n, err := parse.Parse([]byte(`{"k": "v"}`))
	if err != nil {
		t.Fatal(err)
	}
	p := NewPreviewPanel(80, 20)
	p.SetNode(n)
	// Should not panic on scroll keys.
	for _, key := range []string{"j", "k", "ctrl+d", "ctrl+u"} {
		p.HandleKey(key)
	}
}

func TestPreviewPanel_Resize(t *testing.T) {
	p := NewPreviewPanel(80, 20)
	p.SetFocused(true)
	p.Resize(100, 30)
	_ = p.View()
}

func TestPreviewPanel_Update(t *testing.T) {
	n, err := parse.Parse([]byte(`{"a": 1}`))
	if err != nil {
		t.Fatal(err)
	}
	p := NewPreviewPanel(80, 20)
	p.SetNode(n)
	p.HandleKey(tea.KeyMsg{Type: tea.KeyDown}.String())
	_ = p.View()
}
