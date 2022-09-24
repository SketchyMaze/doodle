/*
Package gamepad provides game controller logic for the game.

Controls

The gamepad controls are currently hard-coded for Xbox 360 style controllers, and the
controller mappings vary depending on the "mode" of control.

N Style and X Style

If the gamepad control is set to "NStyle" then the A/B and X/Y buttons will be swapped
to match the labels of a Nintendo style controller. Since the game has relatively few
inputs needed now, this module maps them to a PrimaryButton and a SecondaryButton,
defined as:

PrimaryButton: A or X button
SecondaryButton: B or Y button

Mouse Mode

- Left stick moves the mouse cursor (a cursor sprite is drawn on screen)
- Right stick scrolls the level (title screen or level editor)
- PrimaryButton emulates a left click.
- SecondaryButton emulates a right click.
- Left Shoulder emulates a middle click.
- Left Trigger (L2) closes the top-most window in the Editor (Backspace key)
- Right Shoulder toggles between Mouse Mode and other scene-specific mode.

Gameplay Mode

- Left stick moves the player character (left/right only).
- D-Pad also moves the player character (left/right only).
- PrimaryButton is to "Use"
- SecondaryButton is to "Jump"
- If the player has antigravity, up/down controls on left stick or D-Pad work too.
- Right Shoulder toggles between GameplayMode and MouseMode.
*/
package gamepad

import (
	"git.kirsle.net/SketchyMaze/doodle/pkg/balance"
	"git.kirsle.net/SketchyMaze/doodle/pkg/log"
	"git.kirsle.net/SketchyMaze/doodle/pkg/shmem"
	"git.kirsle.net/SketchyMaze/doodle/pkg/sprites"
	"git.kirsle.net/go/render"
	"git.kirsle.net/go/render/event"
	"git.kirsle.net/go/ui"
)

// Global state variables for gamepad support.
var (
	// PlayScene tells us whether antigravity is on, so directional controls work all directions.
	PlayModeAntigravity bool
	SceneName           string // Set by doodle.Goto() so we know what scene name we're on.

	playerOne *int // controller index for Player 1 (main).
	p1style   Style
	p1mode    Mode

	// Mouse cursor
	cursorVisible bool
	cursorSprite  *ui.Image
	cursorLast    render.Point // detect if mouse cursor took over from gamepad

	// MouseMode button state history to emulate mouse-ups.
	leftClickLast   bool
	rightClickLast  bool
	middleClickLast bool
	leftTriggerLast bool

	// MouseMode right stick last position, for arrow key (level scroll) emulation.
	rightStickLast event.Vector

	// Gameplay mode last left-stick position.
	leftStickLast event.Vector
	dpLeftLast    bool // D-Pad last positions.
	dpRightLast   bool
	dpDownLast    bool
	dpUpLast      bool

	// Right Shoulder last state (mode switch)
	rightShoulderLast bool
)

// SetControllerIndex sets which gamepad will be "Player One"
func SetControllerIndex(index int) {
	playerOne = &index
}

// UnsetController detaches the Player One controller.
func UnsetController() {
	playerOne = nil
}

// SetStyle sets the controller button style.
func SetStyle(s Style) {
	p1style = s
}

// SetMode sets the controller mode.
func SetMode(m Mode) {
	p1mode = m
}

// PrimaryButton returns whether the A or X button is pressed.
func PrimaryButton(ctrl event.GameController) bool {
	if p1style == NStyle {
		return ctrl.ButtonB() || ctrl.ButtonY()
	}
	return ctrl.ButtonA() || ctrl.ButtonX()
}

// SecondaryButton returns whether the B or Y button is pressed.
func SecondaryButton(ctrl event.GameController) bool {
	if p1style == NStyle {
		return ctrl.ButtonA() || ctrl.ButtonX()
	}
	return ctrl.ButtonB() || ctrl.ButtonY()
}

// Loop hooks the render events on each game tick.
func Loop(ev *event.State) {
	// If we don't have a controller registered, watch out for one until we do.
	if playerOne == nil {
		if len(ev.Controllers) > 0 {
			for idx, ctrl := range ev.Controllers {
				SetControllerIndex(idx)
				log.Info("Gamepad: using controller #%d (%d) as Player 1", idx, ctrl.Name())
				break
			}
		} else {
			return
		}
	}

	// Get our SDL2 controller.
	ctrl, ok := ev.GetController(*playerOne)
	if !ok {
		log.Error("gamepad: controller #%d has gone away! Detaching as Player 1", playerOne)
		playerOne = nil
		return
	}

	// Right Shoulder = toggle controller mode, handle this first.
	if ctrl.ButtonR1() {
		if !rightShoulderLast {
			if SceneName == "Play" {
				// Toggle between GameplayMode and MouseMode.
				if p1mode == GameplayMode {
					shmem.FlashError("Mouse Mode: Left stick moves the cursor.")
					p1mode = MouseMode
				} else {
					shmem.FlashError("Game Mode: Left stick moves the player.")
					p1mode = GameplayMode
				}

				// Reset all button states.
				ev.Left = false
				ev.Right = false
				ev.Up = false
				ev.Down = false
				ev.Enter = false
				ev.Space = false
			}

			rightShoulderLast = true
			return
		}
	} else if rightShoulderLast {
		rightShoulderLast = false
	}

	// If we are in Play Mode, translate gamepad events into key events.
	if p1mode == GameplayMode {
		// D-Pad controls to move the player character.
		if ctrl.ButtonLeft() {
			ev.Left = true
			dpLeftLast = true
		} else if dpLeftLast {
			ev.Left = false
			dpLeftLast = false
		}

		if ctrl.ButtonRight() {
			ev.Right = true
			dpRightLast = true
		} else if dpRightLast {
			ev.Right = false
			dpRightLast = false
		}

		// Antigravity on? Up/Down arrow emulation.
		if PlayModeAntigravity {
			if ctrl.ButtonUp() {
				ev.Up = true
				dpUpLast = true
			} else if dpUpLast {
				ev.Up = false
				dpUpLast = false
			}

			if ctrl.ButtonDown() {
				ev.Down = true
				dpDownLast = true
			} else if dpDownLast {
				ev.Down = false
				dpDownLast = false
			}
		}

		// "Use" button.
		if PrimaryButton(ctrl) {
			ev.Space = true
			ev.Enter = true // to click thru modals
			leftClickLast = true
		} else if leftClickLast {
			ev.Space = false
			ev.Enter = false
			leftClickLast = false
		}

		// Jump button.
		if SecondaryButton(ctrl) {
			ev.Up = true
			rightClickLast = true
		} else if rightClickLast {
			ev.Up = false
			rightClickLast = false
		}

		// Left control stick to move the player character.
		// TODO: analog movements.
		leftStick := ctrl.LeftStick()
		if leftStick.X != 0 {
			if leftStick.X < -balance.GameControllerScrollMin {
				ev.Left = true
				ev.Right = false
			} else if leftStick.X > balance.GameControllerScrollMin {
				ev.Right = true
				ev.Left = false
			} else {
				ev.Right = false
				ev.Left = false
			}
		} else if leftStickLast.X != 0 {
			ev.Left = false
			ev.Right = false
		}

		// Antigravity on?
		if PlayModeAntigravity {
			if leftStick.Y != 0 {
				if leftStick.Y < -balance.GameControllerScrollMin {
					ev.Up = true
					ev.Down = false
				} else if leftStick.Y > balance.GameControllerScrollMin {
					ev.Down = true
					ev.Up = false
				} else {
					ev.Down = false
					ev.Up = false
				}
			} else if leftStickLast.Y != 0 {
				ev.Up = false
				ev.Down = false
			}
		}

		leftStickLast = leftStick
	}

	// If we are emulating a mouse, handle that now.
	if p1mode == MouseMode {
		// Move the cursor.
		leftStick := ctrl.LeftStick()
		if leftStick.X != 0 || leftStick.Y != 0 {
			cursorVisible = true
		} else if cursorVisible {
			// If the mouse cursor has moved behind our back (e.g., real mouse moved), turn off
			// the MouseMode cursor sprite.
			if cursorLast.X != ev.CursorX || cursorLast.Y != ev.CursorY {
				cursorVisible = false
			}
		}

		ev.CursorX += int(leftStick.X * balance.GameControllerMouseMoveMax)
		ev.CursorY += int(leftStick.Y * balance.GameControllerMouseMoveMax)

		// Constrain the cursor inside window boundaries.
		w, h := shmem.CurrentRenderEngine.WindowSize()
		if ev.CursorX < 0 {
			ev.CursorX = 0
		} else if ev.CursorX > w {
			ev.CursorX = w
		}
		if ev.CursorY < 0 {
			ev.CursorY = 0
		} else if ev.CursorY > h {
			ev.CursorY = h
		}

		// Store last cursor point so we can detect mouse movement outside the gamepad.
		cursorLast = render.NewPoint(ev.CursorX, ev.CursorY)

		// Are we clicking?
		if PrimaryButton(ctrl) {
			ev.Button1 = true // left-click
			leftClickLast = true
		} else if leftClickLast {
			ev.Button1 = false
			leftClickLast = false
		}

		// Right-click
		if SecondaryButton(ctrl) {
			ev.Button3 = true // right-click
			rightClickLast = true
		} else if rightClickLast {
			ev.Button3 = false
			rightClickLast = false
		}

		// Middle-click
		if ctrl.ButtonL1() {
			ev.Button2 = true // middle click
			middleClickLast = true
		} else if middleClickLast {
			ev.Button2 = false
			middleClickLast = false
		}

		// Left Trigger = Backspace (close active window)
		if ctrl.ButtonL2() {
			if !leftTriggerLast {
				ev.SetKeyDown(`\b`, true)
			}
			leftTriggerLast = true
		} else if leftTriggerLast {
			ev.SetKeyDown(`\b`, false)
			leftTriggerLast = false
		}

		// Arrow Key emulation on the right control stick, e.g. for Level Editor.
		rightStick := ctrl.RightStick()
		if rightStick.X != 0 {
			if rightStick.X < -balance.GameControllerScrollMin {
				ev.Left = true
				ev.Right = false
			} else if rightStick.X > balance.GameControllerScrollMin {
				ev.Right = true
				ev.Left = false
			} else {
				ev.Right = false
				ev.Left = false
			}
		} else if rightStickLast.X != 0 {
			ev.Left = false
			ev.Right = false
		}

		if rightStick.Y != 0 {
			if rightStick.Y < -balance.GameControllerScrollMin {
				ev.Up = true
				ev.Down = false
			} else if rightStick.Y > balance.GameControllerScrollMin {
				ev.Down = true
				ev.Up = false
			} else {
				ev.Down = false
				ev.Up = false
			}
		} else if rightStickLast.Y != 0 {
			ev.Up = false
			ev.Down = false
		}

		rightStickLast = rightStick
	}
}

// Draw the cursor on screen if the game controller is emulating a mouse.
func Draw(e render.Engine) {
	if playerOne == nil || p1mode != MouseMode || !cursorVisible {
		return
	}

	if cursorSprite == nil {
		img, err := sprites.LoadImage(e, balance.CursorIcon)
		if err != nil {
			log.Error("gamepad: couldn't load cursor sprite (%s): %s", balance.CursorIcon, err)
			return
		}
		cursorSprite = img
	}

	cursorSprite.Present(e, shmem.Cursor)
}
