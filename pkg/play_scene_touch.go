package doodle

import (
	"time"

	"git.kirsle.net/apps/doodle/pkg/balance"
	"git.kirsle.net/apps/doodle/pkg/log"
	"git.kirsle.net/apps/doodle/pkg/usercfg"
	"git.kirsle.net/go/render"
	"git.kirsle.net/go/render/event"
	"git.kirsle.net/go/ui"
)

/*
Touchscreen Control functionality used in the Play Scene.
*/

// LoopTouchable is called as part of PlayScene.Loop while the simulation is running.
//
// It looks for touch events on proportional regions of the window and emulates key
// input bindings to move the character, jump, etc.
//
// TODO: this function manipulates the event.State to set Up, Down, Left, Right and
// Space keys and may need love for reconfigurable keybinds later.
func (s *PlayScene) LoopTouchable(ev *event.State) {
	var (
		middle = s.touchGetMiddleBox()
		cursor = render.NewPoint(ev.CursorX, ev.CursorY)
	)

	// Don't do any of this if the mouse is over the menu bar, so
	// clicking on the menus doesn't make the character move or jump.
	if cursor.Inside(s.menubar.Rect()) || s.supervisor.GetModal() != nil ||
		s.supervisor.IsPointInWindow(cursor) {
		return
	}

	// Detect if the player is idle.
	// Idle means that they are not holding any directional or otherwise input key.
	// Keyboard inputs and touch events from this function will set these keys.
	// See if it stays unset long enough to consider idle.
	var isGrounded = (s.Player.HasGravity() && s.Player.Grounded()) || !s.Player.HasGravity()
	if !ev.Up && !ev.Down && !ev.Right && !ev.Left && !ev.Space && isGrounded {
		if s.idleLastStart.IsZero() {
			s.idleLastStart = time.Now()
		} else if time.Since(s.idleLastStart) > balance.PlayModeIdleTimeout {
			if !s.playerIsIdle {
				log.Debug("LoopTouchable: No keys pressed in a while, idle UI start")
			}
			s.playerIsIdle = true

			// Fade in the hint UI by stepping up the alpha value.
			if s.idleHelpAlpha < balance.PlayModeAlphaMax {
				s.idleHelpAlpha += balance.PlayModeAlphaStep
			}

			// cap it from overflow
			if s.idleHelpAlpha > balance.PlayModeAlphaMax {
				s.idleHelpAlpha = balance.PlayModeAlphaMax
			}
		}
	} else {
		s.idleLastStart = time.Time{}
		s.playerIsIdle = false
		s.idleHelpAlpha = 0
	}

	// Click (touch) event?
	if ev.Button1 {
		// Clicked left or right of middle = move left or right.
		// By default the middle box is a dead zone, but if player
		// is already moving laterally allow for quick precision.
		if ev.Left || ev.Right {
			if cursor.X < s.d.width/2 {
				ev.Left = true
				ev.Right = false
			} else if cursor.X > s.d.width/2 {
				ev.Right = true
				ev.Left = false
			}
		} else {
			if cursor.X < middle.X {
				ev.Left = true
				ev.Right = false
			} else if cursor.X > middle.X+middle.W {
				ev.Left = false
				ev.Right = true
			}
		}

		// Clicked above middle = jump.
		ev.Up = cursor.Y < middle.Y

		// Clicked below middle = down (antigravity)
		ev.Down = cursor.Y > middle.Y+middle.H

		// Clicked on the middle box = Use.
		if cursor.X >= middle.X && cursor.X <= middle.X+middle.W &&
			cursor.Y >= middle.Y && cursor.Y <= middle.Y+middle.H {
			ev.Space = true

			// Also cancel any lateral movement.
			ev.Left = false
			ev.Right = false
		}

		s.isTouching = true
	} else {
		if s.isTouching {
			ev.Left = false
			ev.Right = false
			ev.Up = false
			ev.Down = false
			ev.Space = false
			s.isTouching = false
		}
	}
}

// DrawTouchable draws any UI elements if needed for the touch UI.
func (s *PlayScene) DrawTouchable() {
	if usercfg.Current.HideTouchHints {
		return
	}

	var (
		middle     = s.touchGetMiddleBox()
		background = render.RGBA(200, 200, 200, uint8(s.idleHelpAlpha))
		font       = balance.TouchHintsFont
	)
	font.Color.Alpha = uint8(s.idleHelpAlpha)
	font.Shadow.Alpha = uint8(s.idleHelpAlpha)

	// If the player is idle for a while, start showing them a hint UI about
	// the touch screen controls.
	if s.playerIsIdle {
		// Draw the "Use" button over the middle box.
		useBtn := ui.NewLabel(ui.Label{
			Text: "Touch here\nto 'use'\nobjects",
			Font: font,
		})
		useBtn.SetBackground(background)
		useBtn.Resize(middle)
		useBtn.Compute(s.d.Engine)
		useBtn.Present(s.d.Engine, middle.Point())

		// Move Left and Move Right hints.
		moveLeft := ui.NewLabel(ui.Label{
			Text: "Touch here to\nmove left",
			Font: font,
		})
		moveLeft.SetBackground(background)
		moveLeft.Compute(s.d.Engine)
		moveLeft.Present(s.d.Engine, render.Point{
			X: (middle.X / 2) - (moveLeft.Size().W / 2),
			Y: (s.d.height / 2) - (moveLeft.Size().H / 2),
		})

		// Move Left and Move Right hints.
		moveRight := ui.NewLabel(ui.Label{
			Text: "Touch here to\nmove right",
			Font: font,
		})
		moveRight.SetBackground(background)
		moveRight.Compute(s.d.Engine)
		moveRight.Present(s.d.Engine, render.Point{
			X: (middle.X+middle.W+s.d.width)/2 - (moveRight.Size().W / 2),
			Y: (s.d.height / 2) - (moveRight.Size().H / 2),
		})

		// Jump hints.
		moveUp := ui.NewLabel(ui.Label{
			Text: "Touch anywhere above the middle of\nthe screen to jump up in the air",
			Font: font,
		})
		moveUp.SetBackground(background)
		moveUp.Compute(s.d.Engine)
		moveUp.Present(s.d.Engine, render.Point{
			X: (s.d.width / 2) - (moveUp.Size().W / 2),
			Y: (middle.Y / 2) - (moveUp.Size().H / 2),
		})

		// Keybind hints.
		keyHints := ui.NewLabel(ui.Label{
			Text: "Keyboard controls:\n" +
				"WASD or arrow keys for movement\n" +
				"Space key to 'use' objects.",
			Font: font,
		})
		keyHints.SetBackground(background)
		keyHints.Compute(s.d.Engine)
		keyHints.Present(s.d.Engine, render.Point{
			X: (s.d.width / 2) - (keyHints.Size().W / 2),
			Y: (middle.Y+middle.H+s.d.height)/2 - (keyHints.Size().H / 2),
		})
	}
}

// Get the middle box of the screen and return it.
// X,Y are screen positions and W,H is the box size.
func (s *PlayScene) touchGetMiddleBox() render.Rect {
	// Carve up the screen segments.
	var (
		// The application window dimensions.
		width  = s.d.width
		height = s.d.height

		// The middle box.
		middleMinSize = 96 // minimum dimensions
		middle        = render.Rect{
			X: (width / 2) - (middleMinSize / 2),
			Y: (height / 2) - (middleMinSize / 2),
			W: middleMinSize,
			H: middleMinSize,
		}
	)
	return middle
}
