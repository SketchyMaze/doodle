package uix

import "git.kirsle.net/go/render"

// Functions relating to the Doodad JavaScript API for Canvas Actors.

// MakeSelfAPI generates the `Self` object for the scripting API in
// reference to a live Canvas actor in the level.
func (w *Canvas) MakeSelfAPI(actor *Actor) map[string]interface{} {
	return map[string]interface{}{
		"Filename": actor.Doodad().Filename,
		"Title":    actor.Doodad().Title,

		// functions
		"ID": actor.ID,
		"Size": func() render.Rect {
			var size = actor.Doodad().ChunkSize()
			return render.NewRect(size, size)
		},
		"GetTag":   actor.Doodad().Tag,
		"Position": actor.Position,
		"MoveTo": func(p render.Point) {
			actor.MoveTo(p)
			actor.SetGrounded(false)
		},
		"SetHitbox":      actor.SetHitbox,
		"SetVelocity":    actor.SetVelocity,
		"SetMobile":      actor.SetMobile,
		"SetInventory":   actor.SetInventory,
		"HasInventory":   actor.HasInventory,
		"SetGravity":     actor.SetGravity,
		"AddAnimation":   actor.AddAnimation,
		"IsAnimating":    actor.IsAnimating,
		"IsPlayer":       actor.IsPlayer,
		"PlayAnimation":  actor.PlayAnimation,
		"StopAnimation":  actor.StopAnimation,
		"ShowLayer":      actor.ShowLayer,
		"ShowLayerNamed": actor.ShowLayerNamed,
		"Inventory":      actor.Inventory,
		"AddItem":        actor.AddItem,
		"RemoveItem":     actor.RemoveItem,
		"HasItem":        actor.HasItem,
		"ClearInventory": actor.ClearInventory,
		"Destroy":        actor.Destroy,
		"GetLinks": func() []map[string]interface{} {
			var result = []map[string]interface{}{}
			for _, linked := range w.GetLinkedActors(actor) {
				result = append(result, w.MakeSelfAPI(linked))
			}
			return result
		},
	}
}
