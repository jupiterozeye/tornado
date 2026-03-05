# Tornado

Tornado is a terminal-first SQL client focused on speed, clarity, and keyboard-driven workflows.

It gives you a clean TUI for exploring schemas, running queries, filtering results, and switching themes without leaving your terminal.

## Features

- SQLite and PostgreSQL support
- Explorer pane for tables and schema navigation
- SQL query editor with modal editing
- Results table with filter mode, preview, and copy actions
- Theme picker with live preview
- Persistent config for saved preferences (including theme)

## Install

### Go install (recommended)

```bash
go install github.com/jupiterozeye/tornado@latest
```

Then run:

```bash
tornado
```

If `tornado` is not found, add your Go bin directory to `PATH`:

```bash
export PATH="$(go env GOPATH)/bin:$PATH"
```

For Go 1.17+, `GOBIN` may be used instead of `GOPATH/bin` if set.

### Build from source

```bash
git clone https://github.com/jupiterozeye/tornado
cd tornado
make build
./bin/tornado
```

## Nix

This repo includes a `flake.nix`.

Run directly:

```bash
nix run .
```

Enter development shell:

```bash
nix develop
```

Build package:

```bash
nix build .#tornado
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
