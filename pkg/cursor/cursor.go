// Package cursor handles custom mouse cursor sprite images.
package cursor

import (
	"git.kirsle.net/SketchyMaze/doodle/pkg/balance"
	"git.kirsle.net/SketchyMaze/doodle/pkg/log"
	"git.kirsle.net/SketchyMaze/doodle/pkg/native"
	"git.kirsle.net/SketchyMaze/doodle/pkg/shmem"
	"git.kirsle.net/SketchyMaze/doodle/pkg/sprites"
	"git.kirsle.net/go/render"
	"git.kirsle.net/go/ui"
)

type Cursor struct {
	Filename string
	Sprite   *ui.Image
	Offset   render.Point
}

// Current selected cursor to draw on screen.
var Current *Cursor

// NoCursor hides the cursor entirely.
var NoCursor = &Cursor{}

// Draw the cursor on screen. NOTE: Does not draw on touchscreen devices.
func Draw(e render.Engine) {
	if native.IsTouchScreenMode() {
		return
	}

	if Current == nil {
		Current = NewPointer(e)
	}

	if Current.Sprite != nil {
		Current.Sprite.Present(e, shmem.Cursor)
	}
}

// NewPointer initializes the default pointer cursor.
func NewPointer(e render.Engine) *Cursor {
	if pointer != nil {
		return pointer
	}
	img, err := sprites.LoadImage(e, balance.CursorIcon)
	if err != nil {
		log.Error("NewPointer: %s", err)
	}
	return &Cursor{
		Filename: balance.CursorIcon,
		Sprite:   img,
	}
}

// NewPencil initializes the pencil cursor.
func NewPencil(e render.Engine) *Cursor {
	if pencil != nil {
		return pencil
	}
	img, err := sprites.LoadImage(e, balance.PencilIcon)
	if err != nil {
		log.Error("NewPencil: %s", err)
	}
	return &Cursor{
		Filename: balance.PencilIcon,
		Sprite:   img,
	}
}

// NewFlood initializes the Flood cursor.
func NewFlood(e render.Engine) *Cursor {
	if pencil != nil {
		return pencil
	}
	img, err := sprites.LoadImage(e, balance.FloodCursor)
	if err != nil {
		log.Error("NewFlood: %s", err)
	}
	return &Cursor{
		Filename: balance.FloodCursor,
		Sprite:   img,
	}
}

// Cached singletons of the cursors.
var (
	pointer *Cursor
	pencil  *Cursor
	flood   *Cursor
)
