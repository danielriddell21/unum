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

Run `just --list` to see all available recipes, including per-tool smoke-testing shortcuts.

## Project layout

```
cmd/unum/           entry point
internal/
  cli/              root cobra command + flag wiring
  config/           ~/.config/unum/config.json loader
  telemetry/        OpenTelemetry + Umami analytics
  theme/            shared colour/style definitions
  tui/              shared Bubble Tea panel primitives
  web/              shared embedded HTTP server utilities
  <tool>/           one directory per tool (json, diff, hash, …)
    render/
      static/       terminal output
      tui/          Bubble Tea TUI
      web/          embedded HTTP server + browser UI
    …               tool-specific packages (parse/, derive/, lens/, etc.)
testdata/           fixture files used by justfile recipes and tests
tests/              CLI functional tests (go test ./tests/...)
```

## Conventions

unum is the reference for the CLI entrypoint layout shared across the tool
family. See [CONVENTIONS.md](CONVENTIONS.md) before adding an entrypoint.

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
