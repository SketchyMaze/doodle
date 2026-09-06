// Command doodle-android is the Android NDK entry point for SketchyMaze.
//
// It is not built by `go build ./...` on desktop platforms -- it only
// compiles under GOOS=android (see the "android" build tag) because it
// imports "C" and exports a C symbol for SDL's Java glue (org.libsdl.app.
// SDLActivity) to call via JNI once it has dlopen'd this package's compiled
// output as libdoodle.so.
//
// See android/README.md for the full build pipeline: this file gets built
// with `go build -buildmode=c-shared`, once per target ABI, into
// android/app/src/main/jniLibs/<abi>/libdoodle.so.
package main

/*
#include <jni.h>
#include <stdlib.h>

// See pkg/native/android/filepicker.go's own comment on why JNI vtable
// calls (env->Foo(env, ...) in C) need a small static C helper each rather
// than being called directly from Go/cgo expressions.
static const char *get_string_utf_chars(JNIEnv *env, jstring str) {
    if (str == NULL) {
        return NULL;
    }
    return (*env)->GetStringUTFChars(env, str, NULL);
}

static void release_string_utf_chars(JNIEnv *env, jstring str, const char *chars) {
    if (str != NULL && chars != NULL) {
        (*env)->ReleaseStringUTFChars(env, str, chars);
    }
}
*/
import "C"

import (
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"unsafe"

	"git.kirsle.net/SketchyMaze/doodle/assets"
	doodle "git.kirsle.net/SketchyMaze/doodle/pkg"
	"git.kirsle.net/SketchyMaze/doodle/pkg/balance"
	"git.kirsle.net/SketchyMaze/doodle/pkg/branding"
	"git.kirsle.net/SketchyMaze/doodle/pkg/log"
	"git.kirsle.net/SketchyMaze/doodle/pkg/native"
	androidnative "git.kirsle.net/SketchyMaze/doodle/pkg/native/android"
	"git.kirsle.net/SketchyMaze/doodle/pkg/sound"
	"git.kirsle.net/SketchyMaze/doodle/pkg/sprites"
	"git.kirsle.net/SketchyMaze/doodle/pkg/usercfg"
	"git.kirsle.net/SketchyMaze/doodle/pkg/userdir"
	"git.kirsle.net/SketchyMaze/doodle/rtp"
	golog "git.kirsle.net/go/log"
	"git.kirsle.net/go/render/sdl"
	sdl2 "github.com/veandco/go-sdl2/sdl"
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

//export SDL_main
func SDL_main() {
	runtime.LockOSThread()

	// Repoints pkg/userdir at real, writable paths before anything reads
	// from it -- see setupUserdir()'s own doc comment for why this exists
	// instead of relying on the $HOME/$XDG_* env vars MainActivity.
	// loadLibraries() sets.
	setupUserdir()

	if fh, err := golog.NewFileTee(userdir.LogFile); err == nil {
		log.Logger.Config.Writer = fh
	} else {
		sdl2.Log("doodle-android: golog.NewFileTee(%q) failed: %s", userdir.LogFile, err)
	}

	log.Info("Starting %s %s (Android)", branding.AppName, branding.Version)

	if err := usercfg.Load(); err != nil {
		log.Error("Error loading user settings (defaults will be used): %s", err)
	}

	// There's no keyboard/mouse to fall back to: always treat this as a
	// touch-screen device (see pkg/native/touch_screen.go).
	native.ForceTouchScreenModeAlwaysOn = true

	engine := sdl.New(
		branding.AppName,
		balance.Width,
		balance.Height,
	)

	sdl2.GameControllerEventState(1)

	if fonts, err := assets.AssetDir("assets/fonts"); err == nil {
		for _, file := range fonts {
			data, err := assets.Asset("assets/fonts/" + file)
			if err != nil {
				log.Error("Couldn't load font %s: %s", file, err)
				continue
			}
			sdl.InstallFont(file, data)
		}
	} else {
		log.Error("Couldn't list embedded fonts: %s", err)
	}

	extractRTP()
	sound.PreloadAll()

	game := doodle.New(false, engine)
	game.SetupEngine()

	// doodle.New() seeds its width/height fields from the hardcoded
	// balance.Width/Height (1024x768), which is only ever correct by
	// coincidence: on desktop, SDL_CreateWindow() actually creates a window
	// at that exact size, but on Android the OS always overrides the real
	// window size (see android/app/src/main/java/net/kirsle/sketchymaze/
	// ScaledSDLSurface.java), so those fields are wrong from construction
	// until a genuine WindowResized event corrects them inside game.Run()'s
	// event loop -- which produces a landscape-shaped initial UI on a
	// portrait launch (or vice versa) until the player rotates the device
	// once. SetupEngine() above has already created the real window, so
	// sync this up front instead of waiting on that first resize event.
	if w, h := engine.WindowSize(); w > 0 && h > 0 {
		game.SetWindowSize(w, h)
	}

	// Reload usercfg now that SetupEngine has detected touch input.
	usercfg.Load()

	engine.ShowCursor(false)

	if renderer, ok := game.Engine.(*sdl.Renderer); ok {
		if icon, err := sprites.LoadImage(game.Engine, balance.WindowIcon); err == nil {
			renderer.SetWindowIcon(icon.Image)
		}
	}

	if err := game.Run(); err != nil {
		log.Error("game.Run: %s", err)
	}

	// When game.Run() returns (the player quit, e.g. via the in-game File
	// menu), returning from this exported function just hands control back
	// to Java's SDLMain.run(), which calls Activity.finish() -- ending the
	// *Activity*, but not necessarily the underlying Android *process*.
	// Android often keeps a finished app's process alive in the background
	// for a fast relaunch, and that's exactly the problem here: dlopen()
	// (via System.loadLibrary()) doesn't re-run a shared library's init
	// code on a second load within the same process, so a relaunched
	// Activity in the surviving process would call this exported SDL_main
	// a second time against a Go runtime, and SDL2 C-level global state,
	// that already went through one full run and was never designed to be
	// re-entered -- observed as the relaunched app immediately dying in
	// the background (nothing in logcat past the usual startup lines)
	// until the whole task is swiped away in Recents to force a genuinely
	// fresh process. os.Exit() terminates the entire process outright
	// (unlike returning, which only ends this one call), guaranteeing the
	// next launch always starts from a real clean slate.
	os.Exit(0)
}

// main is required for `go build -buildmode=c-shared` but is never called:
// SDLActivity invokes the exported SDL_main function above via JNI instead.
func main() {}

// Java_com_sketchymaze_doodle_MainActivity_nativeFilePickerResult is called
// by MainActivity.onActivityResult() (via a `private native` JNI method
// declaration -- see MainActivity.java) once the user has picked a file (or
// canceled) in the Storage Access Framework picker
// pkg/native/android.PickFile() launched.
//
// This has to live here, in this package's own cgo block, rather than in
// pkg/native/android alongside PickFile(): only a Go "package main" built
// with -buildmode=c-shared can //export a C symbol for the JVM to call by
// name, and cgo types like C.jstring aren't interchangeable across
// separate packages' own "C" pseudo-packages -- so the jstring gets
// converted to a plain Go string here, and only that plain string crosses
// the package boundary into androidnative.DeliverResult().
//
// The exported symbol name follows the standard JNI native-method naming
// convention (Java_<package_with_underscores>_<Class>_<method>), matching
// package com.sketchymaze.doodle's MainActivity.
//
//export Java_com_sketchymaze_doodle_MainActivity_nativeFilePickerResult
func Java_com_sketchymaze_doodle_MainActivity_nativeFilePickerResult(env *C.JNIEnv, thiz C.jobject, jPath C.jstring, jOk C.jboolean) {
	var path string
	// C.jstring doesn't support a direct `== nil` comparison under cgo
	// (see the same issue, worked around the same way, in
	// pkg/native/android/filepicker.go's launchPicker()).
	if unsafe.Pointer(jPath) != nil {
		cPath := C.get_string_utf_chars(env, jPath)
		path = C.GoString(cPath)
		C.release_string_utf_chars(env, jPath, cPath)
	}

	androidnative.DeliverResult(path, jOk != 0)
}

// setupUserdir repoints pkg/userdir's exported directory/file path vars at
// real, writable locations under this app's private internal storage,
// obtained directly via SDL's own Android JNI accessor
// (SDL_AndroidGetInternalStoragePath, wrapped as sdl2.AndroidGetInternal
// StoragePath()) rather than through pkg/userdir's own init(), which
// resolves everything through github.com/kirsle/configdir's XDG-style
// fallback: $HOME/$XDG_CONFIG_HOME/$XDG_CACHE_HOME.
//
// MainActivity.loadLibraries() does set those env vars (to this app's
// getFilesDir()) before this .so is even dlopen'd, specifically so
// pkg/userdir's init() -- which runs at dlopen time, before any of our own
// code gets a chance to run -- resolves correctly on its own. In practice
// that didn't pan out: even with Java-side Os.setenv() confirmed to
// succeed without throwing, the resulting directories never appeared on
// disk, which points at Go's os.Getenv() not actually observing whatever
// Java's setenv() changed (Go's runtime reads its own copy of the
// environment once at startup rather than calling libc getenv() live) --
// rather than keep chasing that, this sidesteps it entirely by asking
// Android directly for a real path and overwriting pkg/userdir's vars
// with it, plus re-creating the directories pkg/userdir's own init() would
// have (unsuccessfully) tried to create.
func setupUserdir() {
	base := sdl2.AndroidGetInternalStoragePath()
	if base == "" {
		sdl2.Log("doodle-android: AndroidGetInternalStoragePath() returned empty; leaving userdir paths as-is")
		return
	}

	userdir.ProfileDirectory = base
	userdir.LevelDirectory = filepath.Join(base, "levels")
	userdir.LevelPackDirectory = filepath.Join(base, "levelpacks")
	userdir.DoodadDirectory = filepath.Join(base, "doodads")
	userdir.CampaignDirectory = filepath.Join(base, "campaigns")
	userdir.ScreenshotDirectory = filepath.Join(base, "screenshots")
	userdir.SaveFile = filepath.Join(base, "savegame.json")
	userdir.LogFile = filepath.Join(base, "logfile.txt")
	userdir.CacheDirectory = filepath.Join(base, "cache")
	userdir.FontDirectory = filepath.Join(base, "cache", "fonts")

	for _, dir := range []string{
		userdir.LevelDirectory,
		userdir.LevelPackDirectory,
		userdir.DoodadDirectory,
		userdir.CampaignDirectory,
		userdir.ScreenshotDirectory,
		userdir.FontDirectory,
	} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			sdl2.Log("doodle-android: MkdirAll(%q) failed: %s", dir, err)
		}
	}
}

// extractRTP unpacks the embedded rtp/ runtime package (see
// rtp/embed_android.go) into the app's private cache directory and
// repoints pkg/sound's SoundRoot/MusicRoot at the extracted copy.
//
// On desktop, rtp/ ships as loose files next to the executable and
// pkg/sound resolves SoundRoot/MusicRoot as plain relative paths against
// the process's working directory. Android has neither "next to the
// executable" nor a meaningful working directory, so there's nothing for
// those relative paths to resolve against -- every sound/music load
// silently failed until this ran. Reassigning the vars here (rather than
// os.Chdir()ing the whole process) keeps this entirely Android-specific
// and one-directional: nothing else in the codebase reads from "rtp" (see
// rtp/embed_android.go's own doc comment), so this can't affect anything
// but sound loading.
func extractRTP() {
	dest := filepath.Join(userdir.CacheDirectory, "rtp")

	err := fs.WalkDir(rtp.Embedded, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		data, err := rtp.Embedded.ReadFile(path)
		if err != nil {
			return err
		}

		out := filepath.Join(dest, path)
		if err := os.MkdirAll(filepath.Dir(out), 0755); err != nil {
			return err
		}
		return os.WriteFile(out, data, 0644)
	})
	if err != nil {
		log.Error("extractRTP: couldn't unpack embedded runtime package: %s", err)
		return
	}

	sound.SoundRoot = filepath.Join(dest, "sfx")
	sound.MusicRoot = filepath.Join(dest, "music")
}
