package uix

import (
	"errors"
	"sort"

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

// GetLinkedActors returns the live Actor instances (Play Mode) which are linked
// to the live actor given.
func (w *Canvas) GetLinkedActors(a *Actor) []*Actor {
	// Identify the linked actor UUIDs from the level file.
	linkedIDs := map[string]interface{}{}
	matching := map[string]*Actor{}
	for _, id := range a.Actor.Links {
		linkedIDs[id] = nil
	}

	// Find live instances of these actors.
	for _, live := range w.actors {
		if _, ok := linkedIDs[live.ID()]; ok {
			matching[live.ID()] = live
		}
	}

	// Sort them deterministically and return.
	keys := []string{}
	for key, _ := range matching {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	result := []*Actor{}
	for _, key := range keys {
		result = append(result, matching[key])
	}
	return result
}
