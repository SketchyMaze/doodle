//go:build android
// +build android

package native

/*
#include <SDL2/SDL.h>
*/
import "C"
import (
	"unsafe"

	"git.kirsle.net/SketchyMaze/doodle/pkg/balance"
	"git.kirsle.net/SketchyMaze/doodle/pkg/branding"
	"git.kirsle.net/SketchyMaze/doodle/pkg/log"
)

// OpenURL opens a web browser (really, whatever app the device has
// registered for an ACTION_VIEW intent on http(s) URLs) to the given URL.
//
// This calls SDL_OpenURL() directly via cgo rather than hand-rolling JNI:
// go-sdl2 doesn't wrap it (checked -- it's not in the v0.4.38 module this
// project pins, despite SDL2's own TODO.md claiming otherwise), but the C
// function itself has been in SDL2 since 2.0.14, and its Android
// implementation already does the exact Intent.ACTION_VIEW dance this
// would otherwise need custom MainActivity.java code for. No #cgo pragma
// is needed here for the same reason go-sdl2's own sdl package doesn't
// have one anywhere: CGO_CFLAGS/CGO_LDFLAGS (set by
// android/scripts/build-native-libs.sh for the whole build) already point
// at this project's SDL2 headers/libs.
func OpenURL(url string) {
	cURL := C.CString(url)
	defer C.free(unsafe.Pointer(cURL))

	if C.SDL_OpenURL(cURL) != 0 {
		log.Error("native.OpenURL(%s): %s", url, C.GoString(C.SDL_GetError()))
	}
}

// OpenLocalURL would open a web browser to a local HTML path on desktop,
// but that doesn't translate to Android:
//
//   - The guidebook (balance.GuidebookPath, the only caller that matters
//     in practice -- see pkg/common_menubar.go and pkg/doodle.go) isn't
//     bundled into the Android build at all (unlike android/'s rtp/ audio
//     -- see rtp/embed_android.go -- nobody's embedded the much larger
//     guidebook/ tree for Android), so open the online copy instead
//     (branding.GuidebookURL, the same URL the desktop build's own
//     "Guidebook Online" menu item already uses).
//   - Other callers (the screenshots directory, an exported doodad's HTML
//     docs) point at paths under this app's private storage, which even a
//     file:// URL couldn't usefully open in an external browser app due to
//     Android's scoped storage restrictions. There's no online equivalent
//     to redirect those to, so they're just logged and otherwise ignored.
func OpenLocalURL(path string) {
	if path == balance.GuidebookPath {
		OpenURL(branding.GuidebookURL)
		return
	}
	log.Warn("native.OpenLocalURL(%s): local file browsing isn't supported on Android", path)
}
