package engine

import "testing"

func TestD2ID(t *testing.T) {
	if got := d2ID("Client_1"); got != "Client_1" {
		t.Errorf("d2ID(Client_1) = %q, want unquoted", got)
	}
	if got := d2ID("has space"); got != `"has space"` {
		t.Errorf("d2ID(has space) = %q, want quoted", got)
	}
	for _, kw := range []string{"style", "shape", "label", "vars"} {
		if got := d2ID(kw); got != `"`+kw+`"` {
			t.Errorf("d2ID(%s) = %q, want quoted reserved keyword", kw, got)
		}
	}
}
