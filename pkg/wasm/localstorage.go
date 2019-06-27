// +build !js

package wasm

// StorageKeys returns the list of localStorage keys matching a prefix.
// This is a no-op when not in wasm.
func StorageKeys(prefix string) []string {
	return []string{}
}

// SetSession sets a binary value on sessionStorage.
// This is a no-op when not in wasm.
func SetSession(key string, value string) {
}

// GetSession retrieves a binary value from sessionStorage.
// This is a no-op when not in wasm.
func GetSession(key string) (string, bool) {
	return "", false
}
