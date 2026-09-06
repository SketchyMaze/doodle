//go:build !android
// +build !android

package native

// ShowKeyboard requests the on-screen keyboard, on platforms that have one
// separate from a physical keyboard. No-op everywhere except Android; see
// keyboard_android.go.
func ShowKeyboard() {}

// HideKeyboard dismisses the on-screen keyboard shown by ShowKeyboard.
func HideKeyboard() {}
