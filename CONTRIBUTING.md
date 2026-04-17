# Contributing to unum

## Requirements

* [Go](https://go.dev) (stable — version from `go.mod`)
* [just](https://github.com/casey/just)
* [golangci-lint](https://golangci-lint.run/welcome/install/) (for `just lint`)
* [gremlins](https://github.com/go-gremlins/gremlins) (for `just mutate`)

## Development workflow

```
just build      # build to bin/unum
just test       # run unit tests
just test -v    # verbose unit test output
just yeet       # run CLI functional tests (tests/)
just lint       # golangci-lint
just ci         # full gate: lint + test + yeet + build
just mutate     # mutation testing via gremlins
```

Run `just ci` before every commit — all four stages must pass. CI runs the same gate on every push to `trunk` and on every pull request targeting `trunk`.

Mutation testing (`just mutate`) runs automatically in CI on pushes to `trunk` via `go-gremlins/gremlins-action`. Thresholds and enabled mutation types are configured in `.gremlins.yaml`. The gate requires ≥60% efficacy (killed/tested mutants) with a coverage floor of 60% — only mutants in covered code are tested.

Run `just --list` to see all available recipes, including per-tool smoke-testing shortcuts.

## Project layout

```
cmd/unum/           entry point
internal/
  cli/              root cobra command + flag wiring
  config/           ~/.config/unum/config.json loader
  json/             unum json tool
    analyze/        analyzer suite (runs enabled lenses)
    lens/<n>/       individual annotation lenses
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

## Commit style

```
type(scope): short imperative description

feat(json): add --flatten lens
fix(diff): correct hunk count for trailing newline
docs: update contributing guide
```

Types: `feat`, `fix`, `docs`, `refactor`, `test`, `chore`

No period at the end of the subject line. Keep the subject under 72 characters.

## Releases

Releases are triggered by pushing a semver tag — maintainers only. Requires `HOMEBREW_TAP_TOKEN` set as a repository secret. See `just release` for the one-command workflow.
```
