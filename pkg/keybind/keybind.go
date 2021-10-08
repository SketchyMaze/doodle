// Package keybind centralizes the global hotkey bindings.
//
// Whenever the app would need to query a hotkey like "F3" or "Ctrl-Z"
// is held down, it should use a method in this file. It can be
// expanded later to allow user customizable bindings or something.
//
// NOTE: arrow key and gameplay controls not yet ported to here.
package keybind

import "git.kirsle.net/go/render/event"

// State returns a version of event.State which is domain specific
// to what the game actually cares about.
type State struct {
	State          *event.State
	Shutdown       bool // Escape key
	Help           bool // F1
	DebugOverlay   bool // F3
	DebugCollision bool // F4
	Undo           bool // Ctrl-Z
	Redo           bool // Ctrl-Y
	NewLevel       bool // Ctrl-N
	Save           bool // Ctrl-S
	SaveAs         bool // Shift-Ctrl-S
	Open           bool // Ctrl-O
	ZoomIn         bool // +
	ZoomOut        bool // -
	ZoomReset      bool // 1
	Origin         bool // 0
	GotoPlay       bool // p
	GotoEdit       bool // e
	PencilTool     bool
	LineTool       bool
	RectTool       bool
	EllipseTool    bool
	EraserTool     bool
	DoodadDropper  bool
	ShellKey       bool
	Enter          bool
	Left           bool
	Right          bool
	Up             bool
	Down           bool
	Use            bool
}

// FromEvent converts a render.Event readout of the current keys
// being pressed but formats them in the way the game uses them.
// For example, WASD and arrow keys both move the player and the
// game only cares which direction.
func FromEvent(ev *event.State) State {
	return State{
		State:          ev,
		Shutdown:       Shutdown(ev),
		Help:           Help(ev),
		DebugOverlay:   DebugOverlay(ev),
		DebugCollision: DebugCollision(ev), // F4
		Undo:           Undo(ev),           // Ctrl-Z
		Redo:           Redo(ev),           // Ctrl-Y
		NewLevel:       NewLevel(ev),       // Ctrl-N
		Save:           Save(ev),           // Ctrl-S
		SaveAs:         SaveAs(ev),         // Shift-Ctrl-S
		Open:           Open(ev),           // Ctrl-O
		ZoomIn:         ZoomIn(ev),         // +
		ZoomOut:        ZoomOut(ev),        // -
		ZoomReset:      ZoomReset(ev),      // 1
		Origin:         Origin(ev),         // 0
		GotoPlay:       GotoPlay(ev),       // p
		GotoEdit:       GotoEdit(ev),       // e
		PencilTool:     PencilTool(ev),
		LineTool:       LineTool(ev),
		RectTool:       RectTool(ev),
		EllipseTool:    EllipseTool(ev),
		EraserTool:     EraserTool(ev),
		DoodadDropper:  DoodadDropper(ev),
		ShellKey:       ShellKey(ev),
		Enter:          Enter(ev),
		Left:           Left(ev),
		Right:          Right(ev),
		Up:             Up(ev),
		Down:           Down(ev),
		Use:            Use(ev),
	}
}

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

// CloseTopmostWindow (Backspace)
func CloseTopmostWindow(ev *event.State) bool {
	result := ev.KeyDown(`\b`)
	ev.SetKeyDown(`\b`, false)
	return result
}

// CloseAllWindows (Shift+Backspace)
func CloseAllWindows(ev *event.State) bool {
	result := ev.KeyDown(`\b`) && ev.Shift
	if result {
		ev.SetKeyDown(`\b`, false)
	}
	return result
}

// NewViewport (V)
func NewViewport(ev *event.State) bool {
	result := ev.KeyDown("v")
	if result {
		ev.SetKeyDown("v", false)
	}
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

// New Level (Ctrl-N)
func NewLevel(ev *event.State) bool {
	return ev.Ctrl && ev.KeyDown("n")
}

// Save (Ctrl-S)
func Save(ev *event.State) bool {
	var result = ev.Ctrl && ev.KeyDown("s")
	if result {
		ev.SetKeyDown("s", false)
	}
	return result
}

// SaveAs (Shift-Ctrl-S)
func SaveAs(ev *event.State) bool {
	return ev.Ctrl && ev.Shift && ev.KeyDown("s")
}

// Open (Ctrl-O)
func Open(ev *event.State) bool {
	return ev.Ctrl && ev.KeyDown("o")
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
	return ev.KeyDown("q")
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

// MiddleClick of the mouse for panning the level.
func MiddleClick(ev *event.State) bool {
	return ev.Button2
}
