// +build js,wasm

package native

import "syscall/js"

// OpenURL opens a new window to the given URL, for WASM environment.
func OpenURL(url string) {
	js.Global().Get("window").Call("open", url)
}

func OpenLocalURL(url string) {
	OpenURL(url)
}
