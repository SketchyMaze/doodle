package uix

import (
	"errors"
	"fmt"

	"git.kirsle.net/apps/doodle/lib/render"
	"git.kirsle.net/apps/doodle/pkg/doodads"
	"git.kirsle.net/apps/doodle/pkg/level"
	"github.com/google/uuid"
	"github.com/robertkrimen/otto"
)

// Actor is an object that marries together the three things that make a
// Doodad instance "tick" while inside a Canvas:
//
// - uix.Actor is a doodads.Drawing so it fulfills doodads.Actor to be a
//   dynamic object during gameplay.
// - It has a pointer to the level.Actor indicating its static level data
//   as defined in the map: its spawn coordinate and configuration.
// - A uix.Canvas that can present the actor's graphics to the screen.
type Actor struct {
	doodads.Drawing
	Actor  *level.Actor
	Canvas *Canvas

	activeLayer int  // active drawing frame for display
	flagDestroy bool // flag the actor for destruction

	// Actor runtime variables.
	hasGravity bool
	hitbox     render.Rect
	data       map[string]string

	// Animation variables.
	animations        map[string]*Animation
	activeAnimation   *Animation
	animationCallback otto.Value
}

// NewActor sets up a uix.Actor.
// If the id is blank, a new UUIDv4 is generated.
func NewActor(id string, levelActor *level.Actor, doodad *doodads.Doodad) *Actor {
	if id == "" {
		id = uuid.Must(uuid.NewRandom()).String()
	}

	size := int32(doodad.Layers[0].Chunker.Size)
	can := NewCanvas(int(size), false)
	can.Name = id

	// TODO: if the Background is render.Invisible it gets defaulted to
	// White somewhere and the Doodad masks the level drawing behind it.
	can.SetBackground(render.RGBA(0, 0, 1, 0))

	can.LoadDoodad(doodad)
	can.Resize(render.NewRect(size, size))

	actor := &Actor{
		Drawing:    doodads.NewDrawing(id, doodad),
		Actor:      levelActor,
		Canvas:     can,
		animations: map[string]*Animation{},
	}

	// Give the Canvas a pointer to its (parent) Actor so it can draw its debug
	// label and show the World Position of the actor within the world.
	can.actor = actor

	return actor
}

// SetGravity configures whether the actor is affected by gravity.
func (a *Actor) SetGravity(v bool) {
	a.hasGravity = v
}

// GetBoundingRect gets the bounding box of the actor's doodad.
func (a *Actor) GetBoundingRect() render.Rect {
	return doodads.GetBoundingRect(a)
}

// SetHitbox sets the actor's elected hitbox.
func (a *Actor) SetHitbox(x, y, w, h int) {
	a.hitbox = render.Rect{
		X: int32(x),
		Y: int32(y),
		W: int32(w),
		H: int32(h),
	}
}

// Hitbox returns the actor's elected hitbox.
func (a *Actor) Hitbox() render.Rect {
	return a.hitbox
}

// SetData sets an arbitrary field in the actor's K/V storage.
func (a *Actor) SetData(key, value string) {
	if a.data == nil {
		a.data = map[string]string{}
	}
	a.data[key] = value
}

// GetData gets an arbitrary field from the actor's K/V storage.
// Missing keys just return a blank string (friendly to the JavaScript
// environment).
func (a *Actor) GetData(key string) string {
	if a.data == nil {
		return ""
	}

	v, _ := a.data[key]
	return v
}

// LayerCount returns the number of layers in this actor's drawing.
func (a *Actor) LayerCount() int {
	return len(a.Doodad.Layers)
}

// ShowLayer sets the actor's ActiveLayer to the index given.
func (a *Actor) ShowLayer(index int) error {
	if index < 0 {
		return errors.New("layer index must be 0 or greater")
	} else if index > len(a.Doodad.Layers) {
		return fmt.Errorf("layer %d out of range for doodad's layers", index)
	}

	a.activeLayer = index
	a.Canvas.Load(a.Doodad.Palette, a.Doodad.Layers[index].Chunker)
	return nil
}

// Destroy deletes the actor from the running level.
func (a *Actor) Destroy() {
	a.flagDestroy = true
}
