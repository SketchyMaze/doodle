package doodle

import (
	"fmt"

	"git.kirsle.net/apps/doodle/doodads"
	"git.kirsle.net/apps/doodle/render"
)

// Frames to cache for FPS calculation.
const maxSamples = 100

// Debug mode options, these can be enabled in the dev console
// like: boolProp DebugOverlay true
var (
	DebugOverlay   = true
	DebugCollision = true
)

var (
	fpsCurrentTicks uint32 // current time we get sdl.GetTicks()
	fpsLastTime     uint32 // last time we printed the fpsCurrentTicks
	fpsCurrent      int
	fpsFrames       int
	fpsSkipped      uint32
	fpsInterval     uint32 = 1000
)

// DrawDebugOverlay draws the debug FPS text on the SDL canvas.
func (d *Doodle) DrawDebugOverlay() {
	if !d.Debug || !DebugOverlay {
		return
	}

	label := fmt.Sprintf(
		"FPS: %d (%dms)  S:%s  F12=screenshot",
		fpsCurrent,
		fpsSkipped,
		d.Scene.Name(),
	)

	err := d.Engine.DrawText(
		render.Text{
			Text:   label,
			Size:   24,
			Color:  DebugTextColor,
			Stroke: DebugTextStroke,
			Shadow: DebugTextShadow,
		},
		render.Point{
			X: DebugTextPadding,
			Y: DebugTextPadding,
		},
	)
	if err != nil {
		log.Error("DrawDebugOverlay: text error: %s", err.Error())
	}
}

// DrawCollisionBox draws the collision box around a Doodad.
func (d *Doodle) DrawCollisionBox(actor doodads.Doodad) {
	if !d.Debug || !DebugCollision {
		return
	}

	var (
		rect = doodads.GetBoundingRect(actor)
		box  = doodads.GetCollisionBox(rect)
	)

	d.Engine.DrawLine(render.DarkGreen, box.Top[0], box.Top[1])
	d.Engine.DrawLine(render.DarkBlue, box.Bottom[0], box.Bottom[1])
	d.Engine.DrawLine(render.DarkYellow, box.Left[0], box.Left[1])
	d.Engine.DrawLine(render.Red, box.Right[0], box.Right[1])
}

// TrackFPS shows the current FPS once per second.
func (d *Doodle) TrackFPS(skipped uint32) {
	fpsFrames++
	fpsCurrentTicks = d.Engine.GetTicks()

	// Skip the first second.
	if fpsCurrentTicks < fpsInterval {
		return
	}

	if fpsLastTime < fpsCurrentTicks-fpsInterval {
		fpsLastTime = fpsCurrentTicks
		fpsCurrent = fpsFrames
		fpsFrames = 0
		fpsSkipped = skipped
	}
}
