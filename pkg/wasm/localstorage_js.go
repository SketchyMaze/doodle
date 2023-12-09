//go:build js && wasm
// +build js,wasm

package wasm

import (
	"strings"
	"syscall/js"

	"git.kirsle.net/SketchyMaze/doodle/pkg/log"
)

// StorageKeys returns the list of localStorage keys matching a prefix.
func StorageKeys(prefix string) []string {
	keys := js.Global().Get("Object").Call("keys", js.Global().Get("localStorage"))

	var result []string
	for i := 0; i < keys.Length(); i++ {
		value := keys.Index(i).String()
		if strings.HasPrefix(value, prefix) {
			result = append(result,
				strings.TrimPrefix(keys.Index(i).String(), prefix),
			)
		}
	}
	log.Info("LS KEYS: %+v", result)
	return result
}

// SetSession sets a text value on sessionStorage.
func SetSession(key string, value string) {
	js.Global().Get("localStorage").Call("setItem", key, value)
}

// GetSession retrieves a text value from sessionStorage.
func GetSession(key string) (string, bool) {
	var value js.Value
	value = js.Global().Get("localStorage").Call("getItem", key)
	return value.String(), value.Type() == js.TypeString
}
