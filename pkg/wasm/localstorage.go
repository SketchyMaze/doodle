// +build !js

package wasm

// SetSession sets a binary value on sessionStorage.
// This is a no-op when not in wasm.
func SetSession(key string, value string) {
}

// GetSession retrieves a binary value from sessionStorage.
// This is a no-op when not in wasm.
func GetSession(key string) (string, bool) {
	return "", false
}
