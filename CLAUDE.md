# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Build
go build ./...

# Run tests
go test ./...

# Run a single test
go test ./internal/... -run TestRender

# Run with fake data for development (starts a local web server)
go run ./cmd/kidscreen.go -dev -fake

# Run with fake data and custom address
go run ./cmd/kidscreen.go -dev -fake -addr :8080

# Run normally (requires a config file)
go run ./cmd/kidscreen.go -config ~/.config/kidscreen/config.yaml -img screen.png
```

## Architecture

Kidscreen has two parts:

1. **Go CLI** (`cmd/`, `internal/`) — fetches data from external APIs, assembles it into an HTML page, screenshots it with headless Chrome, and saves a PNG.
2. **Arduino sketch** (`kidscreen/kidscreen.ino`) — wakes daily, downloads the PNG over Wi-Fi, and updates the e-ink display.

### Data flow (Go CLI)

`screen.go: Run()` constructs a list of `Card`s and a `Header`, then calls either `DevRender` (for development, hot-reloading HTTP server) or `Render` (production, headless Chrome screenshot saved to disk).

All data loading is deferred behind loader functions and executed in parallel in `assembleData()`. Cards whose `Valid()` method returns false are excluded from the final render.

### Card system

`Card` (`card.go`) is the central data type. Cards have three display types: `CardTypeText`, `CardTypeList`, and `CardTypeChart`. Each card has a `loader func(*Card) error` that populates its data on demand. Cards are sorted by `Priority` (higher = shown first).

The pattern for each data source is:
- A `New*Card()` constructor returning a `Card` with a deferred `loader`
- A `NewFake*Card()` counterpart using random/hardcoded data for `--dev --fake` mode
- A `newLazy()` wrapper (`lazy.go`) to ensure expensive fetches (network calls) happen exactly once even when multiple cards share the same data source (e.g., weather data is shared between the precipitation chart card and the header)

### Data sources

| Package | Source | Card(s) produced |
|---|---|---|
| `airquality.go` | Open-Meteo air quality API | AQI chart |
| `weather.go` | Open-Meteo forecast API | Precipitation chart + weather text |
| `calendar.go` | iCal HTTP feeds (e.g. Google Calendar) | Today/Tomorrow event lists |
| `generated.go` | OpenAI chat completions API | Configurable text cards |
| `picture.go` | HTML scraping via XPath | Picture card |

### HTML template

`screen.go.html` is the Go HTML template embedded into the binary. In production it is rendered once per run by the embedded FS. In `--dev` mode, `DevRender` reads the template from disk on every request (so edits are reflected without restarting).

### Configuration

Config is read from a YAML file (default: `~/.config/kidscreen/config.yaml`). `ReadConfig()` merges user values onto defaults from `defaultConfig()` and validates the result. See `configs/config.yaml` for a fully annotated example.

### Arduino sketch

`kidscreen.ino` requires a `secret.h` file (not committed) that defines `ssid` and `pass`. The sketch wakes once daily at `HOUR_OF_REFRESH` (default 3am), fetches the PNG from `IMAGE_URL`, renders it in 3-bit grayscale mode, and sleeps until the next refresh.
