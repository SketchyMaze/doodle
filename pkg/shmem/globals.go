package shmem

import (
	"fmt"

	"git.kirsle.net/apps/doodle/lib/render"
)

// Shared globals for easy access throughout the app.
// Not an ideal place to keep things but *shrug*
var (
	// Tick is incremented by the main game loop each frame.
	Tick uint64

	// Current render engine (i.e. SDL2 or HTML5 Canvas)
	// The level.Chunk.ToBitmap() uses this to cache a texture image.
	CurrentRenderEngine render.Engine

	// Globally available Flash() function so we can emit text to the Doodle UI.
	Flash func(string, ...interface{})

	// Ajax file cache for WASM use.
	AjaxCache map[string][]byte
)

func init() {
	AjaxCache = map[string][]byte{}

	// Default Flash function in case the app misconfigures it. Output to the
	// console in an obvious way.
	Flash = func(tmpl string, v ...interface{}) {
		fmt.Printf("[shmem.Flash] "+tmpl+"\n", v...)
	}
}