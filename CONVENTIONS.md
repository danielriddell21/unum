# Conventions

Structure shared across the tool family — unum, fiat-lux, galapagos, pandemonium,
vivarium, toolshed, narrata, gambit, rubix — so they read like they were written
by one person. Each repo records only the conventions it follows; this repo
follows the **CLI entrypoint** convention below. **unum is the reference** — when
in doubt, copy what it does.

These cover the *shape* of an entrypoint, not the behaviour inside it.

## CLI entrypoint

Every binary is built the same way:

- **Thin `cmd/<name>/main.go`.** It declares `var version = "dev"` (overridden at
  release via `-ldflags "-X main.version=…"`) and does nothing but
  `cli.Execute(version)`, printing the error and exiting non-zero on failure.
- **`internal/cli` owns the command tree.** `func Execute(version string) error`
  builds the root `*cobra.Command` (with `SilenceUsage`/`SilenceErrors`),
  registers subcommands, and runs it. Keep all flag wiring here, not in `main`.
- **`--version`** comes from cobra's `Version` field. Do **not** add a bespoke
  `version` subcommand — `--version` (and cobra's `-v`) is the one way.
- **`completion`** subcommand in `internal/cli/completion.go`, generating
  bash/zsh/fish/powershell from `cmd.Root()`. Copy unum's verbatim, changing only
  the binary name.

**Many-binary repos.** When a repo ships several small binaries, give each a cobra
root via a shared helper: `cli.Execute(root *cobra.Command)` sets the build
version, silences usage, adds the `completion` subcommand, and runs the command.
Each `cmd/<name>/main.go` builds its own root and calls `cli.Execute(root)`, so
every binary gets `--version` + completion for free.
