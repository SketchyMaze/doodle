package uix

import (
	"errors"
	"fmt"
	"sort"
	"sync"

	"git.kirsle.net/SketchyMaze/doodle/pkg/collision"
	"git.kirsle.net/SketchyMaze/doodle/pkg/doodads"
	"git.kirsle.net/SketchyMaze/doodle/pkg/level"
	"git.kirsle.net/SketchyMaze/doodle/pkg/log"
	"git.kirsle.net/SketchyMaze/doodle/pkg/physics"
	"git.kirsle.net/go/render"
	"github.com/dop251/goja"
	"github.com/google/uuid"
)

// Actor is an object that marries together the three things that make a
// Doodad instance "tick" while inside a Canvas:
//
//   - uix.Actor is a doodads.Drawing so it fulfills doodads.Actor to be a
//     dynamic object during gameplay.
//   - It has a pointer to the level.Actor indicating its static level data
//     as defined in the map: its spawn coordinate and configuration.
//   - A uix.Canvas that can present the actor's graphics to the screen.
type Actor struct {
	Drawing *doodads.Drawing
	Actor   *level.Actor
	Canvas  *Canvas

	activeLayer int  // active drawing frame for display
	flagDestroy bool // flag the actor for destruction
	flagUsing   bool // flag that the (player) has pressed the Use key.

	// Actor runtime variables.
	hasGravity   bool
	hasInventory bool
	wet          bool
	isMobile     bool // Mobile character, such as the player or an enemy
	noclip       bool // Disable collision detection
	hidden       bool // invisible, via Hide() and Show()
	frozen       bool // Frozen, via Freeze() and Unfreeze()
	immortal     bool // Invulnerable to damage
	hitbox       render.Rect
	inventory    map[string]int // item inventory. doodad name -> quantity, 0 for key item.

	// Movement data.
	position render.Point
	velocity physics.Vector
	grounded bool

	// Animation variables.
	animations        map[string]*Animation
	activeAnimation   *Animation
	animationCallback goja.Value

	// Mutex.
	muInventory sync.RWMutex
	muData      sync.RWMutex
}

// NewActor sets up a uix.Actor.
// If the id is blank, a new UUIDv4 is generated.
func NewActor(id string, levelActor *level.Actor, doodad *doodads.Doodad) *Actor {
	if id == "" {
		id = uuid.Must(uuid.NewUUID()).String()
	}

	size := doodad.ChunkSize()
	can := NewCanvas(uint8(size), false)
	can.Name = id

	// TODO: if the Background is render.Invisible it gets defaulted to
	// White somewhere and the Doodad masks the level drawing behind it.
	can.SetBackground(render.RGBA(0, 0, 1, 0))

	can.LoadDoodad(doodad)
	can.Resize(doodad.Size)

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

// ID returns the actor's ID. This is the underlying doodle.Drawing.ID().
func (a *Actor) ID() string {
	return a.Drawing.ID()
}

// Doodad offers access to the underlying Doodad object.
// Shortcut to the `.Drawing.Doodad` property path.
func (a *Actor) Doodad() *doodads.Doodad {
	return a.Drawing.Doodad
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

// SetInventory configures whether the actor is capable of carrying items.
func (a *Actor) SetInventory(v bool) {
	a.hasInventory = true
}

// IsMobile returns whether the actor is a mobile character.
func (a *Actor) IsMobile() bool {
	return a.isMobile
}

// IsPlayer returns whether the actor is the player character.
// It's true when the Actor ID is "PLAYER"
func (a *Actor) IsPlayer() bool {
	return a.Canvas.Name == "PLAYER"
}

// HasInventory returns if the actor is capable of carrying items.
func (a *Actor) HasInventory() bool {
	return a.hasInventory
}

// HasGravity returns if gravity applies to the actor.
func (a *Actor) HasGravity() bool {
	return a.hasGravity
}

// Invulnerable returns whether the actor is marked as immortal.
func (a *Actor) Invulnerable() bool {
	return a.immortal
}

// SetInvulnerable sets the actor's immortal flag.
func (a *Actor) SetInvulnerable(v bool) {
	a.immortal = v
}

// Wet returns whether the actor is in contact with water pixels in a level.
func (a *Actor) IsWet() bool {
	return a.wet
}

// SetWet updates the state of the actor's wet-ness.
func (a *Actor) SetWet(v bool) {
	a.wet = v
}

// Size returns the size of the actor, from the underlying doodads.Drawing.
func (a *Actor) Size() render.Rect {
	return a.Drawing.Doodad.Size
}

// Velocity returns the actor's current velocity vector.
func (a *Actor) Velocity() physics.Vector {
	return a.velocity
}

// SetVelocity updates the actor's velocity vector.
func (a *Actor) SetVelocity(v physics.Vector) {
	a.velocity = v
}

// Position returns the actor's position.
func (a *Actor) Position() render.Point {
	return a.position
}

// MoveTo sets the actor's position.
func (a *Actor) MoveTo(p render.Point) {
	a.position = p
}

// MoveBy adjusts the actor's position.
func (a *Actor) MoveBy(p render.Point) {
	a.position.Add(p)
}

// Grounded returns if the actor is touching a floor.
func (a *Actor) Grounded() bool {
	return a.grounded
}

// SetGrounded sets the actor's grounded value. If true, also sets their Y velocity to zero.
func (a *Actor) SetGrounded(v bool) {
	a.grounded = v
	// if v && a.velocity.Y > 0 {
	// 	a.velocity.Y = 0
	// }
}

// Hide makes the actor invisible.
func (a *Actor) Hide() {
	a.hidden = true
}

// Show a hidden actor.
func (a *Actor) Show() {
	a.hidden = false
}

// Freeze an actor. For the player character, this means arrow key inputs
// will stop moving the actor.
func (a *Actor) Freeze() {
	a.frozen = true
}

// Unfreeze an actor.
func (a *Actor) Unfreeze() {
	a.frozen = false
}

// IsFrozen returns true if the actor is frozen.
func (a *Actor) IsFrozen() bool {
	return a.frozen
}

// SetUsing enables the "Use Key" flag, mainly for the player character to activate
// certain doodads in the level.
func (a *Actor) SetUsing(v bool) {
	a.flagUsing = v
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
	if _, ok := a.inventory[itemName]; ok {
		a.inventory[itemName] += quantity
	} else {
		a.inventory[itemName] = quantity
	}
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

// ClearInventory removes all items from the actor's inventory.
func (a *Actor) ClearInventory() {
	a.muInventory.Lock()
	a.inventory = map[string]int{}
	a.muInventory.Unlock()
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
	return collision.GetBoundingRect(a)
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

// Hitbox returns the actor's elected hitbox. If the JavaScript did not set
// a hitbox, it defers to the Doodad's metadata hitbox.
func (a *Actor) Hitbox() render.Rect {
	if a.hitbox.IsZero() && !a.Drawing.Doodad.Hitbox.IsZero() {
		return a.Drawing.Doodad.Hitbox
	}
	return a.hitbox
}

// Options returns the list of all available Doodad options, sorted.
func (a *Actor) Options() []string {
	var result = []string{}
	for option := range a.Doodad().Options {
		result = append(result, option)
	}
	sort.Strings(result)
	return result
}

// Get an option value from the actor. If the option is not configured,
// returns the default Doodad option, or nil if not there either.
func (a *Actor) GetOption(name string) *level.Option {
	// Actor configured option?
	if opt, ok := a.Actor.Options[name]; ok {
		return opt
	}

	// Doodad default option?
	if opt, ok := a.Doodad().Options[name]; ok {
		return &level.Option{
			Name:  opt.Name,
			Type:  opt.Type,
			Value: opt.Default,
		}
	}

	return nil
}

// LayerCount returns the number of layers in this actor's drawing.
func (a *Actor) LayerCount() int {
	return len(a.Doodad().Layers)
}

// ShowLayer sets the actor's ActiveLayer to the index given.
func (a *Actor) ShowLayer(index int) error {
	if index < 0 {
		return errors.New("layer index must be 0 or greater")
	} else if index > len(a.Doodad().Layers) {
		return fmt.Errorf("layer %d out of range for doodad's layers", index)
	}

	a.activeLayer = index
	a.Canvas.Load(a.Doodad().Palette, a.Doodad().Layers[index].Chunker)
	return nil
}

// ShowLayerNamed sets the actor's ActiveLayer to the one named.
func (a *Actor) ShowLayerNamed(name string) error {
	// Find the layer.
	for i, layer := range a.Doodad().Layers {
		if layer.Name == name {
			return a.ShowLayer(i)
		}
	}
	log.Warn("Actor(%s) ShowLayerNamed(%s): layer not found",
		a.Actor.Filename,
		name,
	)

	// XX: returning an error raises a JavaScript exception in doodads. :/ Warning log is enough.
	return nil
}

// Destroy deletes the actor from the running level.
func (a *Actor) Destroy() {
	a.flagDestroy = true
}
