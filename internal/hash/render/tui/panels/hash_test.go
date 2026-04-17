package panels

import (
	"strings"
	"testing"

	"github.com/danielriddell21/unum/internal/hash/types"
)

func TestHashPanel_ViewNilResult(t *testing.T) {
	p := NewHashPanel(80, 20)
	out := p.View("input-widget", nil)
	if !strings.Contains(out, "type something") {
		t.Errorf("nil result: want 'type something' hint, got %q", out)
	}
	if !strings.Contains(out, "input-widget") {
		t.Errorf("nil result: want inputView in output, got %q", out)
	}
}

func TestHashPanel_ViewWithResult(t *testing.T) {
	p := NewHashPanel(80, 20)
	r := &types.Result{Input: "my-svc", Port: 8080, UUID: "u1", Color: "#fff", Short: "ab", Emoji: "🦊", Phrase: "alpha"}
	out := p.View("input", r)
	if !strings.Contains(out, "my-svc") {
		t.Errorf("result: want 'my-svc' in %q", out)
	}
	if !strings.Contains(out, "8080") {
		t.Errorf("result: want port '8080' in %q", out)
	}
}

func TestHashPanel_Resize(t *testing.T) {
	p := NewHashPanel(80, 20)
	p.Resize(100, 30)
	out := p.View("x", nil)
	if out == "" {
		t.Error("after resize: View returned empty")
	}
}
