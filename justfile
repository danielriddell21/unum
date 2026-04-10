# unum — justfile
# Install just: https://github.com/casey/just

binary := "bin/unum"
main := "./cmd/unum"

# ── Core ──────────────────────────────────────────────────────────────────────

# build to bin/unum
[group('core')]
build:
    go build -ldflags "-X main.version=dev" -o {{ binary }} {{ main }}

# run unit tests — pass extra flags e.g. `just test -v` or `just test -run TestParse`
[group('core')]
test *args:
    go test {{ args }} ./...

# run CLI functional tests
[group('core')]
yeet:
    go test -v ./tests/...

# golangci-lint
[group('core')]
lint:
    golangci-lint run ./...

# install to GOPATH/bin
[group('core')]
install:
    go install {{ main }}

# remove build artifacts
[group('core')]
clean:
    rm -rf bin/

# full gate: lint + test + yeet + build
[group('core')]
ci: lint test yeet build

# tag and push a release
[group('core')]
release tag:
    git tag {{ tag }}
    git push origin {{ tag }}

# ── JSON ──────────────────────────────────────────────────────────────────────

# static render
[group('json')]
json file="testdata/sample.json":
    go run {{ main }} json {{ file }}

# TUI navigator
[group('json')]
json-ui file="testdata/sample.json":
    go run {{ main }} json {{ file }} --ui

# browser UI
[group('json')]
json-web file="testdata/sample.json":
    go run {{ main }} json {{ file }} --web

# --stats annotation
[group('json')]
stats file="testdata/sample.json":
    go run {{ main }} json {{ file }} --stats --no-color

# --merkle --hash-only
[group('json')]
hash file="testdata/sample.json":
    go run {{ main }} json {{ file }} --merkle --hash-only

# Go struct generation
[group('json')]
typegen-go file="testdata/sample.json":
    go run {{ main }} json {{ file }} --typegen go

# TypeScript interface generation
[group('json')]
typegen-ts file="testdata/sample.json":
    go run {{ main }} json {{ file }} --typegen ts

# JSON Schema generation
[group('json')]
schema file="testdata/sample.json":
    go run {{ main }} json {{ file }} --typegen jsonschema

# transform to YAML
[group('json')]
yaml file="testdata/sample.json":
    go run {{ main }} json {{ file }} --transform

# ── Diff ──────────────────────────────────────────────────────────────────────

# text diff
[group('diff')]
diff file-a="testdata/diff-a.txt" file-b="testdata/diff-b.txt":
    go run {{ main }} diff {{ file-a }} {{ file-b }}

# TUI diff
[group('diff')]
diff-ui file-a="testdata/diff-a.json" file-b="testdata/diff-b.json":
    go run {{ main }} diff {{ file-a }} {{ file-b }} --ui

# browser diff
[group('diff')]
diff-web file-a="testdata/diff-a.json" file-b="testdata/diff-b.json":
    go run {{ main }} diff {{ file-a }} {{ file-b }} --web

# YAML semantic diff
[group('diff')]
diff-yaml file-a="testdata/diff-a.yaml" file-b="testdata/diff-b.yaml":
    go run {{ main }} diff {{ file-a }} {{ file-b }}

# Terraform plan diff
[group('diff')]
diff-tf file="testdata/diff-a.tfplan.json":
    go run {{ main }} diff {{ file }} {{ file }} --format terraform
