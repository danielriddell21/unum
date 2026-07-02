package terraform

import (
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
