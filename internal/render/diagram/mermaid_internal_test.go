package diagram

import "testing"

func TestD2ID(t *testing.T) {
	if got := d2ID("Client_1"); got != "Client_1" {
		t.Errorf("d2ID(Client_1) = %q, want unquoted", got)
	}
	if got := d2ID("has space"); got != `"has space"` {
		t.Errorf("d2ID(has space) = %q, want quoted", got)
	}
}
