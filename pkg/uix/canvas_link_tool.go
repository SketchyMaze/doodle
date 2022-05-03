package uix

import (
	"errors"
	"sort"

	"git.kirsle.net/apps/doodle/pkg/log"
	"git.kirsle.net/apps/doodle/pkg/shmem"
)

// LinkAdd adds an actor to be linked in the Link tool.
func (w *Canvas) LinkAdd(a *Actor) error {
	if w.linkFirst == nil {
		// First click, hold onto this actor.
		w.linkFirst = a
		shmem.Flash("Doodad '%s' selected, click the next Doodad to link it to",
			a.Doodad().Title,
		)
	} else if w.linkFirst == a {
		// Clicked the same doodad twice, deselect it.
		shmem.Flash("De-selected the doodad for linking.")
		w.linkFirst = nil
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

// PrintLinks is a debug function that describes all actor links to console.
func (w *Canvas) PrintLinks() {
	log.Error("### BEGIN Canvas.PrintLinks() ###")

	// Map all of the actors by their IDs so we can look them up from their links
	var actors = map[string]*Actor{}
	for _, actor := range w.actors {
		actors[actor.ID()] = actor
	}

	for _, actor := range w.actors {
		var id = actor.ID()
		if len(id) > 8 {
			id = id[:8]
		}

		if len(actor.Actor.Links) > 0 {
			log.Info("Actor %s (%s) at %s has %d links:", id, actor.Actor.Filename, actor.Position(), len(actor.Actor.Links))
			for _, link := range actor.Actor.Links {
				if otherActor, ok := actors[link]; ok {
					var linkId = link
					if len(linkId) > 8 {
						linkId = linkId[:8]
					}

					log.Info("\tTo %s (%s) at %s", linkId, otherActor.Actor.Filename, otherActor.Position())
				} else {
					log.Error("\tTo unknown actor ID %s!", link)
				}
			}
		}
	}

	log.Error("### END Canvas.PrintLinks() ###")
}
