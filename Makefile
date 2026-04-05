.PHONY: api-build api-dev api-test help android-build android-platform android-release android-run

help:
	@echo "Ceylon News - Available Commands"
	@echo "  make api-build           - Build production binary"
	@echo "  make api-dev             - Run development server"
	@echo "  make api-test            - Run all tests"
	@echo "  make android-build       - Build Android debug APK"
	@echo "  make android-platform    - Refresh Android platform (remove/add)"
	@echo "  make android-release     - Build Android release APK (requires mobile/build.json)"
	@echo "  make android-run         - Run Android app on device/emulator"

api-build:
	cd api && go build --tags "fts5" -o build/cnapi ./cmd/server

api-dev:
	cd api && go run --tags "fts5" cmd/server/main.go

api-test:
	cd api && go test -v --tags "fts5" ./...

android-build:
	cd mobile && cordova build android --debug

android-platform:
	cd mobile && cordova platform remove android && cordova platform add android

android-release:
	@if [ ! -f mobile/build.json ]; then \
		echo "Error: Missing mobile/build.json. Copy mobile/build.json.example to mobile/build.json and fill in values"; \
		exit 1; \
	fi
	cd mobile && cordova build android --release

android-run:
	cd mobile && cordova run android
