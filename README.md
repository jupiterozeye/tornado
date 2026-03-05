<div align="center">

# Tornado

[![Release](https://img.shields.io/github/v/release/jupiterozeye/tornado?style=flat-square)](https://github.com/jupiterozeye/tornado/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/jupiterozeye/tornado)](https://goreportcard.com/report/github.com/jupiterozeye/tornado)
[![Go Version](https://img.shields.io/github/go-mod/go-version/jupiterozeye/tornado?style=flat-square)](go.mod)
[![License](https://img.shields.io/github/license/jupiterozeye/tornado?style=flat-square)](LICENSE)

</div>

<!-- TODO: Add GIF demo here -->
<!-- ![Demo](docs/demo.gif) -->

## Installation

**Go:**
```bash
go install github.com/jupiterozeye/tornado/cmd/tornado@latest
```

**Nix (flake):**
```bash
nix run github:jupiterozeye/tornado
```

## Usage

```bash
tornado
```

## Quick Usage

- Start Tornado and connect to your database
- Use the Explorer pane to browse tables
- Run SQL from the Query pane
- Inspect, filter (`/`), and copy from the Results pane

## Common Keys

- `space` open command menu
- `e` focus Explorer
- `q` focus Query
- `r` focus Results
- `ctrl+enter` execute query
- `t` open theme picker (from command menu)
- `?` help hints in status/footer
- `q` quit (from command menu)

## Development

```bash
make deps
make run
make test
```

## Tech Stack

- [Bubble Tea](https://github.com/charmbracelet/bubbletea)
- [Bubbles](https://github.com/charmbracelet/bubbles)
- [Lip Gloss](https://github.com/charmbracelet/lipgloss)
- [modernc.org/sqlite](https://pkg.go.dev/modernc.org/sqlite)

## License

MIT
