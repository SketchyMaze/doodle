package doodle

import (
	"fmt"
	"time"

	"git.kirsle.net/apps/doodle/render"
	"github.com/veandco/go-sdl2/sdl"
)

// Frames to cache for FPS calculation.
const maxSamples = 100

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
	if !d.Debug {
		return
	}

	text := fmt.Sprintf(
		"FPS: %d (%dms)  (%d,%d)  size=%d  F12=screenshot",
		fpsCurrent,
		fpsSkipped,
		d.events.CursorX.Now,
		d.events.CursorY.Now,
		len(pixelHistory),
	)
	render.StrokedText(render.TextConfig{
		Text:        text,
		Size:        DebugTextSize,
		Color:       DebugTextColor,
		StrokeColor: DebugTextOutline,
		X:           DebugTextPadding,
		Y:           DebugTextPadding,
	})
}

// TrackFPS shows the current FPS once per second.
func (d *Doodle) TrackFPS(skipped uint32) {
	fpsFrames++
	fpsCurrentTicks = sdl.GetTicks()

	// Skip the first second.
	if fpsCurrentTicks < fpsInterval {
		return
	}

	if fpsLastTime < fpsCurrentTicks-fpsInterval {
		log.Debug("Uptime: %s  FPS: %d   deltaTicks: %d   skipped: %dms",
			time.Now().Sub(d.startTime),
			fpsCurrent,
			fpsCurrentTicks-fpsLastTime,
			skipped,
		)

		fpsLastTime = fpsCurrentTicks
		fpsCurrent = fpsFrames
		fpsFrames = 0
		fpsSkipped = skipped
	}
}
