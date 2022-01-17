package uix

import (
	"errors"
	"fmt"
	"reflect"
	"time"

	"git.kirsle.net/apps/doodle/pkg/log"
	"github.com/dop251/goja"
)

// Animation holds a named animation for a doodad script.
type Animation struct {
	Name     string
	Interval time.Duration
	Layers   []int

	// runtime state variables
	activeLayer int
	nextFrameAt time.Time
}

/*
TickAnimation advances an animation forward.

This method is called by canvas.Loop() only when the actor is currently
`animating` and their current animation's nextFrameAt has been reached by the
current time.Now().

Returns true when the animation has finished and false if there is still more
frames left to animate.
*/
func (a *Actor) TickAnimation(an *Animation) bool {
	an.activeLayer++
	if an.activeLayer < len(an.Layers) {
		a.ShowLayer(an.Layers[an.activeLayer])
	} else if an.activeLayer >= len(an.Layers) {
		// final layer has been shown for 2 ticks, return that the animation has
		// been concluded.
		return true
	}

	// Schedule the next frame of animation.
	an.nextFrameAt = time.Now().Add(an.Interval)

	return false
}

// AddAnimation installs a new animation into the scripting engine for this actor.
//
// The layers can be an array of string names or integer indexes.
func (a *Actor) AddAnimation(name string, interval int64, layers []interface{}) error {
	if len(layers) == 0 {
		return errors.New("no named layers given to AddAnimation()")
	}

	// Find all the layers by name.
	var indexes []int
	for _, name := range layers {
		switch v := name.(type) {
		case string:
			var found bool
			for i, layer := range a.Doodad().Layers {
				if layer.Name == v {
					indexes = append(indexes, i)
					found = true
					break
				}
			}

			if !found {
				return fmt.Errorf("layer named '%s' not found in doodad", v)
			}
		case int64:
			// TODO: I want to find out if this is ever not an int64 coming from
			// JavaScript.
			if reflect.TypeOf(v).String() != "int64" {
				log.Error("AddAnimation: expected an int64 from JavaScript but got a %s", reflect.TypeOf(v))
			}

			iv := int(v)
			if iv < len(a.Doodad().Layers) {
				indexes = append(indexes, iv)
			} else {
				return fmt.Errorf("layer numbered '%d' is out of bounds", iv)
			}
		default:
			return fmt.Errorf(
				"invalid type for layer '%+v': should be a string (named layer) "+
					"or int (indexed layer) but was a %s", v, reflect.TypeOf(name))
		}
	}

	a.animations[name] = &Animation{
		Name:     name,
		Interval: time.Duration(interval) * time.Millisecond,
		Layers:   indexes,
	}

	return nil
}

// PlayAnimation starts an animation and then calls a JavaScript function when
// the last frame has played out. Set a null function to ignore the callback.
func (a *Actor) PlayAnimation(name string, callback goja.Value) error {
	anim, ok := a.animations[name]
	if !ok {
		return fmt.Errorf("animation named '%s' not found", name)
	}

	a.activeAnimation = anim
	a.animationCallback = callback

	// Show the first layer.
	anim.activeLayer = 0
	anim.nextFrameAt = time.Now().Add(anim.Interval)
	a.ShowLayer(anim.Layers[0])

	return nil
}

// IsAnimating returns if the current actor is playing an animation.
func (a *Actor) IsAnimating() bool {
	return a.activeAnimation != nil
}

// StopAnimation stops any current animations.
func (a *Actor) StopAnimation() {
	if a.activeAnimation == nil {
		return
	}

	a.activeAnimation = nil
	a.animationCallback = goja.Null()
}
