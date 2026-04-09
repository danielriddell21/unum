# unum — justfile
# Install just: https://github.com/casey/just

binary := "bin/unum"
main := "./cmd/unum"

# Build the binary
build:
    go build -ldflags "-X main.version=dev" -o {{ binary }} {{ main }}

# Run unit tests — pass extra flags e.g. `just test -v` or `just test -run TestParse`
test *args:
    go test {{ args }} ./...

# Clean build artifacts
clean:
    rm -rf bin/

# Run the tool against a JSON file
run file="testdata/sample.json":
    go run {{ main }} json {{ file }}

# Launch the TUI navigator
ui file="testdata/sample.json":
    go run {{ main }} json {{ file }} --ui

# Lint with golangci-lint
lint:
    golangci-lint run ./...

# Install binary to GOPATH/bin
install:
    go install {{ main }}

# Tag and push a release
release tag:
    git tag {{ tag }}
    git push origin {{ tag }}

# Full CI pipeline
ci: lint test build
