# unum — justfile
# Install just: https://github.com/casey/just

binary := "bin/unum"
main := "./cmd/unum"

# ── Core ──────────────────────────────────────────────────────────────────────

# list available recipes
default:
    @just --list

# build to bin/unum
[group('build')]
build:
    go build -ldflags "-X main.version=dev" -o {{ binary }} {{ main }}

# run unit tests — pass extra flags e.g. `just test -v` or `just test -run TestParse`
[group('test')]
test *args:
    go test {{ args }} ./...

# golangci-lint
[group('dev')]
lint *args:
    golangci-lint run ./... {{ args }}

# format the code
[group('dev')]
fmt:
    golangci-lint fmt

# tidy module dependencies
[group('dev')]
tidy:
    go mod tidy

# full gate: lint + test + yeet + build. all must pass before committing
[group('dev')]
ci: lint test yeet build

# run CLI functional tests
[group('test')]
yeet:
    go test -v ./tests/...

# install to GOPATH/bin
[group('build')]
install:
    go install {{ main }}

# remove build artifacts
[group('dev')]
clean:
    rm -rf bin/

# mutation testing
[group('dev')]
mutate *args:
    gremlins unleash {{ args }}

# tag and push a release
[group('dev')]
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
json-merkle file="testdata/sample.json":
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

# ── Hash ──────────────────────────────────────────────────────────────────────

# full derivation table
[group('hash')]
hash text="my-api-service":
    go run {{ main }} hash "{{ text }}"

# print derived port only
[group('hash')]
hash-port text="my-api-service":
    go run {{ main }} hash "{{ text }}" --port

# TUI with history
[group('hash')]
hash-ui:
    go run {{ main }} hash --ui

# browser UI with history
[group('hash')]
hash-web:
    go run {{ main }} hash --web

# ── Render ────────────────────────────────────────────────────────────────────

# d2 diagram to SVG
[group('diagram')]
diagram file="testdata/sample.d2":
    go run {{ main }} diagram {{ file }}

# mermaid diagram to SVG
[group('diagram')]
diagram-mermaid file="testdata/sample.mmd":
    go run {{ main }} diagram {{ file }}

# diagram to PNG
[group('diagram')]
diagram-png file="testdata/sample.d2":
    go run {{ main }} diagram {{ file }} --format png -o out.png

# diagram to draw.io
[group('diagram')]
diagram-drawio file="testdata/sample.d2":
    go run {{ main }} diagram {{ file }} --format drawio -o out.drawio

# terminal viewer
[group('diagram')]
diagram-ui file="testdata/sample.d2":
    go run {{ main }} diagram {{ file }} --ui

# browser live preview
[group('diagram')]
diagram-web file="testdata/sample.d2":
    go run {{ main }} diagram {{ file }} --web

# ── Image ─────────────────────────────────────────────────────────────────────

# analysis report and quality/size ladder
[group('image')]
image file="testdata/sample.jpg":
    go run {{ main }} image {{ file }}

# optimize to a file
[group('image')]
image-out file="testdata/sample.jpg" quality="70":
    go run {{ main }} image {{ file }} -o out.jpg -q {{ quality }}

# shrink to fit a size budget
[group('image')]
image-target file="testdata/sample.jpg" target="30kb":
    go run {{ main }} image {{ file }} -o out.jpg --target {{ target }}

# downscale to a maximum width
[group('image')]
image-thumb file="testdata/sample.jpg" width="400":
    go run {{ main }} image {{ file }} -o thumb.jpg --max-width {{ width }}

# interactive terminal viewer
[group('image')]
image-ui file="testdata/sample.jpg":
    go run {{ main }} image {{ file }} --ui

# browser UI
[group('image')]
image-web file="testdata/sample.jpg":
    go run {{ main }} image {{ file }} --web
