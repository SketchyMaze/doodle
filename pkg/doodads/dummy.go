package doodads

import (
	"git.kirsle.net/go/render"
	"git.kirsle.net/apps/doodle/pkg/level"
)

// NewDummy creates a placeholder dummy doodad with a giant "X" across it.
func NewDummy(size int) *Doodad {
	dummy := New(size)

	red := &level.Swatch{
		Color: render.Red,
		Name:  "missing color",
	}
	dummy.Palette.Swatches = []*level.Swatch{red}

	for i := 0; i < size; i++ {
		left := render.NewPoint(int32(i), int32(i))
		right := render.NewPoint(int32(size-i), int32(i))

		// Draw the stroke 2 pixels thick
		dummy.Layers[0].Chunker.Set(left, red)
		dummy.Layers[0].Chunker.Set(right, red)
		left.Y++
		right.Y++
		dummy.Layers[0].Chunker.Set(left, red)
		dummy.Layers[0].Chunker.Set(right, red)
	}

	return dummy
}
