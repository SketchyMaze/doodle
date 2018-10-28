# Doodad Scripting Engine

Some ideas for the scripting engine for Doodads inside your level.

# Architecture

The script will be an "attached file" in the Doodad format as a special file
named "index.js" as the entry point.

Each Doodad will have its `index.js` script loaded into an isolated JS
environment where it can't access any data about other Doodads or anything
user specific. The `main()` function is called so the Doodad script can
set itself up.

The `main()` function should:

* Initialize any state variables the Doodad wants to use in its script.
* Subscribe to callback events that the Doodad is interested in catching.

The script interacts with the Doodle application through an API broker object
(a Go surface area of functions).

# API Broker Interface

```go
type API interface {
    // "Self" functions.
    SetFrame(frame int)  // Set the currently visible frame in this Doodad.
    MoveTo(render.Point)

    // Game functions.k
    EndLevel()  // Exit the current level with a victory

    /************************************
     * Event Handler Callback Functions *
     ************************************/

    // When we become visible on screen or disappear off the screen.
    OnVisible()
    OnHidden()

    // OnEnter: the other Doodad has ENTIRELY entered our box. Or if the other
    // doodad is bigger, they have ENTIRELY enveloped ours.
    OnEnter(func(other Doodad))

    // OnCollide: when we bump into another Doodad.
    OnCollide(func(other Doodad))
}
```

## Mockup Script

```javascript
function main() {
  console.log("hello world");

  // Register event callbacks.
  Doodle.OnEnter(onEnter);
}

// onEnter: handle when another Doodad (like the player) completely enters
// the bounding box of our Doodad. Example: a level exit.
function onEnter(other) {

}
```
