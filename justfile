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
