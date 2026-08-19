package terraform

import (
	"fmt"
	"strings"
	"testing"
)

func render(t *testing.T, planJSON string) (markdown, terminal string) {
	t.Helper()
	p, err := Parse([]byte(planJSON))
	if err != nil {
		t.Fatal(err)
	}
	return p.RenderDiff(RenderOptions{MarkerFirst: true}), p.RenderDiff(RenderOptions{})
}

func TestRender_Create(t *testing.T) {
	md, gutter := render(t, `{"resource_changes":[{"address":"aws_instance.web","type":"aws_instance","name":"web","change":{"actions":["create"],"before":null,"after":{"ami":"ami-123","instance_type":"t3.micro"}}}]}`)
	wants := []string{
		"# aws_instance.web will be created",
		`resource "aws_instance" "web" {`,
		"ami           = \"ami-123\"",
		"instance_type = \"t3.micro\"",
	}
	for _, w := range wants {
		if !strings.Contains(md, w) {
			t.Errorf("MarkerFirst diff missing %q\n%s", w, md)
		}
	}
	// MarkerFirst places the '+' in column 0 so GitHub colours the line.
	if !strings.Contains(md, "+   resource \"aws_instance\" \"web\" {") {
		t.Errorf("expected column-0 '+' marker\n%s", md)
	}
	// The gutter (default) style indents the marker instead.
	if !strings.Contains(gutter, "  + resource \"aws_instance\" \"web\" {") {
		t.Errorf("expected indented gutter marker\n%s", gutter)
	}
	// RenderDiff is body-only in both styles: no preamble, no footer.
	for _, out := range []string{md, gutter} {
		if strings.Contains(out, "Plan:") || strings.Contains(out, "Terraform will perform") {
			t.Errorf("RenderDiff should not carry preamble/footer\n%s", out)
		}
	}
}

func TestRender_UpdateMarkerToggle(t *testing.T) {
	p, err := Parse([]byte(`{"resource_changes":[{"address":"a.b","type":"a","name":"b","change":{"actions":["update"],"before":{"x":"old"},"after":{"x":"new"}}}]}`))
	if err != nil {
		t.Fatal(err)
	}
	// Gutter (default) style keeps terraform's '~'.
	if gutter := p.RenderDiff(RenderOptions{}); !strings.Contains(gutter, "~ ") || strings.Contains(gutter, "! ") {
		t.Errorf("gutter update should use '~'\n%s", gutter)
	}
	// MarkerFirst alone keeps '~'.
	if tilde := p.RenderDiff(RenderOptions{MarkerFirst: true}); !strings.Contains(tilde, "~ ") || strings.Contains(tilde, "! ") {
		t.Errorf("MarkerFirst without BangUpdates should use '~'\n%s", tilde)
	}
	// BangUpdates switches to '!' so diff highlighters colour the line.
	if bang := p.RenderDiff(RenderOptions{MarkerFirst: true, BangUpdates: true}); !strings.Contains(bang, "! ") || strings.Contains(bang, "~ ") {
		t.Errorf("BangUpdates should use '!'\n%s", bang)
	}
}

func TestRender_Delete(t *testing.T) {
	md, _ := render(t, `{"resource_changes":[{"address":"aws_instance.legacy","type":"aws_instance","name":"legacy","change":{"actions":["delete"],"before":{"ami":"ami-old"},"after":null}}]}`)
	for _, w := range []string{"# aws_instance.legacy will be destroyed", `ami = "ami-old" -> null`} {
		if !strings.Contains(md, w) {
			t.Errorf("markdown missing %q\n%s", w, md)
		}
	}
}

func TestRender_UpdateInPlace(t *testing.T) {
	md, _ := render(t, `{"resource_changes":[{"address":"aws_s3_bucket.assets","type":"aws_s3_bucket","name":"assets","change":{"actions":["update"],"before":{"versioning":{"enabled":false}},"after":{"versioning":{"enabled":true}}}}]}`)
	for _, w := range []string{"# aws_s3_bucket.assets will be updated in-place", "versioning {", "enabled = false -> true"} {
		if !strings.Contains(md, w) {
			t.Errorf("markdown missing %q\n%s", w, md)
		}
	}
}

func TestRender_Replace(t *testing.T) {
	md, _ := render(t, `{"resource_changes":[{"address":"aws_db.main","type":"aws_db","name":"main","change":{"actions":["delete","create"],"before":{"engine":"postgres"},"after":{"engine":"mysql"}}}]}`)
	for _, w := range []string{"# aws_db.main must be replaced", `engine = "postgres" -> "mysql"`} {
		if !strings.Contains(md, w) {
			t.Errorf("markdown missing %q\n%s", w, md)
		}
	}
}

func TestRender_KnownAfterApply(t *testing.T) {
	md, _ := render(t, `{"resource_changes":[{"address":"aws_instance.web","type":"aws_instance","name":"web","change":{"actions":["create"],"before":null,"after":{"arn":null},"after_unknown":{"arn":true}}}]}`)
	if !strings.Contains(md, "arn = (known after apply)") {
		t.Errorf("expected known-after-apply\n%s", md)
	}
}

func TestRender_KnownAfterApplyUpdate(t *testing.T) {
	md, _ := render(t, `{"resource_changes":[{"address":"a.b","type":"a","name":"b","change":{"actions":["update"],"before":{"id":"old"},"after":{"id":null},"after_unknown":{"id":true}}}]}`)
	if !strings.Contains(md, `id = "old" -> (known after apply)`) {
		t.Errorf("expected old -> known-after-apply\n%s", md)
	}
}

func TestRender_UnchangedHidden(t *testing.T) {
	md, _ := render(t, `{"resource_changes":[{"address":"a.b","type":"a","name":"b","change":{"actions":["update"],"before":{"changed":"x","same1":"y","same2":"z"},"after":{"changed":"X","same1":"y","same2":"z"}}}]}`)
	if !strings.Contains(md, "# (2 unchanged attributes hidden)") {
		t.Errorf("expected 2 unchanged attributes hidden\n%s", md)
	}
	// Singular form.
	md2, _ := render(t, `{"resource_changes":[{"address":"a.b","type":"a","name":"b","change":{"actions":["update"],"before":{"changed":"x","same1":"y"},"after":{"changed":"X","same1":"y"}}}]}`)
	if !strings.Contains(md2, "# (1 unchanged attribute hidden)") {
		t.Errorf("expected singular form\n%s", md2)
	}
}

func TestRender_Sensitive(t *testing.T) {
	md, _ := render(t, `{"resource_changes":[{"address":"aws_db.main","type":"aws_db","name":"main","change":{"actions":["update"],"before":{"pw":"hunter2"},"after":{"pw":"swordfish"},"before_sensitive":{"pw":true},"after_sensitive":{"pw":true}}}]}`)
	if !strings.Contains(md, "pw = (sensitive value) -> (sensitive value)") {
		t.Errorf("expected sensitive masking\n%s", md)
	}
	if strings.Contains(md, "hunter2") || strings.Contains(md, "swordfish") {
		t.Errorf("sensitive value leaked\n%s", md)
	}
}

func TestRender_ListDiff(t *testing.T) {
	md, _ := render(t, `{"resource_changes":[{"address":"a.b","type":"a","name":"b","change":{"actions":["update"],"before":{"tags":["keep","old"]},"after":{"tags":["keep","new"]}}}]}`)
	for _, w := range []string{"tags = [", `"keep",`, `"old",`, `"new",`} {
		if !strings.Contains(md, w) {
			t.Errorf("markdown missing %q\n%s", w, md)
		}
	}
}

func TestRender_ListAdded(t *testing.T) {
	md, _ := render(t, `{"resource_changes":[{"address":"a.b","type":"a","name":"b","change":{"actions":["create"],"before":null,"after":{"tags":["a","b"]}}}]}`)
	for _, w := range []string{"tags = [", `"a",`, `"b",`} {
		if !strings.Contains(md, w) {
			t.Errorf("markdown missing %q\n%s", w, md)
		}
	}
}

func TestRender_NoChanges(t *testing.T) {
	md, gutter := render(t, `{"resource_changes":[{"address":"a.b","type":"a","name":"b","change":{"actions":["no-op"],"before":{"x":1},"after":{"x":1}}}]}`)
	if md != "" || gutter != "" {
		t.Errorf("RenderDiff should be empty when there are no changes, got md=%q gutter=%q", md, gutter)
	}
}

func TestRender_NoANSI(t *testing.T) {
	// The package is colour-free: no render option produces ANSI escapes.
	p, err := Parse([]byte(`{"resource_changes":[{"address":"a.b","type":"a","name":"b","change":{"actions":["create"],"before":null,"after":{"x":"y"}}}]}`))
	if err != nil {
		t.Fatal(err)
	}
	for _, opts := range []RenderOptions{{}, {MarkerFirst: true}, {MarkerFirst: true, BangUpdates: true}} {
		if out := p.RenderDiff(opts); strings.Contains(out, "\x1b[") {
			t.Errorf("RenderDiff(%+v) leaked ANSI escapes:\n%s", opts, out)
		}
	}
}

func TestRender_MultipleResourcesSeparated(t *testing.T) {
	md, _ := render(t, `{"resource_changes":[
		{"address":"a.1","type":"a","name":"1","change":{"actions":["create"],"before":null,"after":{"x":1}}},
		{"address":"a.2","type":"a","name":"2","change":{"actions":["create"],"before":null,"after":{"x":1}}}
	]}`)
	if strings.Count(md, "resource ") != 2 {
		t.Errorf("expected 2 resource blocks\n%s", md)
	}
}

func listPlan(before, after string) string {
	return `{"resource_changes":[{"address":"auth0_resource_server.api","type":"auth0_resource_server","name":"api","change":{"actions":["update"],"before":{"id":"api","scopes":` +
		before + `},"after":{"id":"api","scopes":` + after + `}}}]}`
}

func markerCount(out, marker string) int {
	n := 0
	for _, line := range strings.Split(out, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), marker+" ") {
			n++
		}
	}
	return n
}

func TestRender_ListAlignment(t *testing.T) {
	tests := []struct {
		name           string
		before, after  string
		wantAdded      int
		wantRemoved    int
		want, notWant  []string
		wantAttrHidden bool
	}{
		{
			name:           "identical lists are hidden",
			before:         `["read:users","write:users"]`,
			after:          `["read:users","write:users"]`,
			wantAttrHidden: true,
		},
		{
			name:           "reorder only is hidden",
			before:         `["read:users","create:users","delete:users"]`,
			after:          `["create:users","delete:users","read:users"]`,
			wantAttrHidden: true,
		},
		{
			name:        "insert at front touches one line",
			before:      `["create:users","delete:users","read:users"]`,
			after:       `["admin:all","create:users","delete:users","read:users"]`,
			wantAdded:   1,
			wantRemoved: 0,
			want:        []string{`+ "admin:all",`},
			notWant:     []string{`- "create:users",`, `- "read:users",`},
		},
		{
			name:        "insert in the middle touches one line",
			before:      `["create:users","delete:users","update:users"]`,
			after:       `["create:users","delete:users","read:reports","update:users"]`,
			wantAdded:   1,
			wantRemoved: 0,
			want:        []string{`+ "read:reports",`},
			notWant:     []string{`- "update:users",`},
		},
		{
			name:      "insert at the end touches one line",
			before:    `["create:users"]`,
			after:     `["create:users","delete:users"]`,
			wantAdded: 1,
			want:      []string{`+ "delete:users",`},
		},
		{
			name:        "removal from the middle touches one line",
			before:      `["create:users","delete:users","read:users"]`,
			after:       `["create:users","read:users"]`,
			wantRemoved: 1,
			want:        []string{`- "delete:users",`},
			notWant:     []string{`- "read:users",`},
		},
		{
			name:        "in-place replacement shows old and new",
			before:      `["read:users"]`,
			after:       `["read:reports"]`,
			wantAdded:   1,
			wantRemoved: 1,
			want:        []string{`- "read:users",`, `+ "read:reports",`},
		},
		{
			name:        "object elements align on content",
			before:      `[{"value":"create:users","description":"Create"},{"value":"read:users","description":"Read"}]`,
			after:       `[{"value":"create:users","description":"Create"},{"value":"read:reports","description":"Reports"},{"value":"read:users","description":"Read"}]`,
			wantAdded:   1,
			wantRemoved: 0,
			want:        []string{`+ {"description":"Reports","value":"read:reports"},`},
			notWant:     []string{`- {"description":"Read","value":"read:users"},`},
		},
		{
			name:        "list emptied entirely",
			before:      `["read:users","write:users"]`,
			after:       `[]`,
			wantRemoved: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, gutter := render(t, listPlan(tt.before, tt.after))
			if tt.wantAttrHidden {
				if strings.Contains(gutter, "scopes") {
					t.Errorf("unchanged list should be hidden\n%s", gutter)
				}
				if !strings.Contains(gutter, "(2 unchanged attributes hidden)") {
					t.Errorf("expected both attributes hidden\n%s", gutter)
				}
				return
			}
			if got := markerCount(gutter, "+"); got != tt.wantAdded {
				t.Errorf("added lines = %d, want %d\n%s", got, tt.wantAdded, gutter)
			}
			if got := markerCount(gutter, "-"); got != tt.wantRemoved {
				t.Errorf("removed lines = %d, want %d\n%s", got, tt.wantRemoved, gutter)
			}
			for _, w := range tt.want {
				if !strings.Contains(gutter, w) {
					t.Errorf("output missing %q\n%s", w, gutter)
				}
			}
			for _, w := range tt.notWant {
				if strings.Contains(gutter, w) {
					t.Errorf("output should not contain %q\n%s", w, gutter)
				}
			}
		})
	}
}

func TestRender_ListAlignmentFallsBackWhenHuge(t *testing.T) {
	var before, after strings.Builder
	before.WriteByte('[')
	after.WriteByte('[')
	for i := range 600 {
		if i > 0 {
			before.WriteByte(',')
			after.WriteByte(',')
		}
		fmt.Fprintf(&before, `"scope:%d"`, i)
		fmt.Fprintf(&after, `"scope:%d"`, i)
	}
	before.WriteByte(']')
	after.WriteString(`,"scope:new"]`)

	_, gutter := render(t, listPlan(before.String(), after.String()))
	if !strings.Contains(gutter, `+ "scope:new",`) {
		t.Errorf("oversized list should still render the new element\n%s", gutter[:min(400, len(gutter))])
	}
}
