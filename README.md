<a href="https://aimeos.org/">
    <img src="https://raw.githubusercontent.com/ipmanlk/CeylonNews/master/res/icon/android/xxxhdpi.png" alt="Aimeos logo" title="Aimeos" align="right" height="60" />
</a>

Ceylon News
======================

Your daily aggregator for Sri Lankan news. Built with Apache Cordova and Go.

## Project Structure

```
ceylonnews/
├── mobile/          # Cordova mobile app (Android)
└── api/             # Go backend API
```

## How to build

### Prerequisites
- Android Studio
- Node.js v18+
- Apache Cordova
- Go 1.24+

### Building

All builds use the Makefile:

```bash
# Mobile App
make android-build      # Debug APK
make android-release   # Release APK (requires mobile/build.json)
make android-platform  # Refresh Android platform

# Backend API
make api-build         # Production binary
make api-dev           # Run development server
make api-test          # Run tests
```

### Release Build Setup

For release builds, create `mobile/build.json`:

```json
{
  "android": {
    "release": {
      "keystore": "/path/to/your/keystore.jks",
      "alias": "your-key-alias",
      "storePassword": "your-store-password",
      "password": "your-key-password"
    }
  }
}
```

## Technologies
- Apache Cordova
- Onsen UI
- jQuery
- Go
- SQLite
- Goose

## License

CeylonNews is licensed under the terms of the MIT
license and is available for free.

## Links

* [Google PlayStore](https://play.google.com/store/apps/details?id=xyz.navinda.ceylonnews&hl=en)
* [Issue tracker](https://github.com/ipmanlk/CeylonNews/issues)
* [Source code](https://github.com/ipmanlk/CeylonNews)
