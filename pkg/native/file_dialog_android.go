//go:build android
// +build android

package native

import (
	"errors"

	androidnative "git.kirsle.net/SketchyMaze/doodle/pkg/native/android"
)

func init() {
	FileDialogsReady = true
}

// OpenFile invokes Android's Storage Access Framework document picker
// (ACTION_OPEN_DOCUMENT) to choose an existing file, and returns a local
// path to a copy of it under this app's cache directory -- content://
// URIs, which is what the picker actually returns, aren't usable with
// Go's plain os.Open() the way a real filesystem path is.
//
// See pkg/native/android/filepicker.go for the JNI plumbing and
// MainActivity.java's showFilePicker()/onActivityResult() for the Java
// side. The title/filter arguments aren't used -- see PickFile's doc
// comment for why filter can't usefully translate to a SAF filter here,
// and the system picker UI supplies its own title chrome, so there's
// nowhere to put a custom one.
func OpenFile(title string, filter string) (string, error) {
	return androidnative.PickFile(false, "")
}

// SaveFile would invoke ACTION_CREATE_DOCUMENT to let the user pick a save
// destination, but round-tripping that back out to the content:// URI the
// user picked (after whatever code called this has actually written the
// file) isn't wired up -- there's no caller of native.SaveFile anywhere in
// this codebase today to design that round-trip against, and returning a
// local path that silently doesn't end up at the location the user picked
// would be a worse experience than an explicit "not supported" error.
// Revisit this if/when something actually calls it.
func SaveFile(title string, filter string) (string, error) {
	return "", errors.New("SaveFile is not supported on Android yet")
}
