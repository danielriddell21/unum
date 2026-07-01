# unum — Claude Code instructions

unum is a unified CLI dev-tool suite built in Go. Correctness and simplicity over cleverness — prefer a well-tested, focused function over a clever abstraction. If the standard library does it, use it.

## Before every commit

* Run `just ci` autonomously (lint + test + yeet + build). All must pass.
* If adding a new function or package: write unit tests alongside the code, in the same commit.
* If adding a new CLI command: add functional tests to `tests/`.
* Never commit until the user explicitly confirms. Propose changes as diffs, run `just ci` autonomously, then stop and wait before `git commit`.

## Testing standards

* Unit tests live next to the code they test (`*_test.go` in the same package).
* Functional tests live in `tests/` and use the compiled binary via `os/exec`.
* New parsers must have table-driven tests covering: identical input, each change kind, error cases.
* New CLI commands must have functional tests covering: happy path, `--no-color`, exit codes.

## Code quality

* Run `just lint` before proposing a diff. Fix all lint errors before committing.
* Prefer early returns over nesting.
* Do not add error handling or fallbacks for scenarios that cannot happen.
* Comments in `internal/` and `cmd/`: only *inside* functions, and only for non-obvious logic — never a doc comment on a declaration, a file or package header, or a `doc.go`. (`//go:build`, `//go:generate`, and `//nolint` are directives, not comments, and stay.)
* Public library code (module root or `pkg/…`, where present) keeps full godoc — a doc comment on every exported symbol and the package, enforced by the revive `exported` rule.

## Conventions

* The CLI entrypoint structure is shared across the tool family and documented in [CONVENTIONS.md](CONVENTIONS.md). unum is the reference — keep new code consistent with it.
