# AGENTS.md - Ceylon News Mobile App

## Overview

Apache Cordova Android app for Sri Lankan news aggregation.

- **Package**: `xyz.navinda.ceylonnews`
- **Platforms**: Android only
- **Min SDK**: 28
- **Target SDK**: 35

## Prerequisites

- Android Studio
- Node.js v18+
- Apache Cordova CLI: `npm install -g cordova`

## Setup

```bash
cd mobile
npm install

# Add Android platform (if not present)
cordova platform add android

# Or refresh platform
make android-platform
```

## Build Commands

All builds via Makefile from repo root:

```bash
# Debug APK
make android-build

# Release APK (requires mobile/build.json)
make android-release

# Run on device/emulator
make android-run

# Refresh Android platform
make android-platform
```

## Project Structure

```
mobile/
├── config.xml          # App config, permissions, icons
├── package.json        # Cordova dependencies
├── build.json          # Release keystore config (gitignored)
├── build.json.example  # Template for build.json
├── res/                # Icons, splash screens, colors
├── platforms/          # Generated Android project (gitignored)
├── plugins/            # Cordova plugins
└── www/                # Web assets (HTML/JS/CSS)
    ├── index.html      # Main entry
    ├── home.html       # Home page
    ├── article.html    # Article view
    ├── search.html     # Search page
    ├── settings.html   # Settings page
    ├── sw.js           # Service worker
    ├── css/            # Stylesheets
    ├── js/             # JavaScript
    │   ├── core.js     # App initialization
    │   ├── api.js      # Backend API client
    │   ├── storage.js  # Local storage
    │   ├── indexeddb.js # IndexedDB wrapper
    │   └── pages/      # Page-specific JS
    ├── img/            # Images
    └── fonts/          # Font files
```

## Release Builds

Requires `mobile/build.json` with keystore credentials:

```bash
cp build.json.example build.json
# Edit build.json with your keystore details
```

**Never commit `build.json` or keystore files.**

Example `build.json`:
```json
{
  "android": {
    "release": {
      "keystore": "/path/to/keystore.jks",
      "alias": "your-key-alias",
      "storePassword": "your-store-password",
      "password": "your-key-password"
    }
  }
}
```

## Configuration

### config.xml

Key settings:
- `android-minSdkVersion`: 28
- `android-targetSdkVersion`: 35
- `android-usesCleartextTraffic`: true (allows HTTP)
- Orientation: portrait only

### CDN Dependencies

- **Phosphor Icons** - `https://unpkg.com/@phosphor-icons/web`
  - Used for all icons throughout the app
  - Supports multiple weights: regular, thin, light, bold, fill, duotone
  - Icon classes: `ph`, `ph-fill`, `ph-bold`, etc.

### Cordova Plugins

- `cordova-plugin-network-information` - Network state detection

## Development Workflow

```bash
# 1. Make changes to www/ files

# 2. Build debug APK
make android-build

# 3. Run on connected device/emulator
make android-run

# Or manually:
cd mobile && cordova run android
```

## Key Files

- `www/index.html` - App entry point
- `www/js/core.js` - App initialization, routing, global state
- `www/js/api.js` - Backend API communication
- `www/js/storage.js` - LocalStorage for saved articles
- `www/js/indexeddb.js` - IndexedDB for read history
- `www/sw.js` - Service worker for offline support

## UI Framework

- **Vanilla JavaScript** - No framework, pure DOM manipulation
- **Phosphor Icons** - Icon library via CDN (`https://unpkg.com/@phosphor-icons/web`)

## API Integration

The app connects to the Go backend API. Default endpoint configuration is in `www/js/api.js`.

## Assets

- Icons: `res/android/icon/` (ldpi through xxxhdpi)
- Splash: `res/android/splash/`
- Colors: `res/android/xml/colors.xml`

## Git Ignore

Do not commit:
- `platforms/`
- `plugins/` (if using npm install)
- `node_modules/`
- `build.json`
- `*.jks` (keystore files)
- `*.pepk` (private key)
