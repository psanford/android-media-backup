PLATFORM_JAR=$(ANDROID_HOME)/platforms/android-30/android.jar
OTHER_JARS=./jars/work-runtime-2.5.0-sources.jar

android-media-backup.apk: android/libs/android-media-backup.aar
	(cd android && ./gradlew assembleDebug)
	mv android/build/outputs/apk/debug/android-debug.apk $@

android/libs/android-media-backup.aar: jgo.jar $(wildcard *.go) $(wildcard **/*.go)  $(wildcard **/*.java)
	mkdir -p $(@D)
	go run gioui.org/cmd/gogio -buildmode archive -target android -appid io.sanford.android_media_backup -o $@ .

jgo.jar: jgo/Jni.java
	mkdir -p classes
	javac -cp "$(PLATFORM_JAR):"  -d classes $^
	jar cf $@ -C classes .
	rm -rf classes
