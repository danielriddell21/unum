# Contributing to unum

## Requirements

- [Go](https://go.dev) (stable)
- [just](https://github.com/casey/just)
- [golangci-lint](https://golangci-lint.run/welcome/install/) (for `just lint`)

## Development workflow

```bash
just build      # build to bin/unum
just test       # run tests
just test -v    # verbose test output
just lint       # golangci-lint
just run        # static mode with testdata/sample.json
just ui         # TUI mode with testdata/sample.json
just web        # web mode with testdata/sample.json
just ci         # full gate: lint + test + build
```

## Adding a new tool

1. Create `internal/<toolname>/` with the same structure as `internal/json/`
2. Register the cobra subcommand in `internal/cli/root.go`
3. Add `just` recipes for the common modes

## Adding a new lens

Lenses live in `internal/json/lens/<name>/`. Each lens must implement `analyze.Analyzer`:

```go
type Analyzer interface {
    Name() string
    Run(ctx context.Context, root *node.Node, opts analyze.Options) error
}
```

Rules:
- A lens may only import `node/` and its own external dependency
- No lens may import another lens
- `analyze/` is the only package that knows about all lenses
- Annotate nodes with `node.AnnotationKey{Lens: YourLensName, Name: "key"}`

Register the new lens in `internal/json/cmd.go` under `allAnalyzers` and wire up the flag.

## Release

Releases are triggered by pushing a semver tag. Maintainers only:

```bash
just release v0.2.0
```

This creates and pushes the tag, which triggers the GoReleaser GitHub Actions workflow. Requires `HOMEBREW_TAP_TOKEN` to be set as a repository secret.
