package uix

import (
	"errors"
	"fmt"
	"sort"
	"sync"

	"git.kirsle.net/apps/doodle/pkg/doodads"
	"git.kirsle.net/apps/doodle/pkg/level"
	"git.kirsle.net/apps/doodle/pkg/log"
	"git.kirsle.net/go/render"
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
	isMobile   bool // Mobile character, such as the player or an enemy
	noclip     bool // Disable collision detection
	hitbox     render.Rect
	inventory  map[string]int    // item inventory. doodad name -> quantity, 0 for key item.
	data       map[string]string // arbitrary key/value store. DEPRECATED ??

	// Animation variables.
	animations        map[string]*Animation
	activeAnimation   *Animation
	animationCallback otto.Value

	// Mutex.
	muInventory sync.RWMutex
	muData      sync.RWMutex
}

// NewActor sets up a uix.Actor.
// If the id is blank, a new UUIDv4 is generated.
func NewActor(id string, levelActor *level.Actor, doodad *doodads.Doodad) *Actor {
	if id == "" {
		id = uuid.Must(uuid.NewRandom()).String()
	}

	size := doodad.Layers[0].Chunker.Size
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
		inventory:  map[string]int{},
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

// SetMobile configures whether the actor is a mobile character (i.e. is the
// player or a mobile enemy). Mobile characters can set off certain traps when
// touched but non-mobile actors don't set each other off if touching.
func (a *Actor) SetMobile(v bool) {
	a.isMobile = v
}

// IsMobile returns whether the actor is a mobile character.
func (a *Actor) IsMobile() bool {
	return a.isMobile
}

// SetNoclip sets the noclip setting for an actor. If true, the actor can
// clip through level geometry.
func (a *Actor) SetNoclip(v bool) {
	a.noclip = v
}

// AddItem adds an item doodad to the actor's inventory.
// Item name is usually the doodad filename.
func (a *Actor) AddItem(itemName string, quantity int) {
	a.muInventory.Lock()
	a.inventory[itemName] = quantity
	a.muInventory.Unlock()
}

// RemoveItem removes a quantity of an item from the actor's inventory.
//
// Provide a quantity of 0 to remove the item completely.
// Otherwise provides a number greater than zero and you will subtract this
// quantity from the item. If the item then is at <= zero, it is removed from
// inventory.
func (a *Actor) RemoveItem(itemName string, quantity int) bool {
	a.muInventory.RLock()
	defer a.muInventory.RUnlock()

	if _, ok := a.inventory[itemName]; ok {
		// If quantity is zero, remove the item entirely.
		if quantity <= 0 {
			delete(a.inventory, itemName)
		} else {
			// Subtract the quantity from inventory. If we have run down to
			// zero left, remove the item entirely.
			a.inventory[itemName] -= quantity
			if a.inventory[itemName] <= 0 {
				delete(a.inventory, itemName)
			}
		}
		return true
	}
	return false
}

// HasItem checks the actor's inventory for the item and returns the quantity.
//
// A return value of -1 means the item was not found.
// The value 0 indicates a key item (one with no quantity).
// Values >= 1 would be consumable items.
func (a *Actor) HasItem(itemName string) int {
	a.muInventory.RLock()
	defer a.muInventory.RUnlock()

	if quantity, ok := a.inventory[itemName]; ok {
		return quantity
	}
	return -1
}

// ListItems returns a sorted list of the items in the actor's inventory.
func (a *Actor) ListItems() []string {
	a.muInventory.RLock()
	defer a.muInventory.RUnlock()

	var (
		result = make([]string, len(a.inventory))
		i      = 0
	)
	for k := range a.inventory {
		result[i] = k
		i++
	}

	sort.Strings(result)
	return result
}

// Inventory returns a copy of the actor's inventory struct.
func (a *Actor) Inventory() map[string]int {
	a.muInventory.RLock()
	defer a.muInventory.RUnlock()

	var result = map[string]int{}
	for k, v := range a.inventory {
		result[k] = v
	}

	return result
}

// GetBoundingRect gets the bounding box of the actor's doodad.
func (a *Actor) GetBoundingRect() render.Rect {
	return doodads.GetBoundingRect(a)
}

// SetHitbox sets the actor's elected hitbox.
func (a *Actor) SetHitbox(x, y, w, h int) {
	a.hitbox = render.Rect{
		X: x,
		Y: y,
		W: w,
		H: h,
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

	a.muData.Lock()
	a.data[key] = value
	a.muData.Unlock()
}

// GetData gets an arbitrary field from the actor's K/V storage.
// Missing keys just return a blank string (friendly to the JavaScript
// environment).
func (a *Actor) GetData(key string) string {
	if a.data == nil {
		return ""
	}

	a.muData.RLock()
	v, _ := a.data[key]
	a.muData.RUnlock()

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

// ShowLayerNamed sets the actor's ActiveLayer to the one named.
func (a *Actor) ShowLayerNamed(name string) error {
	// Find the layer.
	for i, layer := range a.Doodad.Layers {
		if layer.Name == name {
			return a.ShowLayer(i)
		}
	}
	log.Warn("Actor(%s) ShowLayerNamed(%s): layer not found",
		a.Actor.Filename,
		name,
	)
	return fmt.Errorf("the layer named %s was not found", name)
}

// Destroy deletes the actor from the running level.
func (a *Actor) Destroy() {
	a.flagDestroy = true
}
