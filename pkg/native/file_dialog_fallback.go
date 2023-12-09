//go:build js && wasm
// +build js,wasm

package native

import (
	"git.kirsle.net/SketchyMaze/doodle/pkg/shmem"
)

func init() {
	FileDialogsReady = false
}

// OpenFile fallback uses the shell prompt.
func OpenFile(title string, filter string) (string, error) {
	shmem.Prompt(title, func(value string) {

	})
	return "", nil
}
