package terraform

import "testing"

func TestParse_Actions(t *testing.T) {
	tests := []struct {
		name    string
		actions string
		want    Action
	}{
		{"create", `["create"]`, ActionCreate},
		{"delete", `["delete"]`, ActionDelete},
		{"update", `["update"]`, ActionUpdate},
		{"read", `["read"]`, ActionRead},
		{"noop", `["no-op"]`, ActionNoOp},
		{"replace_delete_create", `["delete","create"]`, ActionReplace},
		{"replace_create_delete", `["create","delete"]`, ActionReplace},
		{"empty", `[]`, ActionNoOp},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			plan := []byte(`{"resource_changes":[{"address":"x.y","change":{"actions":` + tc.actions + `,"before":null,"after":null}}]}`)
			p, err := Parse(plan)
			if err != nil {
				t.Fatal(err)
			}
			if got := p.Changes[0].Action; got != tc.want {
				t.Errorf("action = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestParse_TypeNameFromFields(t *testing.T) {
	plan := []byte(`{"resource_changes":[{"address":"module.a.aws_s3_bucket.assets","type":"aws_s3_bucket","name":"assets","change":{"actions":["create"],"before":null,"after":{}}}]}`)
	p, err := Parse(plan)
	if err != nil {
		t.Fatal(err)
	}
	rc := p.Changes[0]
	if rc.Type != "aws_s3_bucket" || rc.Name != "assets" {
		t.Errorf("type/name = %q/%q, want aws_s3_bucket/assets", rc.Type, rc.Name)
	}
}

func TestParse_TypeNameFallbackFromAddress(t *testing.T) {
	tests := []struct {
		address  string
		wantType string
		wantName string
	}{
		{"aws_instance.web", "aws_instance", "web"},
		{"module.app.aws_s3_bucket.assets", "aws_s3_bucket", "assets"},
		{"single", "single", ""},
	}
	for _, tc := range tests {
		t.Run(tc.address, func(t *testing.T) {
			plan := []byte(`{"resource_changes":[{"address":"` + tc.address + `","change":{"actions":["create"],"before":null,"after":{}}}]}`)
			p, err := Parse(plan)
			if err != nil {
				t.Fatal(err)
			}
			rc := p.Changes[0]
			if rc.Type != tc.wantType || rc.Name != tc.wantName {
				t.Errorf("type/name = %q/%q, want %q/%q", rc.Type, rc.Name, tc.wantType, tc.wantName)
			}
		})
	}
}

func TestParse_InvalidJSON(t *testing.T) {
	if _, err := Parse([]byte(`{broken`)); err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestParse_Empty(t *testing.T) {
	p, err := Parse([]byte(`{"resource_changes":[]}`))
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Changes) != 0 {
		t.Errorf("got %d changes, want 0", len(p.Changes))
	}
}

func TestParse_SensitiveAsBool(t *testing.T) {
	// Terraform emits a bare `false` for after_sensitive when nothing is
	// sensitive; asObject must degrade to a nil map, not error.
	plan := []byte(`{"resource_changes":[{"address":"a.b","change":{"actions":["create"],"before":null,"after":{"x":1},"after_sensitive":false,"after_unknown":false}}]}`)
	p, err := Parse(plan)
	if err != nil {
		t.Fatal(err)
	}
	if p.Changes[0].AfterSensitive != nil || p.Changes[0].AfterUnknown != nil {
		t.Error("expected nil maps for bare-false sensitive/unknown")
	}
}

func TestCounts(t *testing.T) {
	plan := []byte(`{"resource_changes":[
		{"address":"a.1","change":{"actions":["create"],"before":null,"after":{}}},
		{"address":"a.2","change":{"actions":["create"],"before":null,"after":{}}},
		{"address":"a.3","change":{"actions":["update"],"before":{},"after":{}}},
		{"address":"a.4","change":{"actions":["delete"],"before":{},"after":null}},
		{"address":"a.5","change":{"actions":["delete","create"],"before":{},"after":{}}},
		{"address":"a.6","change":{"actions":["no-op"],"before":{},"after":{}}}
	]}`)
	p, err := Parse(plan)
	if err != nil {
		t.Fatal(err)
	}
	if p.AddCount() != 2 || p.ChangeCount() != 1 || p.DestroyCount() != 1 || p.ReplaceCount() != 1 {
		t.Errorf("counts = add:%d change:%d destroy:%d replace:%d, want 2 1 1 1",
			p.AddCount(), p.ChangeCount(), p.DestroyCount(), p.ReplaceCount())
	}
	if p.Count(ActionCreate) != 2 {
		t.Errorf("Count(ActionCreate) = %d, want 2", p.Count(ActionCreate))
	}
	if !p.HasChanges() {
		t.Error("HasChanges() = false, want true")
	}
}

func TestChanged_SkipsNoOpAndRead(t *testing.T) {
	plan := []byte(`{"resource_changes":[
		{"address":"a.1","change":{"actions":["create"],"before":null,"after":{}}},
		{"address":"a.2","change":{"actions":["no-op"],"before":{},"after":{}}},
		{"address":"a.3","change":{"actions":["read"],"before":null,"after":{}}}
	]}`)
	p, err := Parse(plan)
	if err != nil {
		t.Fatal(err)
	}
	changed := p.Changed()
	if len(changed) != 1 || changed[0].Address != "a.1" {
		t.Errorf("changed = %v, want just a.1", changed)
	}
}

func TestActionString(t *testing.T) {
	tests := map[Action]string{
		ActionNoOp:    "no-op",
		ActionCreate:  "create",
		ActionRead:    "read",
		ActionUpdate:  "update",
		ActionDelete:  "delete",
		ActionReplace: "replace",
	}
	for a, want := range tests {
		if got := a.String(); got != want {
			t.Errorf("%d.String() = %q, want %q", a, got, want)
		}
	}
}

func TestSplitAddress(t *testing.T) {
	tests := []struct {
		in       string
		wantType string
		wantName string
	}{
		{"aws_instance.web", "aws_instance", "web"},
		{"module.app.aws_s3_bucket.assets", "aws_s3_bucket", "assets"},
		{"lonely", "lonely", ""},
	}
	for _, tc := range tests {
		typ, name := splitAddress(tc.in)
		if typ != tc.wantType || name != tc.wantName {
			t.Errorf("splitAddress(%q) = %q/%q, want %q/%q", tc.in, typ, name, tc.wantType, tc.wantName)
		}
	}
}

// identical before/after must render as no changes at all.
func TestRender_Identical(t *testing.T) {
	plan := []byte(`{"resource_changes":[{"address":"a.b","type":"a","name":"b","change":{"actions":["no-op"],"before":{"x":1},"after":{"x":1}}}]}`)
	p, err := Parse(plan)
	if err != nil {
		t.Fatal(err)
	}
	if p.HasChanges() {
		t.Error("HasChanges() = true for a no-op plan, want false")
	}
	if out := p.RenderDiff(RenderOptions{}); out != "" {
		t.Errorf("expected empty diff, got:\n%s", out)
	}
}
