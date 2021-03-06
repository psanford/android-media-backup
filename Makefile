PLATFORM_JAR=$(ANDROID_HOME)/platforms/android-30/android.jar

android-media-backup.apk: android-media-backup/jgo.jar $(wildcard *.go) $(wildcard **/*.go)  $(wildcard **/*.java)
	go run gioui.org/cmd/gogio -target android ./android-media-backup

android-media-backup/jgo.jar: $(wildcard **/*.java)
	mkdir -p classes
	javac -cp $(PLATFORM_JAR) -sourcepath $(PLATFORM_JAR) -d classes $^
	jar cf $@ -C classes .
	rm -rf classes
