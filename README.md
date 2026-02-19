# Tornado

A visually appealing TUI for SQL database viewing and management.

## Features (Planned)

- Connect to SQLite and PostgreSQL databases
- Browse tables and schemas
- Run custom SQL queries
- View database traffic with real-time charts

## Built With

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Bubbles](https://github.com/charmbracelet/bubbles) - UI components
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) - Styling
- [ntcharts](https://github.com/NimbleMarkets/ntcharts) - Terminal charts

## Quick Start

```bash
# Install dependencies
make deps

# Run the application
make run
```

## Project Structure

```
tornado/
├── cmd/tornado/          # Application entry point
├── internal/
│   ├── app/              # Root model and screen navigation
│   ├── db/               # Database interface and implementations
│   ├── models/           # Data structures
│   ├── ui/               # UI components and screens
│   └── telemetry/        # Metrics collection
├── Makefile
└── README.md
```

## Key Bindings

| Key | Action |
|-----|--------|
| `q` / `Ctrl+C` | Quit |
| `Tab` | Switch focus |
| `Enter` | Confirm/Execute |

## License

MIT
