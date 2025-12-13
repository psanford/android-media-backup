# Legacy Gio-based build (to be removed after migration)
OTHER_JARS=./jars/work-runtime-2.5.0-sources.jar
GIO_AAR=android/libs/android-media-backup.aar

gio-apk: $(GIO_AAR)
	(cd android && ./gradlew assembleDebug)
	mv android/build/outputs/apk/debug/android-debug.apk media-backup.apk

$(GIO_AAR): $(shell find . -name '*.go' -o -name '*.java' -type f)
	mkdir -p $(@D)
	go run gioui.org/cmd/gogio -ldflags "-X 'github.com/psanford/android-media-backup/version.Version=$(shell date --rfc-3339=seconds)'" -buildmode archive -target android -appid io.sanford.media_backup -o $@ .

# New gomobile-based build
MOBILE_AAR=app/libs/mobile.aar
VERSION=$(shell date --rfc-3339=seconds)

.PHONY: gomobile-init
gomobile-init:
	go install golang.org/x/mobile/cmd/gomobile@latest
	gomobile init

$(MOBILE_AAR): $(shell find . -name '*.go' -type f) gomobile-init
	mkdir -p $(@D)
	gomobile bind -v -o $@ -target=android -androidapi 23 \
		-ldflags "-X 'github.com/psanford/android-media-backup/version.Version=$(VERSION)'" \
		./mobile

.PHONY: go-aar
go-aar: $(MOBILE_AAR)

.PHONY: apk
apk: $(MOBILE_AAR)
	(cd app && ../gradlew assembleDebug)
	cp app/build/outputs/apk/debug/app-debug.apk media-backup.apk

.PHONY: clean
clean:
	rm -rf $(GIO_AAR) $(MOBILE_AAR) media-backup.apk app/build android/build
