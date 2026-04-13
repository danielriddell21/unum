# unum — Claude Code instructions

unum is a unified CLI dev-tool suite built in Go. Correctness and simplicity over cleverness — prefer a well-tested, focused function over a clever abstraction. If the standard library does it, use it.

## Before every commit

* Run `just ci` autonomously (lint + test + yeet + build). All must pass.
* If adding a new function or package: write unit tests alongside the code, in the same commit.
* If adding a new CLI command: add functional tests to `tests/`.
* Never commit until the user explicitly confirms. Propose changes as diffs, run `just ci` autonomously, then stop and wait before `git commit`.

## Architecture invariants

These are hard constraints — violating them will be caught in review:

* Lenses are isolated: a lens may only import `node/` and its own external dependencies.
* No lens may import another lens.
* `analyze/` is the only package that knows about all lenses.
* Annotate nodes via `node.AnnotationKey{Lens: YourLensName, Name: "key"}` — never raw strings.

Every new lens must implement the `analyze.Analyzer` interface:

```go
type Analyzer interface {
    Name() string
    Run(ctx context.Context, root *node.Node, opts analyze.Options) error
}
```

Register it in `internal/json/cmd.go` under `allAnalyzers` and add a corresponding flag.

## Testing standards

* Unit tests live next to the code they test (`*_test.go` in the same package).
* Functional tests live in `tests/` and use the compiled binary via `os/exec`.
* New parsers must have table-driven tests covering: identical input, each change kind, error cases.
* New CLI commands must have functional tests covering: happy path, `--no-color`, exit codes.

## Code quality

* Run `just lint` before proposing a diff. Fix all lint errors before committing.
* Prefer early returns over nesting.
* Do not add error handling or fallbacks for scenarios that cannot happen.
* Do not add comments unless the logic is non-obvious.
