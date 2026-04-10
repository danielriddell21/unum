# Contributing to unum

## Requirements

- [Go](https://go.dev) (stable — version from `go.mod`)
- [just](https://github.com/casey/just)
- [golangci-lint](https://golangci-lint.run/welcome/install/) (for `just lint`)

## Development workflow

```bash
just build      # build to bin/unum
just test       # run unit tests
just test -v    # verbose unit test output
just yeet       # run CLI functional tests (tests/)
just lint       # golangci-lint
just ci         # full gate: lint + test + yeet + build
```

Run `just ci` before every commit — all four stages must pass.

Manual smoke-testing shortcuts:

```bash
# json tool
just json                           # static render  (testdata/sample.json)
just json-ui                        # TUI navigator
just json-web                       # browser UI
just stats                          # --stats annotation lens
just hash                           # --merkle --hash-only
just typegen-go                     # Go struct generation
just typegen-ts                     # TypeScript interface generation
just schema                         # JSON Schema generation
just yaml                           # transform to YAML

# diff tool
just diff                           # text diff  (diff-a.txt / diff-b.txt)
just diff-ui                        # TUI diff  (diff-a.json / diff-b.json)
just diff-web                       # browser diff  (diff-a.json / diff-b.json)
just diff-yaml                      # semantic YAML diff
just diff-tf                        # Terraform plan diff
```

CI runs `lint`, `test`, and `build` on every push to `trunk` and on every pull request targeting `trunk`.

## Project layout

```
cmd/unum/           entry point
internal/
  cli/              root cobra command + flag wiring
  config/           ~/.config/unum/config.json loader
  json/             unum json tool
    analyze/        analyzer suite (runs enabled lenses)
    lens/<name>/    individual annotation lenses
    parse/          JSON parser + validator
    render/
      static/       terminal syntax-highlighted output
      tui/          Bubble Tea TUI navigator
      web/          embedded HTTP server + browser UI
  diff/             unum diff tool
    format/         file-extension format detection
    node/           shared Diff / DiffNode / Hunk types
    parse/          text, JSON, YAML, Terraform parsers
    render/
      static/       coloured +/- terminal output
      tui/          Bubble Tea TUI (unified + split views)
      web/          embedded HTTP server + browser diff UI
testdata/           fixture files used by justfile recipes and tests
tests/              CLI functional tests (go test ./tests/...)
```

## Adding a new tool

1. Create `internal/<toolname>/` mirroring the structure of `internal/json/` or `internal/diff/`
2. Register the cobra subcommand in `internal/cli/root.go`
3. Add `just` recipes for the common modes (`run`, `ui`, `web`)

## Adding a new JSON lens

Lenses live in `internal/json/lens/<name>/`. Each lens must implement `analyze.Analyzer`:

```go
type Analyzer interface {
    Name() string
    Run(ctx context.Context, root *node.Node, opts analyze.Options) error
}
```

Rules:
- A lens may only import `node/` and its own external dependencies
- No lens may import another lens
- `analyze/` is the only package that knows about all lenses
- Annotate nodes with `node.AnnotationKey{Lens: YourLensName, Name: "key"}`

Register the new lens in `internal/json/cmd.go` under `allAnalyzers` and add a flag for it.

## Adding a new diff format

1. Add a `parse/<format>.go` file that produces a `*node.Diff` (with `diff.Root` set for semantic diffs, `diff.Hunks` for text-based diffs)
2. Add the format constant to `internal/diff/node/node.go` and detection logic to `internal/diff/format/detect.go`
3. Wire the new case into the `switch fmt_` block in `internal/diff/cmd.go`
4. All three renderers (static, TUI, web) handle `diff.Root != nil` generically — no renderer changes needed for new semantic formats

## Commit style

```
type(scope): short imperative description

feat(json): add --flatten lens
fix(diff): correct hunk count for trailing newline
docs: update contributing guide
```

Types: `feat`, `fix`, `docs`, `refactor`, `test`, `chore`

No period at the end of the subject line. Keep the subject under 72 characters.

## Release

Releases are triggered by pushing a semver tag. Maintainers only:

```bash
just release v0.3.0
```

This creates and pushes the tag, which triggers the GoReleaser GitHub Actions workflow. Requires `HOMEBREW_TAP_TOKEN` to be set as a repository secret.
