package shmem

import (
	"fmt"

	"git.kirsle.net/apps/doodle/lib/render"
)

// Shared globals for easy access throughout the app.
// Not an ideal place to keep things but *shrug*
var (
	// Current render engine (i.e. SDL2 or HTML5 Canvas)
	// The level.Chunk.ToBitmap() uses this to cache a texture image.
	CurrentRenderEngine render.Engine

	// Globally available Flash() function so we can emit text to the Doodle UI.
	Flash func(string, ...interface{})
)

func init() {
	// Default Flash function in case the app misconfigures it. Output to the
	// console in an obvious way.
	Flash = func(tmpl string, v ...interface{}) {
		fmt.Printf("[shmem.Flash] "+tmpl+"\n", v...)
	}
}
