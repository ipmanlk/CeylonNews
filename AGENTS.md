# AGENTS.md - Ceylon News

## Project Structure

- `api/` - Go backend API (module: `ipmanlk/cnapi`)
- `mobile/` - Apache Cordova Android app

## Build Commands

All builds use the Makefile at repo root:

```bash
# Backend API
make api-build      # Production binary (requires fts5 build tag)
make api-dev        # Run development server
make api-test       # Run tests

# Mobile App
make android-build      # Debug APK
make android-release    # Release APK (requires mobile/build.json)
make android-platform   # Refresh Android platform
make android-run        # Run on device/emulator
```

## API (Go)

### Critical Build Tag

All Go builds MUST include `--tags "fts5"` for SQLite FTS5 support:

```bash
go build --tags "fts5" -o build/cnapi ./cmd/server
go test -v --tags "fts5" ./...
```

### Development

**Local dev (requires browser-api service):**
```bash
cd api
# Copy and configure env
cp .env.example .env
# Start browser-api first (Docker)
docker compose -f docker-compose.dev.yml up browser-api -d
# Run dev server
make api-dev
```

**Docker dev (full stack):**
```bash
cd api
docker compose -f docker-compose.dev.yml up
# API on :8080, browser-api on :8000
```

**Hot reload with Air:**
```bash
cd api
air  # uses .air.toml config
```

### Entry Point

- `api/cmd/server/main.go` - Server entry point
- `api/internal/app/app.go` - Application initialization

### Dependencies

- SQLite with FTS5 (via mattn/go-sqlite3)
- Goose for migrations
- Browser API service for JS-rendered pages

## Mobile (Cordova)

### Prerequisites

- Android Studio
- Node.js v18+
- Apache Cordova CLI

### Setup

```bash
cd mobile
npm install
cordova platform add android  # if needed
```

### Release Builds

Requires `mobile/build.json` (see `build.json.example`). Never commit keystore credentials.

### Config

- `mobile/config.xml` - App config, permissions, icons
- `mobile/www/` - Web assets (HTML/JS/CSS)

## Key Conventions

- API uses structured logging (slog) with JSON format in production
- Database migrations run automatically on startup
- Scheduler handles periodic news scraping (configurable via env)
- Browser API service required for scraping JS-heavy sites
