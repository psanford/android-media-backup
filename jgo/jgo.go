package jgo

/*
#include <jni.h>
*/
import "C"

import (
	"fmt"
	"log"
	"sync"
	"unsafe"

	"gioui.org/app"
	_ "gioui.org/app/permission/networkstate"
	_ "gioui.org/app/permission/storage"
	"git.wow.st/gmp/jni"
)

type PermResult struct {
	Authorized bool
	Err        error
}

var (
	pendingResultMux sync.Mutex
	pendingResults   []chan PermResult
)

func RequestPermission(viewEvt app.ViewEvent) <-chan PermResult {
	pendingResultMux.Lock()
	pendingResult := make(chan PermResult, 1)
	pendingResults = append(pendingResults, pendingResult)
	pendingResultMux.Unlock()

	go func() {
		jvm := jni.JVMFor(app.JavaVM())
		log.Printf("AppCtx: %d", app.AppContext())
		err := jni.Do(jvm, func(env jni.Env) error {

			var uptr = app.AppContext()
			appCtx := *(*jni.Object)(unsafe.Pointer(&uptr))
			loader := jni.ClassLoaderFor(env, appCtx)
			cls, err := jni.LoadClass(env, loader, "io.sanford.jgo.Jni")
			if err != nil {
				log.Printf("Load io.sanford.jgo.Jni error: %s", err)
			}

			mid := jni.GetMethodID(env, cls, "<init>", "()V")

			inst, err := jni.NewObject(env, cls, mid)
			if err != nil {
				log.Printf("NewObject err: %s", err)
			}

			mid = jni.GetMethodID(env, cls, "register", "(Landroid/view/View;)V")

			jni.CallVoidMethod(env, inst, mid, jni.Value(viewEvt.View))
			return err
		})

		if err != nil {
			log.Printf("Err: %s", err)
		}
	}()

	return pendingResult
}

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

const (
	ConnStateUnknown ConnState = -2
	NoNetwork        ConnState = -1
	NotWifi          ConnState = 0
	Wifi             ConnState = 1
)

func ConnectionState(viewEvt app.ViewEvent) (ConnState, error) {
	jvm := jni.JVMFor(app.JavaVM())
	log.Printf("AppCtx: %d", app.AppContext())
	var state = -2
	err := jni.Do(jvm, func(env jni.Env) error {

		var uptr = app.AppContext()
		appCtx := *(*jni.Object)(unsafe.Pointer(&uptr))
		loader := jni.ClassLoaderFor(env, appCtx)
		cls, err := jni.LoadClass(env, loader, "io.sanford.jgo.Jni")
		if err != nil {
			log.Printf("Load io.sanford.jgo.Jni error: %s", err)
		}

		mid := jni.GetStaticMethodID(env, cls, "connectionState", "(Landroid/content/Context;)I")
		state, err = jni.CallStaticIntMethod(env, cls, mid, jni.Value(appCtx))

		return err
	})

	return ConnState(state), err
}

//export Java_io_sanford_jgo_Jni_permissionResult
func Java_io_sanford_jgo_Jni_permissionResult(env *C.JNIEnv, cls C.jclass, jok C.jboolean) {
	log.Printf("permissionResult: %d", jok)

	var authorized bool
	if jok > 0 {
		authorized = true
	}

	result := PermResult{
		Authorized: authorized,
	}

	pendingResultMux.Lock()
	for _, pending := range pendingResults {
		pending <- result
	}
	pendingResults = pendingResults[:0]
	pendingResultMux.Unlock()
}
