// Package loadscreen implements a modal "Loading" screen for the game, which
// can be shown or hidden by gameplay scenes as needed.
package loadscreen

import (
	"strings"

	"git.kirsle.net/SketchyMaze/doodle/pkg/balance"
	"git.kirsle.net/SketchyMaze/doodle/pkg/level"
	"git.kirsle.net/SketchyMaze/doodle/pkg/log"
	"git.kirsle.net/SketchyMaze/doodle/pkg/shmem"
	"git.kirsle.net/SketchyMaze/doodle/pkg/uix"
	"git.kirsle.net/go/render"
	"git.kirsle.net/go/ui"
)

// Configuration values.
const (
	ProgressWidth  = 300
	ProgressHeight = 16
)

// State variables for the loading screen.
var (
	visible      bool
	withProgress bool
	subtitle     string // custom subtitle text, SetSubtitle().

	// Animated title bar
	titleBase = "Loading"
	animState = 0
	animation = []string{
		".  ",
		".. ",
		"...",
		" ..",
		"  .",
		"   ",
	}
	animSpeed uint64 = 32
	titleVar  string

	// UI widgets.
	window         *ui.Frame
	canvas         *uix.Canvas
	secondary      *ui.Label // subtitle text
	progressTrough *ui.Frame
	progressBar    *ui.Frame
	progressText   *ui.Label
)

// Show the basic loading screen without a progress bar.
func Show() {
	setup()
	visible = true
	withProgress = false
	subtitle = ""
}

// ShowWithProgress initializes the loading screen with a progress bar starting at zero.
func ShowWithProgress() {
	setup()
	visible = true
	withProgress = true
	subtitle = ""

	// TODO: hide the progress trough. With the new LoadUnloadChunks,
	// the progress bar does nothing - all the loading time is spent
	// reading the level from disk and then it starts.
	withProgress = false

	SetProgress(0)
}

// SetSubtitle specifies secondary text beneath the Loading banner.
// The subtitle is blanked on Show() and ShowWithProgress() and must
// be specified by the caller if desired. Pass multiple values for
// multiple lines of text.
func SetSubtitle(value ...string) {
	subtitle = strings.Join(value, "\n")
}

// IsActive returns whether the loading screen is currently visible.
func IsActive() bool {
	return visible
}

// Resized the window.
func Resized() {
	if visible {
		size := render.NewRect(shmem.CurrentRenderEngine.WindowSize())
		window.Resize(size)
	}
}

/*
Hide the loading screen.

NOTICE: the loadscreen is hidden on an async goroutine and it is NOT SAFE to clean up
textures used by the wallpaper images, but this is OK because the loadscreen uses the
same wallpaper every time and is called many times during gameplay, it can hold its
textures.
*/
func Hide() {
	visible = false
}

// SetProgress sets the current progress value for loading screens having a progress bar.
func SetProgress(v float64) {
	// Resize the progress bar in the trough.
	if progressTrough != nil {
		var (
			troughSize = progressTrough.Size()
			height     = progressBar.Size().H
		)
		progressBar.Resize(render.Rect{
			W: int(float64(troughSize.W-4) * v),
			H: height,
		})
	}
}

// Common function to initialize the loading screen.
func setup() {
	if window != nil {
		return
	}

	titleVar = titleBase + animation[animState]

	// Create the parent container that will stretch full screen.
	window = ui.NewFrame("Loadscreen Window")
	window.SetBackground(render.RGBA(0, 0, 1, 40)) // makes the wallpaper darker? :/

	// "Loading" text.
	label := ui.NewLabel(ui.Label{
		TextVariable: &titleVar,
		Font:         balance.LoadScreenFont,
	})
	label.Compute(shmem.CurrentRenderEngine)
	window.Place(label, ui.Place{
		Top:    128,
		Center: true,
	})

	// Subtitle text.
	secondary = ui.NewLabel(ui.Label{
		TextVariable: &subtitle,
		Font:         balance.LoadScreenSecondaryFont,
	})
	window.Place(secondary, ui.Place{
		Top:    128 + label.Size().H + 64,
		Center: true,
	})

	// Progress bar.
	progressTrough = ui.NewFrame("Progress Trough")
	progressTrough.Configure(ui.Config{
		Width:       ProgressWidth,
		Height:      ProgressHeight,
		BorderSize:  2,
		BorderStyle: ui.BorderSunken,
		Background:  render.DarkGrey,
	})
	window.Place(progressTrough, ui.Place{
		// Nestle it between the Title and Subtitle.
		Center: true,
		Top:    128 + label.Size().H + 16,
	})

	progressBar = ui.NewFrame("Progress Bar")
	progressBar.Configure(ui.Config{
		Width:      0,
		Height:     ProgressHeight - 4,
		Background: render.Green,
	})
	progressTrough.Pack(progressBar, ui.Pack{
		Side: ui.W,
	})
}

// Loop is called on every game loop. If the loadscreen is not active, nothing happens.
// Otherwise the loading screen UI is drawn to screen.
func Loop(windowSize render.Rect, e render.Engine) {
	if !visible {
		return
	}

	if window != nil {
		// Initialize the wallpaper canvas?
		if canvas == nil {
			canvas = uix.NewCanvas(128, false)
			canvas.LoadLevel(&level.Level{
				Chunker:   level.NewChunker(100),
				Palette:   level.NewPalette(),
				PageType:  level.Bounded,
				Wallpaper: "blueprint.png",
			})
		}
		canvas.Resize(windowSize)
		canvas.Compute(e)
		canvas.Present(e, render.Origin)

		window.Resize(windowSize)
		window.Compute(e)
		window.Present(e, render.Origin)

		// Show/hide the progress bar.
		progressTrough.Compute(e)
		if withProgress && progressTrough.Hidden() {
			progressTrough.Show()
		} else if !withProgress && !progressTrough.Hidden() {
			progressTrough.Hide()
		}

		// Show/hide the subtitle text.
		if len(subtitle) > 0 && secondary.Hidden() {
			secondary.Show()
		} else if subtitle == "" && !secondary.Hidden() {
			secondary.Hide()
		}

		// Animate the ellipses.
		if shmem.Tick%animSpeed == 0 {
			titleVar = titleBase + animation[animState]
			animState++
			if animState >= len(animation) {
				animState = 0
			}
		}

	}
}

// PreloadAllChunkBitmaps is a helper function to eager cache all bitmap
// images of the chunks in a level drawing. It is designed to work with the
// loading screen and will set the Progress percent based on the total number
// of chunks vs. chunks remaining to pre-cache bitmaps from.
func PreloadAllChunkBitmaps(chunker *level.Chunker) {
	// If we're using the smarter (experimental) chunk loader, return.
	if balance.Feature.LoadUnloadChunk {
		return
	}

	loadChunksTarget := len(chunker.Chunks)

	// Skipping the eager rendering of chunks?
	if !balance.EagerRenderLevelChunks {
		log.Info("PreloadAllChunkBitmaps: skipping eager render")
		return
	}

	for {
		remaining := chunker.PrerenderN(10)

		// Set the load screen progress % based on number of chunks to render.
		if loadChunksTarget > 0 {
			percent := float64(loadChunksTarget-remaining) / float64(loadChunksTarget)
			SetProgress(percent)
		}

		if remaining == 0 {
			break
		}
	}
}
