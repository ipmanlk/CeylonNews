# AGENTS.md - Ceylon News API

## Module

Module path: `ipmanlk/cnapi`
Go version: 1.24+

## Critical Build Tag

**ALL Go commands MUST include `--tags "fts5"`** for SQLite FTS5 support:

```bash
go build --tags "fts5" -o build/cnapi ./cmd/server
go test -v --tags "fts5" ./...
go run --tags "fts5" cmd/server/main.go
```

Without this tag, builds will fail with SQLite errors.

## Project Structure

```
api/
├── cmd/server/main.go      # Application entry point
├── internal/
│   ├── app/app.go          # App initialization & lifecycle
│   ├── api/                # HTTP handlers & routes
│   ├── config/             # Environment config loading
│   ├── database/           # Database migrations & stores
│   ├── fetcher/            # HTTP + browser API clients
│   ├── http/               # HTTP utilities
│   ├── model/              # Data models
│   ├── scheduler/          # Periodic scraping scheduler
│   ├── scraper/            # Source config & article scraping
│   └── service/            # Business logic layer
├── pkg/                    # Public packages
├── sources/                # News source TOML configs
└── browser-api/            # Python service for JS-rendered pages
```

## Development

### Local Dev (requires browser-api service)

```bash
cd api

# 1. Copy and configure environment
cp .env.example .env

# 2. Start browser-api first (Docker)
docker compose -f docker-compose.dev.yml up browser-api -d

# 3. Run dev server
make api-dev
```

### Docker Dev (full stack)

```bash
cd api
docker compose -f docker-compose.dev.yml up
# API on :8080, browser-api on :8000
```

### Hot Reload with Air

```bash
cd api
air  # uses .air.toml config
```

### Testing

```bash
cd api
go test -v --tags "fts5" ./...

# Run specific package
go test -v --tags "fts5" ./internal/scraper
```

## Database

- SQLite with FTS5 extension (via mattn/go-sqlite3)
- Migrations run automatically on startup using Goose
- Migration files in `internal/database/migrations/`
- Data directory: `api/data/` (gitignored)

## Source Configuration

News sources are defined in `sources/*.toml`. See `sources/SPEC.md` for full specification.

Key points:
- Each source has a unique `id` (kebab-case, never change)
- Sources support multiple languages (`en`, `si`, `ta`)
- Pipeline: Discovery → Extraction → Validation → Transformation → Storage
- Discovery types: `rss` or `html`
- Browser flag for JS-rendered pages

Example source structure:
```toml
id = "daily-mirror"
name = "Daily Mirror"

[[languages]]
language = "en"
max_items = 5

[languages.discovery]
type = "rss"
url = "https://www.dailymirror.lk/rss/..."

[languages.extraction]
browser = true
```

## Browser API Service

The `browser-api/` directory contains a Python service for scraping JavaScript-heavy sites.

- Required for sources with `browser = true`
- Runs as separate Docker service
- Endpoint: `http://browser-api:8000` (internal) or `http://localhost:8000` (local dev)

## Key Conventions

- **Logging**: Structured logging (slog) with JSON format in production
- **Config**: Environment-based via `.env` file
- **Scheduler**: Periodic news scraping (configurable via `SCHEDULER_ENABLED` and `SCHEDULER_SCRAPE_INTERVAL`)
- **Error handling**: Return errors up the stack, log at appropriate level
- **Context**: Pass `context.Context` for cancellation/timeouts

## Environment Variables

See `.env.example` for all options. Key ones:

```bash
# Database
DB_DRIVER=sqlite3
DB_DSN=./data/db.sqlite

# Fetcher
FETCHER_BROWSER_API_URL=http://localhost:8000
FETCHER_HTTP_TIMEOUT=15s

# Scheduler
SCHEDULER_ENABLED=true
SCHEDULER_SCRAPE_INTERVAL=1h

# HTTP Server
HTTP_HOST=0.0.0.0
HTTP_PORT=8080

# Logging
LOG_LEVEL=info
LOG_FORMAT=json
```
