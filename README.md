# unum

> *unum* — omnia in uno.

A unified developer tool suite, built in Go.

---

## Tools

| Command | Description |
|---|---|
| `unum json` | JSON viewer, validator, and analyzer — three output modes, six analysis lenses |
| `unum diff` | Diff visualizer — text, JSON, YAML, and Terraform plans; static, TUI, and web output |

Full documentation → [Wiki](https://github.com/danielriddell21/unum/wiki)

---

## Install

### Homebrew
```bash
brew install danielriddell21/unum/unum
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
