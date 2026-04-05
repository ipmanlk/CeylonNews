<a href="https://aimeos.org/">
    <img src="https://raw.githubusercontent.com/ipmanlk/CeylonNews/master/res/icon/android/xxxhdpi.png" alt="Aimeos logo" title="Aimeos" align="right" height="60" />
</a>

Ceylon News
======================

This app brings you the latest local news in Sri Lanka. It's built using my [CeylonNewsBackend](https://github.com/ipmanlk/CeylonNewsBackend) project and couple of other open source tools.

## Project Structure

```
ceylonnews/
├── mobile/          # Cordova mobile app (Android)
└── api/             # Go backend API
```

## How to build

### Prerequisites
- Android Studio
- Node.js v18 or above.
- Apache Cordova ([Docs](https://cordova.apache.org/docs/en/latest/)).
- Go 1.24+ (for backend)

### Building

#### Mobile App
```bash
cordova build android
```

Or with release signing:
```bash
make mobile-release
```

#### Backend API
```bash
make api-build
```

## Technologies
- Apache Cordova
- Onsen UI
- jQuery
- Go

## License

CeylonNews is licensed under the terms of the MIT
license and is available for free.

## Links

* [Google PlayStore](https://play.google.com/store/apps/details?id=xyz.navinda.ceylonnews&hl=en)
* [Issue tracker](https://github.com/ipmanlk/CeylonNews/issues)
* [Source code](https://github.com/ipmanlk/CeylonNews)
