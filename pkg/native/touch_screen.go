package native

import (
	"git.kirsle.net/SketchyMaze/doodle/pkg/log"
	"git.kirsle.net/SketchyMaze/doodle/pkg/shmem"
	"git.kirsle.net/go/render/event"
)

// Common code to handle basic touch screen detection.

var (
	isTouchScreenMode  bool
	lastFingerDownTick uint64
)

// IsTouchScreenMode is activated when the user has touched the screen, and false when the mouse is moved.
func IsTouchScreenMode() bool {
	return isTouchScreenMode
}

/*
UpdateTouchScreenMode evaluates whether the primary game input is a touch screen rather than a mouse cursor.

The game always hides the OS cursor (if it exists) and may draw its own custom mouse cursor.

On a touch screen device (such as a mobile), the custom mouse cursor should not be drawn, either, as
it would jump around to wherever the user last touched and be distracting.

TouchScreenMode is activated when a touch event has been triggered (at least one finger was down).

TouchScreenMode deactivates and shows the mouse cursor when no finger is held down, and then a mouse
event has occurred. So if the user has a touch screen laptop, wiggling the actual mouse input will
bring the cursor back.
*/
func UpdateTouchScreenMode(ev *event.State) {

	// If a finger is presently down, record the current tick.
	if ev.IsFingerDown {
		lastFingerDownTick = shmem.Tick
	}

	if !isTouchScreenMode {
		// We are NOT in touch screen mode. Touching the screen will change this.
		if ev.IsFingerDown {
			log.Info("TouchScreenMode ON")
			isTouchScreenMode = true
		}
	} else {
		// We ARE in touch screen mode. Wait for all fingers to be lifted.
		if !ev.IsFingerDown {

			// If we have registered a mouse event a few ticks after the finger was
			// removed, it is a real mouse cursor and we exit touch screen mode.
			if ev.IsMouseEvent && shmem.Tick-lastFingerDownTick > 5 {
				log.Info("TouchScreenMode OFF")
				isTouchScreenMode = false
			}
		}
	}
}
