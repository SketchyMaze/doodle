// +build js,wasm

package wasm

import (
	"syscall/js"
)

// SetSession sets a text value on sessionStorage.
func SetSession(key string, value string) {
	// b64 := base64.StdEncoding.EncodeToString(value)
	panic("SesSession: " + key)
	js.Global().Get("sessionStorage").Call("setItem", key, value)
}

// GetSession retrieves a text value from sessionStorage.
func GetSession(key string) (string, bool) {
	panic("GetSession: " + key)
	var value js.Value
	value = js.Global().Get("sessionStorage").Call("getItem", key)
	return value.String(), value.Type() == js.TypeString
}
