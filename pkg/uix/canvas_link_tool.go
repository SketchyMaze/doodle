package uix

import (
	"errors"

	"git.kirsle.net/apps/doodle/pkg/drawtool"
	"git.kirsle.net/apps/doodle/pkg/shmem"
)

// LinkStart initializes the Link tool.
func (w *Canvas) LinkStart() {
	w.Tool = drawtool.LinkTool
	w.linkFirst = nil
}

// LinkAdd adds an actor to be linked in the Link tool.
func (w *Canvas) LinkAdd(a *Actor) error {
	if w.linkFirst == nil {
		// First click, hold onto this actor.
		w.linkFirst = a
		shmem.Flash("Doodad '%s' selected, click the next Doodad to link it to",
			a.Doodad().Title,
		)
	} else {
		// Second click, call the OnLinkActors handler with the two actors.
		if w.OnLinkActors != nil {
			w.OnLinkActors(w.linkFirst, a)
		} else {
			return errors.New("Canvas.LinkAdd: no OnLinkActors handler is ready")
		}

		// Reset the link state.
		w.linkFirst = nil
	}
	return nil
}
