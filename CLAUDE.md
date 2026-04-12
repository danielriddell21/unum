# unum — Claude Code instructions

## Before every commit
- Run `just ci` (lint + test + yeet + build). All must pass.
- If adding a new function or package: write unit tests alongside the code, in the same commit.
- If adding a new CLI command: add functional tests to `tests/static/`.
- Never commit until the user explicitly confirms.

## Testing standards
- Unit tests live next to the code they test (`*_test.go` in the same package).
- Functional tests live in `tests/` split into three subdirectories:
  - `tests/static/` — CLI binary tests (exit codes, flag output, no-color). Uses compiled binary via `os/exec`.
  - `tests/tui/` — Bubble Tea TUI tests via `teatest` (headless, no PTY needed). Imports TUI packages directly.
  - `tests/web/` — HTTP API and browser tests via `rod` (headless Chromium). Fails if Chromium is not installed.
- New parsers must have table-driven tests covering: identical input, each change kind, error cases.
- New CLI commands must have functional tests in `tests/static/` covering: happy path, `--no-color`, exit codes.

## Code quality
- Run `just lint` before proposing a diff. Fix all lint errors before committing.
- Do not add error handling or fallbacks for scenarios that cannot happen.
- Do not add comments unless the logic is non-obvious.
