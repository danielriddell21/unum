package hash

import (
	"regexp"
	"strings"
	"testing"
)

var uuidRe = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-5[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
var colorRe = regexp.MustCompile(`^#[0-9a-f]{6}$`)

func TestDerive_Deterministic(t *testing.T) {
	r1 := Derive("my-api-service")
	r2 := Derive("my-api-service")
	if r1 != r2 {
		t.Errorf("Derive is not deterministic:\n  r1=%+v\n  r2=%+v", r1, r2)
	}
}

func TestDerive_DifferentInputs(t *testing.T) {
	r1 := Derive("service-a")
	r2 := Derive("service-b")
	if r1.Port == r2.Port && r1.UUID == r2.UUID && r1.Color == r2.Color {
		t.Error("different inputs produced identical results")
	}
}

func TestDerive_EmptyInput(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Derive panicked on empty input: %v", r)
		}
	}()
	r := Derive("")
	if r.Port < 1024 {
		t.Errorf("port %d below minimum 1024", r.Port)
	}
}

func TestDerive_Port(t *testing.T) {
	inputs := []string{"hello", "world", "foo", "bar", "my-service", "database", "redis"}
	for _, input := range inputs {
		r := Derive(input)
		if r.Port < 1024 {
			t.Errorf("input %q: port %d below minimum 1024", input, r.Port)
		}
	}
}

func TestDerive_UUID(t *testing.T) {
	r := Derive("test-service")
	if !uuidRe.MatchString(r.UUID) {
		t.Errorf("UUID %q does not match v5 pattern", r.UUID)
	}
}

func TestDerive_Color(t *testing.T) {
	r := Derive("test-service")
	if !colorRe.MatchString(r.Color) {
		t.Errorf("color %q does not match #rrggbb pattern", r.Color)
	}
}

func TestDerive_Short(t *testing.T) {
	r := Derive("test-service")
	if len(r.Short) != 8 {
		t.Errorf("short %q: want 8 chars, got %d", r.Short, len(r.Short))
	}
	matched, _ := regexp.MatchString(`^[0-9a-f]{8}$`, r.Short)
	if !matched {
		t.Errorf("short %q is not lowercase hex", r.Short)
	}
}

func TestDerive_Phrase(t *testing.T) {
	r := Derive("test-service")
	parts := strings.Split(r.Phrase, "-")
	if len(parts) != 3 {
		t.Errorf("phrase %q: want 3 words, got %d", r.Phrase, len(parts))
	}
	for _, p := range parts {
		if p == "" {
			t.Errorf("phrase %q has empty word", r.Phrase)
		}
	}
}

func TestDerive_Emoji(t *testing.T) {
	r := Derive("test-service")
	if r.Emoji == "" {
		t.Error("emoji should not be empty")
	}
}

func TestDerive_InputPreserved(t *testing.T) {
	input := "my-api-service"
	r := Derive(input)
	if r.Input != input {
		t.Errorf("Input field: got %q, want %q", r.Input, input)
	}
}
