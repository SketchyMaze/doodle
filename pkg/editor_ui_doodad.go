package doodle

// XXX REFACTOR XXX
// This function only uses EditorUI and not Doodle and is a candidate for
// refactor into a subpackage if EditorUI itself can ever be decoupled.

import (
	"git.kirsle.net/apps/doodle/pkg/balance"
	"git.kirsle.net/apps/doodle/pkg/doodads"
	"git.kirsle.net/apps/doodle/pkg/level"
	"git.kirsle.net/apps/doodle/pkg/log"
	"git.kirsle.net/apps/doodle/pkg/uix"
	"git.kirsle.net/go/render"
)

// DraggableActor is a Doodad being dragged from the Doodad palette.
type DraggableActor struct {
	canvas *uix.Canvas
	doodad *doodads.Doodad // if a new one from the palette
	actor  *level.Actor    // if a level actor
}

// startDragActor begins the drag event for a Doodad onto a level.
// actor may be nil (if you drag a new doodad from the palette) or otherwise
// is an existing actor from the level.
func (u *EditorUI) startDragActor(doodad *doodads.Doodad, actor *level.Actor) {
	u.Supervisor.DragStart()

	if doodad == nil {
		if actor != nil {
			obj, err := doodads.LoadFile(actor.Filename)
			if err != nil {
				log.Error("startDragExistingActor: actor doodad name %s not found: %s", actor.Filename, err)
				return
			}
			doodad = obj
		} else {
			panic("EditorUI.startDragActor: doodad AND/OR actor is required, but neither were given")
		}
	}

	// Create the canvas to render on the mouse cursor.
	drawing := uix.NewCanvas(doodad.Layers[0].Chunker.Size, false)
	drawing.LoadDoodad(doodad)
	drawing.Resize(doodad.Rect())
	drawing.SetBackground(render.RGBA(0, 0, 1, 0)) // TODO: invisible becomes white
	drawing.MaskColor = balance.DragColor          // blueprint effect
	u.DraggableActor = &DraggableActor{
		canvas: drawing,
		doodad: doodad,
		actor:  actor,
	}
}
