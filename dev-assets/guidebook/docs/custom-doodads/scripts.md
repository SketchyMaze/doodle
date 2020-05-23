# Doodad Scripts

Doodads are programmed using JavaScript which gives them their behavior
and ability to interact with the player and other doodads.

Doodad scripts are run during "Play Mode" when a level _containing_ the doodad
is being played. You can install a JavaScript (.js) file into a doodad using
the command-line `doodad` program.

An example Doodad script looks like the following:

```javascript
// main() is called on level initialization for each
// instance ("actor") of the doodad.
function main() {
    // Logs go to the game's log file (standard output on Linux/Mac).
    console.log("%s initialized!", Self.Title);

    // If our doodad has 'solid' parts that should prohibit movement,
    // define the hitbox here. Coordinates are relative so 0,0 is the
    // top-left pixel of the doodad's sprite.
    Self.SetHitbox(0, 0, 64, 12);

    // Handle a collision when another doodad (or player) has entered
    // the space of our doodad.
    Events.OnCollide(function(e) {
        // The `e` object holds information about the event.
        console.log("Actor %s has entered our hitbox!", e.Actor.ID());

        // InHitbox is `true` if we defined a hitbox for ourselves, and
        // the colliding actor is inside of the hitbox we defined.
        if (e.InHitbox) {
            // To prohibit movement, return false from the OnCollide handler.
            // If you don't return false, the actor is allowed to keep on
            // moving through.
            return false;
        }

        // When movement is finalized, OnCollide is called one final time
        // with e.Settled=true; it is only then that a doodad should run
        // event handlers for a logical collide event.
        if (e.Settled) {
            // do something
            Message.Publish("power", true);
        }
    });

    // OnLeave is called when an actor, who was previously colliding with
    // us, is no longer doing so.
    Events.OnLeave(function(e) {
        console.log("Actor %s has stopped colliding!", e.Actor.ID());
    })
}
```

# Installing a Doodad Script

Use the command-line `doodad` tool to attach a script to your doodad file:

```bash
# Attach the JavaScript at "script.js" to the doodad file "filename.doodad"
doodad install-script script.js filename.doodad

# To view the script currently attached to a doodad
# (prints the script to your terminal)
doodad show --script filename.doodad
```

# Testing Your Script

The best way to test your doodad script is to use it in a level!

Run the game in a console to watch the log output, and you can use functions
like `console.log()` in your script to help debug issues. Drag your custom
doodad into a level and playtest it! Your script's main() function is called
when the level instance of your doodad is initialized.

# JavaScript API

The following global variables are available to all Doodad scripts.

## Self

Self holds data about the current doodad instance loaded inside of a level.

**String attributes:**

* Self.Title: the doodad title.
* Self.Filename: the doodad filename (useful for inventory items).

Methods are below.

### Self.ID() string

Returns the "actor ID" of the doodad instance loaded inside of a level. This
is usually a random UUID string that was saved with the level data.

### Self.GetTag(string name) string

Return a "tag" that was saved with the doodad's file data.

Tags are an arbitrary key/value data store attached to the doodad file.
You can use the `doodad.exe` tool shipped with the game to view and manage tags
on your own custom doodads:

```bash
# Command-line doodad tool usage:

# Show information about a doodad, look for the "Tags:" section.
doodad show filename.doodad

# Set a tag. "-t" for short.
doodad edit-doodad --tag 'color=blue' filename.doodad

# Set the tag to empty to remove it.
doodad edit-doodad -t 'color=' filename.doodad
```

This is useful for a set of multiple doodads to share the same script but
have different behavior depending on how each is tagged.

### Self.Position() Point

Returns the doodad's current position in the level.

Point is an object with .X and .Y integer values.

```javascript
var p = Self.Position()
console.log("I am at %d,%d", p.X, p.Y)
```

### Self.SetHitbox(x, y, w, h int)

Configure the "solid hitbox" of this doodad.

The X and Y coordinates are relative to the doodad's sprite: (0,0) is the top
left pixel of the doodad. The W and H are the width and height of the hitbox
starting at those coordinates.

When another doodad enters the area of your doodad's sprite (for example, the
player character has entered the square shape of your doodad sprite) your script
begins to receive OnCollide events from the approaching actor.

The OnCollide event tells you if the invading doodad is inside your custom
hitbox which you define here (`InHitbox`) making it easy to make choices based
on that status.

Here's an example script for a hypothetical "locked door" doodad that acts
solid but only on a thin rectangle in the middle of its sprite:

```javascript
// Example script for a "locked door"
function main() {
    // Suppose the doodad's sprite size is 64x64 pixels square.
    // The door is in side profile where the door itself ranges from pixels
    //    (20, 0) to (24, 64)
    Self.SetHitbox(20, 0, 24, 64)

    // OnCollide handlers.
    Events.OnCollide(function(e) {
        // The convenient e.InHitbox tells you if the colliding actor is
        // inside the hitbox we defined.
        if (e.InHitbox) {
            // Return false to protest the collision (act solid).
            return false;
        }
    });
}
```

### Self.SetVelocity(Velocity)

Set the doodad's velocity. Velocity is a type that can be created with the
Velocity() constructor, which takes an X and Y value:

```javascript
Self.SetVelocity( Velocity(3.2, 7.0) );
```

A positive X velocity propels the doodad to the right. A positive Y velocity
propels the doodad downward.

### Self.SetMobile(bool)

Call `SetMobile(true)` if the doodad will move on its own.

This is for mobile doodads such as the player character and enemy mobs.
Stationary doodads like buttons, doors, and trapdoors do not mark themselves
as mobile.

Mobile doodads incur extra work for the game doing collision checking so only
set this to `true` if your doodad will move (i.e. changes its Velocity or
Position).

```javascript
Self.SetMobile(true);
```

### Self.SetGravity(bool)

Set whether gravity applies to this doodad. By default doodads are stationary
and do not fall downwards. The player character and some mobile enemies that
want to be affected by gravity should opt in to this.

```javascript
Self.SetGravity(true);
```

### Self.ShowLayer(index int)

Switch the active layer of the doodad to the layer at this index.

A doodad file can contain multiple layers, or images. The first and default
layer is at index zero, the second layer at index 1, and so on.

```javascript
Self.ShowLayer(0);  // 0 is the first and default layer
Self.ShowLayer(1);  // show the second layer instead
```

### Self.ShowLayerNamed(name string)

Switch the active layer by name instead of index.

Each layer has an arbitrary name that it can be addressed by instead of needing
to keep track of the layer index.

Doodads created by the command-line `doodad` tool will have their layers named
automatically by their file name. The layer **indexes** will retain the same
order of file names passed in, with 0 being the first file:

```bash
# Doodad tool-created doodads have layers named after their file names.
# example "open-1.png" will be named "open-1"
doodad convert door.png open-1.png open-2.png open-3.png my-door.doodad
```

### Self.AddAnimation(name string, interval int, layers list)

Register a named animation for your doodad. `interval` is the time in
milliseconds before going to the next frame. `layers` is an array of layer
names or indexes to be used for the animation.

Doodads can animate by having multiple frames (images) in the same file.
Layers are ordered (layer 0 is the first, then increments from there) and
each has a name. This function can take either identifier to specify
which layers are part of the animation.

```javascript
// Animation named "open" using named layers, 100ms delay between frames.
Self.AddAnimation("open", 100, ["open-1", "open-2", "open-3"]);

// Animation named "close" using layers by index.
Self.AddAnimation("close", 100, [3, 2, 1]);
```

### Self.PlayAnimation(name string, callback func())

This starts playing the named animation. The callback function will be called
when the animation has completed.

```javascript
Self.PlayAnimation("open", function() {
    console.log("I've finished opening!");

    // The callback is optional; use null if you don't need it.
    Self.PlayAnimation("close", null);
});
```

### Self.IsAnimating() bool

Returns true if an animation is currently being played.

### Self.StopAnimation()

Stops any currently playing animation.


* Self.Doodad(): a pointer to the doodad's file data.
  * Self.Doodad().Title: get the title of the doodad file.
  * Self.Doodad().Author: the name of the author who wrote the doodad.
  * Self.Doodad().Script: the doodad's JavaScript source code. Note that
    modifying this won't have any effect in-game, as the script had already
    been loaded into the interpreter.
  * Self.Doodad().GameVersion: the version of {{ app_name }} that was used
    when the doodad was created.

### Self.Destroy()

This destroys the current instance of the doodad as it appears in a level.

For example, a Key destroys itself when it's picked up so that it disappears
from the level and can't be picked up again. Call this function when the
doodad instance should be destroyed and removed from the active level.

-----

## Console Logging

Like in node.js and the web browser, `console.log` and friends are available
for logging from a doodad script. Logs are emitted to the same place as the
game's logs are.

```javascript
console.log("Hello world!");
console.log("Interpolate strings '%s' and numbers '%d'", "string", 123);
console.debug("Debug messages shown when the game is in debug mode");
console.warn("Warning-level messages");
console.error("Error-level messages");
```

-----

## Timers and Intervals

Like in a web browser, functions such as setTimeout and setInterval are
supported in doodad scripts.

### setTimeout(function, milliseconds int) int

setTimeout calls your function after the specified number of milliseconds.

1000ms are in one second.

Returns an integer "timeout ID" that you'll need if you want to cancel the
timeout with clearTimeout.

### setInterval(function, milliseconds int) int

setInterval calls your function repeatedly after every specified number of
milliseconds.

Returns an integer "interval ID" that you'll need if you want to cancel the
interval with clearInterval.

### clearTimeout(id int)

Cancels the timeout with the given ID.

### clearInterval(id int)

Cancels the interval with the given ID.

-----

## Type Constructors

Some methods may need data of certain native types that aren't available in
JavaScript. These global functions will initialize data of the correct types:

### RGBA(red, green, blue, alpha uint8)

Creates a Color type from red, green, blue and alpha values (integers between
0 and 255).

### Point(x, y int)

Creates a Point object with X and Y coordinates.

### Vector(x, y float64)

Creates a Vector object with X and Y dimensions.

-----

## Global Functions

Some useful globally available functions:

### EndLevel()

This ends the current level, i.e. to be used by the goal flag.

### Flash(message string, args...)

Flash a message on screen to the user.

Flashed messages appear at the bottom of the screen and fade out after a few
moments. If multiple messages are flashed at the same time, they stack from the
bottom of the window with the newest message on bottom.

Don't abuse this feature as spamming it may annoy the player.

### GetTick() uint64

Returns the current game tick. This value started at zero when the game was
launched and increments every frame while running.

### time.Now() time.Time

This exposes the Go standard library function `time.Now()` that returns the
current date and time as a Go time.Time value.

### time.Add(t time.Time, milliseconds int64) time.Time

Add a number of milliseconds to a Go Time value.

--------

## Event Handlers

Doodad scripts can respond to certain events using functions on the global
`Events` variable.

### Events.OnCollide( func(event) )

OnCollide is called when another actor is colliding with your doodad's sprite
box. The function is given a CollideEvent object which has the following
attributes:

* Actor: the doodad which is colliding with your doodad.
* Overlap (Rect): a rectangle of where the two doodads' boxes are overlapping,
  relative to your doodad sprite's box. That is, if the Actor was moving in from
  the left side of your doodad, the X value would be zero and W would be the
  number of pixels of overlap.
* InHitbox (bool): true if the colliding actor's hitbox is intersecting with
  the hitbox you defined with SetHitbox().
* Settled (bool): This is `false` when the game is trying to move the colliding
  doodad and is sussing out whether or not your doodad will act solid and
  protest its movement. When the game has settled the location of the colliding
  doodad it will call OnCollide a final time with Settled=true. If your doodad
  has special behavior when touched (i.e. a button that presses in), you should
  wait until Settled=true before running your handler for that.

### Events.OnLeave( func(event) )

Called when an actor that _was_ colliding with your doodad is no longer
colliding (or has left your doodad's sprite box).

### Events.RunKeypress( func(event) )

Handle a keypress. `event` is an `event.State` from the render engine.

TODO: document that.

-----

## Pub/Sub Communication

Doodads in a level are able to send and receive messages to other doodads,
either those that they are **linked** to or those that listen on a more
'broadcast' frequency.

> **Linking** is when the level author connected two doodads together with
> the Link Tool. The two doodads' scripts can communicate with each other
> in-game over that link.

For example, if the level author links a Button to an Electric Door, the button
can send a "power" event to the door so that it can open when a player touches
the button.

Doodads communicate in a "publisher/subscriber" model: one doodad publishes an
event with a name and data, and other doodads subscribe to the named event to
receive that data.

### Official, Standard Pub/Sub Messages

The following message names and data types are used by the game's default
doodads. You're free to use these in your own custom doodads.

If extending this list with your own custom events, be careful to choose a
unique namespace to prevent collision with other users' custom doodads and
their custom event names.

| Name | Data Type | Description |
|------|-----------|--------------|
| power | boolean | Communicates a "powered" (true) or "not powered" state, as in a Button to an Electric Door. |
| broadcast:state-change | boolean | An "ON/OFF" button was hit and all state blocks should flip. |

### Message.Publish(name string, data...)

Publish a named message to all of your **linked** doodads.

`data` is a list of arbitrary arguments to send with the message.

```javascript
// Example button doodad that emits a "power" (bool) state to linked doodads
// that subscribe to this event.
function main() {
    // When an actor collides with the button, emit a powered state.
    Events.OnCollide(function(e) {
        Message.Publish("power", true);
    });

    // When the actor leaves the button, turn off the power.
    Events.OnLeave(function(e) {
        Message.Publish("power", false);
    })
}
```

### Message.Subscribe(name string, function)

Subscribe to a named message from any **linked** doodads.

The function receives the data that was passed with the message. Its data type
is arbitrary and will depend on the type of message.

```javascript
// Example electronic device doodad that responds to power from linked buttons.
function main() {
    // Boolean to store if our electric device has juice.
    var powered = false;

    // Do something while powered
    setInterval(function() {
        if (powered) {
            console.log("Brmm...")
        }
    }, 1000);

    // Subscribe to the `power` event by a linked button or other power source.
    Message.Subscribe("power", function(boolValue) {
        console.log("Powered %s!", boolValue === true ? "on" : "off");
        powered = boolValue;
    });
}
```

### Message.Broadcast(name string, data...)

This publishes a named message to **every** doodad in the level, whether it
was linked to the broadcaster or not.

For example the "ON/OFF" button globally toggles a boolean state in every
state block that subscribes to the `broadcast:state-change` event.

If you were to broadcast an event like `power` it would activate every single
power-sensitive doodad on the level.

```javascript
// Example two-state block that globally receives the state-change broadcast.
function main() {
    var myState = false;
    Message.Subscribe("broadcast:state-change", function(boolValue) {
      // Most two-state blocks just flip their own state, ignoring the
      // boolValue passed with this message.
      myState = !myState;
    });
}

// Example ON/OFF block that emits the state-change broadcast. It also
// subscribes to the event to keep its own state in sync with all the other
// ON/OFF blocks on the level when they get hit.
function main() {
    var myState = false;

    // Listen for other ON/OFF button activations to keep our state in
    // sync with theirs.
    Message.Subscribe("broadcast:state-change", function(boolValue) {
        myState = boolValue;
    });

    // When collided with, broadcast the state toggle to all state blocks.
    Events.OnCollide(function(e) {
        if (e.Settled) {
            myState = !!myState;
            Message.Broadcast("broadcast:state-change", myState);
        }
    })
}
```
