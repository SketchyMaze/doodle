package uix

// Functions relating to the Doodad JavaScript API for Canvas Actors.

// MakeSelfAPI generates the `Self` object for the scripting API in
// reference to a live Canvas actor in the level.
func (w *Canvas) MakeSelfAPI(actor *Actor) map[string]interface{} {
	return map[string]interface{}{
		"Filename": actor.Doodad().Filename,
		"Title":    actor.Doodad().Title,

		// functions
		"ID":             actor.ID,
		"GetTag":         actor.Doodad().Tag,
		"Position":       actor.Position,
		"SetHitbox":      actor.SetHitbox,
		"SetVelocity":    actor.SetVelocity,
		"SetMobile":      actor.SetMobile,
		"SetGravity":     actor.SetGravity,
		"AddAnimation":   actor.AddAnimation,
		"IsAnimating":    actor.IsAnimating,
		"PlayAnimation":  actor.PlayAnimation,
		"StopAnimation":  actor.StopAnimation,
		"ShowLayer":      actor.ShowLayer,
		"ShowLayerNamed": actor.ShowLayerNamed,
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
