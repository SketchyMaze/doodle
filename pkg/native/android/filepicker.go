//go:build android
// +build android

// Package android holds the Android-specific implementation code behind
// pkg/native's cross-platform functions -- kept out of pkg/native itself
// (which stays platform-generic aside from its small _android.go files) so
// the JNI/cgo plumbing has somewhere to live without cluttering it.
package android

/*
#include <jni.h>
#include <stdlib.h>

// JNIEnv methods are entries in a function-pointer table (env->Foo(env,
// ...) in C), which Go/cgo expressions can't call directly -- so each JNI
// call this package needs gets one small static C helper here instead,
// called from Go as a plain function.
static jstring new_string_utf(JNIEnv *env, const char *str) {
    return (*env)->NewStringUTF(env, str);
}

static void call_show_file_picker(JNIEnv *env, jobject activity, jboolean forSave, jstring suggestedName) {
    jclass cls = (*env)->GetObjectClass(env, activity);
    if (cls == NULL) {
        return;
    }

    jmethodID mid = (*env)->GetMethodID(env, cls, "showFilePicker", "(ZLjava/lang/String;)V");
    if (mid != NULL) {
        (*env)->CallVoidMethod(env, activity, mid, forSave, suggestedName);
    }

    (*env)->DeleteLocalRef(env, cls);
}
*/
import "C"

import (
	"errors"
	"sync"
	"unsafe"

	"git.kirsle.net/SketchyMaze/doodle/pkg/log"
	sdl2 "github.com/veandco/go-sdl2/sdl"
)

// Only one file picker can reasonably be on screen at a time (it's a
// full-screen system Activity), so a single pending-request slot -- guarded
// by a mutex, not a request-ID registry -- is enough to correctly pair a
// DeliverResult() call with whichever PickFile() call is currently blocked
// waiting for one, and safely ignore a stray/duplicate delivery otherwise.
var (
	mu      sync.Mutex
	pending chan pickResult
)

type pickResult struct {
	path string
	ok   bool
}

// PickFile launches Android's Storage Access Framework document picker
// (ACTION_OPEN_DOCUMENT for reads, ACTION_CREATE_DOCUMENT for the
// forSave=true case handled by MainActivity.java's showFilePicker()) and
// blocks until the user picks something or cancels.
//
// The filter argument (a "*.ext *.ext2"-style string, matching how
// pkg/native's OpenFile/SaveFile callers already call it -- see
// pkg/native/file_dialog_native.go's doc comment) isn't used here: SAF
// filters by MIME type, not extension, and this game's own extensions
// (.level, .doodad) have no registered MIME type to filter by, so the
// picker is opened showing every file type instead of trying to translate
// an extension list into one.
func PickFile(forSave bool, suggestedName string) (string, error) {
	mu.Lock()
	if pending != nil {
		mu.Unlock()
		return "", errors.New("a file picker is already open")
	}
	ch := make(chan pickResult, 1)
	pending = ch
	mu.Unlock()

	defer func() {
		mu.Lock()
		pending = nil
		mu.Unlock()
	}()

	if err := launchPicker(forSave, suggestedName); err != nil {
		return "", err
	}

	result := <-ch
	if !result.ok {
		return "", errors.New("canceled")
	}
	return result.path, nil
}

// DeliverResult is called by cmd/doodle-android's JNI-exported bridge
// function (Java_..._MainActivity_nativeFilePickerResult) once
// MainActivity.onActivityResult() has a result -- a local file path already
// copied out of the picked content:// URI, since Go's plain os.Open() can't
// read those directly (see MainActivity.java's copyUriToCache()) -- or ok=
// false if the user canceled the picker.
func DeliverResult(path string, ok bool) {
	mu.Lock()
	ch := pending
	mu.Unlock()

	if ch == nil {
		log.Warn("android.DeliverResult: no file picker was waiting for a result (path=%q ok=%v)", path, ok)
		return
	}
	ch <- pickResult{path: path, ok: ok}
}

func launchPicker(forSave bool, suggestedName string) error {
	// jobject (unlike JNIEnv*) doesn't support a direct `== nil` comparison
	// under cgo, so check the plain unsafe.Pointer before casting either.
	envPtr := sdl2.AndroidGetJNIEnv()
	activityPtr := sdl2.AndroidGetActivity()
	if envPtr == nil || activityPtr == nil {
		return errors.New("android.launchPicker: no JNI environment/activity available")
	}
	env := (*C.JNIEnv)(envPtr)
	activity := C.jobject(activityPtr)

	var jForSave C.jboolean
	if forSave {
		jForSave = 1
	}

	var jName C.jstring
	if suggestedName != "" {
		cName := C.CString(suggestedName)
		defer C.free(unsafe.Pointer(cName))
		jName = C.new_string_utf(env, cName)
	}

	C.call_show_file_picker(env, activity, jForSave, jName)
	return nil
}
