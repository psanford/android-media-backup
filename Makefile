OTHER_JARS=./jars/work-runtime-2.5.0-sources.jar
AAR=android/libs/android-media-backup.aar

media-backup.apk: $(AAR)
	(cd android && ./gradlew assembleDebug)
	mv android/build/outputs/apk/debug/android-debug.apk $@

$(AAR): $(shell find . -name '*.go' -o -name '*.java' -type f)
	mkdir -p $(@D)
	go run gioui.org/cmd/gogio -buildmode archive -target android -appid io.sanford.android_media_backup -o $@ .
