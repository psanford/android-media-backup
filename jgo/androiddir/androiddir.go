package androiddir

/*
#include <jni.h>
*/
import "C"

import (
	"log"
	"unsafe"

	"gioui.org/app"
	"git.wow.st/gmp/jni"
)

func CacheDir() string {
	var cacheDirName string
	jvm := jni.JVMFor(app.JavaVM())
	err := jni.Do(jvm, func(env jni.Env) error {

		var uptr = app.AppContext()
		appCtx := *(*jni.Object)(unsafe.Pointer(&uptr))

		cls := jni.GetObjectClass(env, appCtx)
		mid := jni.GetMethodID(env, cls, "getCacheDir", "()Ljava/io/File;")

		file, err := jni.CallObjectMethod(env, appCtx, mid)
		if err != nil {
			return err
		}

		cls = jni.GetObjectClass(env, file)
		mid = jni.GetMethodID(env, cls, "getAbsolutePath", "()Ljava/lang/String;")

		jname, err := jni.CallObjectMethod(env, file, mid)
		if err != nil {
			return err
		}

		cacheDirName = jni.GoString(env, jni.String(jname))

		return nil
	})

	if err != nil {
		log.Printf("get cache dir err: %s", err)
	}

	return cacheDirName
}
