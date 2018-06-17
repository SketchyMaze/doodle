// Package render manages the SDL rendering context for Doodle.
package render

import "github.com/veandco/go-sdl2/sdl"

// Renderer is a singleton instance of the SDL renderer.
var Renderer *sdl.Renderer
