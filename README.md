# unum

> *unum* — omnia in uno.

A unified developer tool suite, built in Go.

[![CI](https://github.com/danielriddell21/unum/actions/workflows/ci.yaml/badge.svg)](https://github.com/danielriddell21/unum/actions/workflows/ci.yaml)
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=danielriddell21_unum&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=danielriddell21_unum)
[![codecov](https://codecov.io/gh/danielriddell21/unum/graph/badge.svg?token=ICL8H52H58)](https://codecov.io/gh/danielriddell21/unum)
[![Go 1.26](https://img.shields.io/badge/go-1.26-blue)](https://go.dev)
[![MIT License](https://img.shields.io/badge/licence-MIT-green)](LICENSE)

---

## Tools

| Command | Description |
|---|---|
| `unum json` | JSON viewer, validator, and analyzer — three output modes, six analysis lenses |
| `unum diff` | Diff visualizer — text, JSON, YAML, and Terraform plans; static, TUI, and web output |
| `unum hash` | Deterministic deriver — maps any string to a stable port, UUID, color, short ID, emoji, and passphrase |

Full documentation → [Wiki](https://github.com/danielriddell21/unum/wiki)

---

## Install

### Homebrew
```bash
brew install danielriddell21/tap/unum
```

### Go install
```bash
go install github.com/danielriddell21/unum/cmd/unum@latest
```

### From source
```bash
git clone https://github.com/danielriddell21/unum
cd unum
just install
```

---

## Acknowledgements

- The initial architecture and design of this project was conceived with the assistance of [Claude](https://claude.ai) (Anthropic), which was also used throughout development for identifying and fixing bugs.
- The TUI layout and interaction model was inspired by [lazygit](https://github.com/jesseduffield/lazygit) and [lazydocker](https://github.com/jesseduffield/lazydocker) by Jesse Duffield.
