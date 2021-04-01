package wifi

/*
#include <jni.h>
*/
import "C"

import (
	"fmt"
	"log"
	"unsafe"

	"gioui.org/app"
	"git.wow.st/gmp/jni"
)

type ConnState int

func (cs ConnState) String() string {
	switch cs {
	case ConnStateUnknown:
		return "ConnStateUnknown"
	case NoNetwork:
		return "NoNetwork"
	case NotWifi:
		return "NotWifi"
	case Wifi:
		return "Wifi"
	default:
		return fmt.Sprintf("ConnStateUnkown<%d>", cs)
	}
}

var foo = "bar"

const (
	ConnStateUnknown ConnState = -2
	NoNetwork        ConnState = -1
	NotWifi          ConnState = 0
	Wifi             ConnState = 1
)

func ConnectionState() (ConnState, error) {
	jvm := jni.JVMFor(app.JavaVM())
	var state = -2
	err := jni.Do(jvm, func(env jni.Env) error {

		var uptr = app.AppContext()
		appCtx := *(*jni.Object)(unsafe.Pointer(&uptr))
		loader := jni.ClassLoaderFor(env, appCtx)
		cls, err := jni.LoadClass(env, loader, "io.sanford.media_backup.Jni")
		if err != nil {
			log.Printf("Load io.sanford.media_backup.Jni error: %s", err)
		}

		mid := jni.GetStaticMethodID(env, cls, "connectionState", "(Landroid/content/Context;)I")
		state, err = jni.CallStaticIntMethod(env, cls, mid, jni.Value(appCtx))

		return err
	})

	return ConnState(state), err
}
