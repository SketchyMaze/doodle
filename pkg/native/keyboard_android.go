//go:build android
// +build android

package native

import "github.com/veandco/go-sdl2/sdl"

// ShowKeyboard requests the on-screen keyboard.
//
// SDL_StartTextInput() already does exactly this on Android: SDL's Android
// video backend wires SDL_HasScreenKeyboardSupport/SDL_ShowScreenKeyboard,
// and SDL_StartTextInput() calls that automatically (see
// src/video/SDL_video.c and src/video/android/SDL_androidkeyboard.c in the
// SDL2 sources -- android/scripts/fetch-sdl-sources.sh fetches a copy under
// android/third_party/SDL if you want to read it yourself). No JNI or
// Android-specific plumbing needed here at all -- go-sdl2 already wraps
// this SDL2 C function directly.
func ShowKeyboard() {
	sdl.StartTextInput()
}

// HideKeyboard dismisses the on-screen keyboard shown by ShowKeyboard.
func HideKeyboard() {
	sdl.StopTextInput()
}
