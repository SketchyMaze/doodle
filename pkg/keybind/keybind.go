// Package keybind centralizes the global hotkey bindings.
//
// Whenever the app would need to query a hotkey like "F3" or "Ctrl-Z"
// is held down, it should use a method in this file. It can be
// expanded later to allow user customizable bindings or something.
//
// NOTE: arrow key and gameplay controls not yet ported to here.
package keybind

import "git.kirsle.net/go/render/event"

// Shutdown (Escape) signals the game to start closing down.
func Shutdown(ev *event.State) bool {
	return ev.Escape
}

// Help (F1) can be checked one time.
func Help(ev *event.State) bool {
	result := ev.KeyDown("F1")
	ev.SetKeyDown("F1", false)
	return result
}

// DebugOverlay (F3) can be checked one time.
func DebugOverlay(ev *event.State) bool {
	result := ev.KeyDown("F3")
	ev.SetKeyDown("F3", false)
	return result
}

// DebugCollision (F4) can be checked one time.
func DebugCollision(ev *event.State) bool {
	result := ev.KeyDown("F4")
	ev.SetKeyDown("F4", false)
	return result
}

// Undo (Ctrl-Z)
func Undo(ev *event.State) bool {
	return ev.Ctrl && ev.KeyDown("z")
}

// Redo (Ctrl-Y)
func Redo(ev *event.State) bool {
	return ev.Ctrl && ev.KeyDown("y")
}

// ZoomIn (+)
func ZoomIn(ev *event.State) bool {
	return ev.KeyDown("=") || ev.KeyDown("+")
}

// ZoomOut (-)
func ZoomOut(ev *event.State) bool {
	return ev.KeyDown("-")
}

// ZoomReset (1)
func ZoomReset(ev *event.State) bool {
	return ev.KeyDown("1")
}

// Origin (0) -- scrolls the canvas back to 0,0 in Editor Mode.
func Origin(ev *event.State) bool {
	return ev.KeyDown("0")
}

// GotoPlay (P) play tests the current level in the editor.
func GotoPlay(ev *event.State) bool {
	return ev.KeyDown("p")
}

// GotoEdit (E) opens the current played level in Edit Mode, if the
// player has come from the editor originally.
func GotoEdit(ev *event.State) bool {
	return ev.KeyDown("e")
}

// LineTool (L) selects the Line Tool in the editor.
func LineTool(ev *event.State) bool {
	return ev.KeyDown("l")
}

// PencilTool (F) selects the freehand pencil tool in the editor.
// GotoPlay (P) play tests the current level in the editor.
func PencilTool(ev *event.State) bool {
	return ev.KeyDown("f")
}

// RectTool (R) selects the rectangle in the editor.
func RectTool(ev *event.State) bool {
	return ev.KeyDown("r")
}

// EllipseTool (C) selects this tool in the editor.
func EllipseTool(ev *event.State) bool {
	return ev.KeyDown("c")
}

// EraserTool (X) selects this tool in the editor.
func EraserTool(ev *event.State) bool {
	return ev.KeyDown("x")
}

// DoodadDropper (D) opens the doodad dropper in the editor.
func DoodadDropper(ev *event.State) bool {
	return ev.KeyDown("d")
}

// ShellKey (`) opens the developer console.
func ShellKey(ev *event.State) bool {
	v := ev.KeyDown("`")
	ev.SetKeyDown("`", false)
	return v
}

// Enter key.
func Enter(ev *event.State) bool {
	v := ev.Enter
	ev.Enter = false
	return v
}

// Shift key.
func Shift(ev *event.State) bool {
	return ev.Shift
}

// Left arrow.
func Left(ev *event.State) bool {
	return ev.Left || ev.KeyDown("a")
}

// Right arrow.
func Right(ev *event.State) bool {
	return ev.Right || ev.KeyDown("d")
}

// Up arrow.
func Up(ev *event.State) bool {
	return ev.Up || ev.KeyDown("w")
}

// Down arrow.
func Down(ev *event.State) bool {
	return ev.Down || ev.KeyDown("s")
}

// "Use" button.
func Use(ev *event.State) bool {
	return ev.Space || ev.KeyDown("q")
}
