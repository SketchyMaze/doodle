package doodle

import (
	"fmt"
	"strings"

	"git.kirsle.net/apps/doodle/lib/render"
	"git.kirsle.net/apps/doodle/lib/ui"
	"git.kirsle.net/apps/doodle/pkg/balance"
	"git.kirsle.net/apps/doodle/pkg/doodads"
)

// Frames to cache for FPS calculation.
const maxSamples = 100

// Debug mode options, these can be enabled in the dev console
// like: boolProp DebugOverlay true
var (
	DebugOverlay   = true
	DebugCollision = true

	DebugTextPadding int32 = 8
	DebugTextSize          = 24
	DebugTextColor         = render.SkyBlue
	DebugTextStroke        = render.Grey
	DebugTextShadow        = render.Black
)

var (
	fpsCurrentTicks uint32 // current time we get sdl.GetTicks()
	fpsLastTime     uint32 // last time we printed the fpsCurrentTicks
	fpsCurrent      int
	fpsFrames       int
	fpsSkipped      uint32
	fpsInterval     uint32 = 1000

	// XXX: some opt-in WorldIndex variables for the debug overlay.
	// This is the world pixel that the mouse cursor is over,
	// the Cursor + Scroll position of the canvas.
	debugWorldIndex render.Point
)

// DrawDebugOverlay draws the debug FPS text on the SDL canvas.
func (d *Doodle) DrawDebugOverlay() {
	if !d.Debug || !DebugOverlay {
		return
	}

	var (
		darken        = balance.DebugStrokeDarken
		Yoffset int32 = 20 // leave room for the menu bar
		Xoffset int32 = 5
		keys          = []string{
			"  FPS:",
			"Scene:",
			"Pixel:",
			"Mouse:",
		}
		values = []string{
			fmt.Sprintf("%d   (skip: %dms)", fpsCurrent, fpsSkipped),
			d.Scene.Name(),
			debugWorldIndex.String(),
			fmt.Sprintf("%d,%d", d.event.CursorX.Now, d.event.CursorY.Now),
		}
	)

	key := ui.NewLabel(ui.Label{
		Text: strings.Join(keys, "\n"),
		Font: render.Text{
			Size:         balance.DebugFontSize,
			FontFilename: balance.ShellFontFilename,
			Color:        balance.DebugLabelColor,
			Stroke:       balance.DebugLabelColor.Darken(darken),
		},
	})
	key.Compute(d.Engine)
	key.Present(d.Engine, render.NewPoint(
		DebugTextPadding+Xoffset,
		DebugTextPadding+Yoffset,
	))

	value := ui.NewLabel(ui.Label{
		Text: strings.Join(values, "\n"),
		Font: render.Text{
			Size:         balance.DebugFontSize,
			FontFilename: balance.DebugFontFilename,
			Color:        balance.DebugValueColor,
			Stroke:       balance.DebugValueColor.Darken(darken),
		},
	})
	value.Compute(d.Engine)
	value.Present(d.Engine, render.NewPoint(
		DebugTextPadding+Xoffset+key.Size().W+DebugTextPadding,
		DebugTextPadding+Yoffset, // padding to not overlay menu bar
	))
}

// DrawCollisionBox draws the collision box around a Doodad.
func (d *Doodle) DrawCollisionBox(actor doodads.Actor) {
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
