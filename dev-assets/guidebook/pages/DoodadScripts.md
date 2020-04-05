# Doodad Scripts

Doodads are programmed using JavaScript which gives them their behavior
and ability to interact with the player and other doodads.

An example Doodad script looks like the following:

```javascript
// The main function is called when the doodad is initialized in Play Mode
// at the start of the level.
function main() {
    // Important global variables:
    // - Self: information about the current Doodad running this script.
    // - Events: handle events raised during gameplay.
    // - Message: publish or subscribe to named messages to interact with
    //   other doodads.

    // Logs go to the game's log file (standard output on Linux/Mac).
    console.log("%s initialized!", Self.Doodad().Title);

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
    });

    // OnLeave is called when an actor, who was previously colliding with
    // us, is no longer doing so.
    Events.OnLeave(function(e) {
        console.log("Actor %s has stopped colliding!", e.Actor.ID());
    })
}
```

# JavaScript API

## Global Variables

The following global variables are available to all Doodad scripts.

### Self

Self holds information about the current doodad. The full surface area of
the Self object is subject to change, but some useful things you can access
from it include:

* Self.Doodad(): a pointer to the doodad's file data.
  * Self.Doodad().Title: get the title of the doodad file.
  * Self.Doodad().Author: the name of the author who wrote the doodad.
  * Self.Doodad().Script: the doodad's JavaScript source code. Note that
    modifying this won't have any effect in-game, as the script had already
    been loaded into the interpreter.
  * Self.Doodad().GameVersion: the version of {{ app_name }} that was used
    when the doodad was created.

### Events

### Message

## Global Functions

The following useful functions are also available globally:

### Timers and Intervals

Doodad scripts implement setTimeout() and setInterval() functions similar
to those found in web browsers.

```javascript
// Call a function after 5 seconds.
setTimeout(function() {
    console.log("I've been called!");
}, 5000);
```

setTimeout() and setInterval() return an ID number for the timer created.
If you wish to cancel a timer before it has finished, or to stop an interval
from running, you need to pass its ID number into `clearTimeout()` or
`clearInterval()`, respectively.

```javascript
// Start a 1-second interval
var id = setInterval(function() {
    console.log("Tick...");
}, 1000);

// Cancel it after 30 seconds.
setTimeout(function() {
    clearInterval(id);
}, 30000);
```

### Console Logging

Doodad scripts also implement the `console.log()` and similar functions as
found in web browser APIs. They support "printf" style variable placeholders.

```javascript
console.log("Hello world!");
console.error("The answer is %d!", 42);
console.warn("Actor '%s' has collided with us!", e.Actor.ID());
console.debug("This only logs when the game is in debug mode!");
```

### RGBA(red, green, blue, alpha uint8)

RGBA initializes a Color variable using the game's native Color type. May
be useful for certain game APIs that take color values.

Example: RGBA(255, 0, 255, 255) creates an opaque magenta color.

### Point(x, y int)

Returns a Point object which refers to a location in the game world. This
type is required for certain game APIs.
