# Legacy Gio-based build (to be removed after migration)
OTHER_JARS=./jars/work-runtime-2.5.0-sources.jar
GIO_AAR=android/libs/android-media-backup.aar

.PHONY: gio-apk
gio-apk: $(GIO_AAR)
	(cd android && ./gradlew assembleDebug)
	mv android/build/outputs/apk/debug/android-debug.apk media-backup-gio.apk

$(GIO_AAR): $(shell find . -name '*.go' -o -name '*.java' -type f)
	mkdir -p $(@D)
	go run gioui.org/cmd/gogio -ldflags "-X 'github.com/psanford/android-media-backup/version.Version=$(shell date --rfc-3339=seconds)'" -buildmode archive -target android -appid io.sanford.media_backup -o $@ .

# New gomobile-based build
MOBILE_AAR=app/libs/mobile.aar
VERSION=$(shell date --rfc-3339=seconds)
TOOLSBIN=$(shell pwd)/.tools/bin
GOMOBILE=$(TOOLSBIN)/gomobile

# Initialize gomobile (one-time setup)
.PHONY: init
init:
	mkdir -p $(TOOLSBIN)
	GOBIN=$(TOOLSBIN) go install golang.org/x/mobile/cmd/gomobile
	GOBIN=$(TOOLSBIN) go install golang.org/x/mobile/cmd/gobind
	PATH=$(TOOLSBIN):$(PATH) $(GOMOBILE) init

# Build the Go mobile AAR library
$(MOBILE_AAR): $(shell find . -path ./app -prune -o -path ./android -prune -o -name '*.go' -print)
	mkdir -p $(@D)
	PATH=$(TOOLSBIN):$(PATH) $(GOMOBILE) bind -v -o $@ -target=android -androidapi 23 \
		-ldflags "-X 'github.com/psanford/android-media-backup/version.Version=$(VERSION)'" \
		./mobile

.PHONY: go-aar
go-aar: $(MOBILE_AAR)

# Build the Android APK (new Kotlin/Compose version)
.PHONY: apk
apk: $(MOBILE_AAR)
	./gradlew :app:assembleDebug
	cp app/build/outputs/apk/debug/app-debug.apk media-backup.apk
	@echo "Built media-backup.apk"

# Build release APK
.PHONY: apk-release
apk-release: $(MOBILE_AAR)
	./gradlew :app:assembleRelease
	cp app/build/outputs/apk/release/app-release-unsigned.apk media-backup-release.apk
	@echo "Built media-backup-release.apk (unsigned)"

.PHONY: clean
clean:
	rm -rf $(GIO_AAR) $(MOBILE_AAR) media-backup.apk media-backup-gio.apk media-backup-release.apk
	rm -rf app/build android/build .gradle

.PHONY: help
help:
	@echo "Available targets:"
	@echo "  init         - Initialize gomobile (one-time setup)"
	@echo "  apk          - Build debug APK (new Kotlin/Compose version)"
	@echo "  apk-release  - Build release APK (unsigned)"
	@echo "  go-aar       - Build Go mobile AAR only"
	@echo "  gio-apk      - Build legacy Gio-based APK"
	@echo "  clean        - Remove build artifacts"
