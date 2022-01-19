package uix

import (
	"git.kirsle.net/apps/doodle/pkg/doodads"
	"git.kirsle.net/apps/doodle/pkg/level"
	"git.kirsle.net/apps/doodle/pkg/log"
	"git.kirsle.net/apps/doodle/pkg/scripting"
	"git.kirsle.net/go/render"
)

// Functions relating to the Doodad JavaScript API for Canvas Actors.

// MakeScriptAPI makes several useful globals available to doodad
// scripts such as Actors.At()
func (w *Canvas) MakeScriptAPI(vm *scripting.VM) {
	vm.Set("Actors", map[string]interface{}{
		// Actors.At(Point)
		"At": func(p render.Point) []*Actor {
			var result = []*Actor{}

			for _, actor := range w.actors {
				var box = actor.Hitbox().AddPoint(actor.Position())
				if actor != nil && p.Inside(box) {
					result = append(result, actor)
				}
			}

			return result
		},

		// Actors.FindPlayer: returns the nearest player character.
		"FindPlayer": func() *Actor {
			for _, actor := range w.actors {
				if actor.IsPlayer() {
					return actor
				}
			}
			return nil
		},

		// Actors.New: create a new actor.
		"New": func(filename string) *Actor {
			doodad, err := doodads.LoadFile(filename)
			if err != nil {
				panic(err)
			}
			actor := NewActor("_new", &level.Actor{}, doodad)
			w.AddActor(actor)

			// Set up the player character's script in the VM.
			if err := w.scripting.AddLevelScript(actor.ID(), filename); err != nil {
				log.Error("Actors.New(%s): scripting.InstallActor(player) failed: %s", filename, err)
			}

			return actor
		},

		// Actors.SetPlayerCharacter: remake the player character.
		"SetPlayerCharacter": func(filename string) {
			if w.OnSetPlayerCharacter != nil {
				w.OnSetPlayerCharacter(filename)
			} else {
				log.Error("Actors.SetPlayerCharacter: caller was not ready")
			}
		},
	})
}

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
		"Grounded":       actor.Grounded,
		"SetHitbox":      actor.SetHitbox,
		"Hitbox":         actor.Hitbox,
		"SetVelocity":    actor.SetVelocity,
		"GetVelocity":    actor.Velocity,
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
		"Freeze":         actor.Freeze,
		"Unfreeze":       actor.Unfreeze,
		"Hide":           actor.Hide,
		"Show":           actor.Show,
		"GetLinks": func() []map[string]interface{} {
			var result = []map[string]interface{}{}
			for _, linked := range w.GetLinkedActors(actor) {
				result = append(result, w.MakeSelfAPI(linked))
			}
			return result
		},
	}
}
