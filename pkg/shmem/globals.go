package shmem

import (
	"fmt"

	"git.kirsle.net/go/render"
)

// Shared globals for easy access throughout the app.
// Not an ideal place to keep things but *shrug*
var (
	// Tick is incremented by the main game loop each frame.
	Tick uint64

	// Current position of the cursor relative to the window.
	Cursor render.Point

	// Current render engine (i.e. SDL2 or HTML5 Canvas)
	// The level.Chunk.ToBitmap() uses this to cache a texture image.
	CurrentRenderEngine render.Engine

	// Offline mode, if True then the updates check in MainScene is skipped.
	OfflineMode bool

	// Globally available Flash() function so we can emit text to the Doodle UI.
	Flash      func(string, ...interface{})
	FlashError func(string, ...interface{})

	// Globally available Prompt() function.
	Prompt func(string, func(string))

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
