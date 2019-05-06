# Doodad Scripts

Each Doodad can have a JavaScript file attached to give them some logic and code
to run in-game when a level is being played. Buttons, trap doors, and other
dynamic doodads have JavaScript code that tells them how to behave.

This game uses [otto](https://github.com/robertkrimen/otto) for its JavaScript
engine, so it only works with ES5 syntax and has a few weird quirks. You get
used to them.

## Example

Provide a script file with a `main` function:

```javascript
function main() {
  console.log("%s initialized!", Self.Doodad.Title);

  var timer = 0;
  Events.OnCollide( function() {
    if (timer > 0) {
      clearTimeout(timer);
    }

    Self.ShowLayer(1);
    timer = setTimeout(function() {
      Self.ShowLayer(0);
      timer = 0;
    }, 200);
  })
}
```

# JavaScript API

# Functions

Global functions available to your script:

## RGBA(uint8 red, uint8 green, uint8 blue, uint8 alpha)

Get a render.Color object from the given color code. Each number
must be between 0 and 255. For example, RGBA(255, 0, 255, 255)
creates an opaque magenta color equivalent to #FF00FF.

The render.Color type may be needed in certain API calls that
require the game's native color type.

## Point(x int, y int)

Get a render.Point object that refers to a position in the game
world.

## Common JavaScript Functions

The following common JavaScript APIs seen in web browsers work
in the doodad scripts:

* int setTimeout(function, int milliseconds)

  Set a timeout to run your function after a delay in
  milliseconds. Returns a timer ID to be used with
  clearTimeout() if you want to cancel the timeout.

* int setInterval(function, int milliseconds)

  Like setTimeout, but repeatedly re-runs the function after
  the delay in milliseconds. Returns a timer ID to be used
  with clearInterval() if you want to cancel the interval.

* clearTimeout(int timerID), clearInterval(int timerID)

  Cancel a timeout or interval by passing its timer ID, which
  was returned when the timer or interval was first created.

* console.log(str message, v...)

  Write to the game's log console. There are also `console.warn`,
  `console.error` and `console.log` variants.

## Self

The global variable `Self` holds an API for the current doodad. The full
surface area of this API is subject to change, but some useful examples you
can do with this are as follows.

### Self.Doodad

Self.Doodad is a pointer into the doodad's metadata file. Not
all properties in there can be written to or read from the
JavaScript engine, but some useful attributes are:

* `str Self.Doodad.Title`: the title of the doodad.
* `str Self.Doodad.Author`: the author name of the doodad.
* `str Self.Doodad.Script`: your own source code. Note that
  editing this won't have any effect in-game, as your doodad's
  source has already been loaded into the interpreter.
* `str Self.Doodad.GameVersion`: the game version that created
  the doodad.

### Self.ShowLayer(int index)

Set the doodad's visible layer to the index.

A layer is a drawing created in the in-game format. Only one layer
is visible on-screen during Edit Mode or Play Mode. Layers can be
used to store alternate versions of your doodad to show different
states or as animation frames.

The first and default layer is always zero. Use `CountLayers()`
to query how many layers are in the doodad.

### int Self.CountLayers()

Returns the number of layers, or frames, available in your doodad's
drawing data. Usually these layers are for alternate drawings or animation frames.

The number is the `len()` of the array, and layers are
zero-indexed, so the first and default layer is always layer 0
and the final layer is like so:

```javascript
// set the final frame as the active one
Self.ShowLayer( Self.CountLayers() - 1 );
```

## Animations

### Self.AddAnimation(string name, int interval, [layers...]) error

Add a new animation using some of the layers in your doodad's drawing.

The interval is counted in milliseconds, with 1000 meaning one second between
frames of the animation.

The layers are an array of strings or integers. If strings, use the layer names
from the drawing. With integers, these are the layer index numbers where 0 is
the first (default) layer.

### Self.PlayAnimation(string name, function callback)

Play the named animation. When the animation is finished, the callback function
will be called. Set the callback to `null` if you don't want a callback function.

### Self.StopAnimation()

Stop and cancel any current animations. Their callback functions will not be
called.

### Self.IsAnimating() bool

Returns `true` if an animation is currently playing.

## Events

Use the Events object to register event handlers from your
doodad. Usually you'll configure these in your main() function.

Example configuring event handlers:

```javascript
function main() {
  Events.OnCollide(function(e) {
    console.log("I've been collided with!");
  })
}
```

### OnCollide

Triggers when another doodad has collided with your doodad's box
space on the level. Arguments TBD.

### OnEnter

Triggers when another doodad has fully intersected your doodad's
box.

### OnLeave

Triggers when a doodad who was intersecting your box has left
your box.

### KeypressEvent

Triggers when the player character has pressed a key.

This only triggers when your doodad is the focus of the camera
in-game, i.e. for the player character doodad.
