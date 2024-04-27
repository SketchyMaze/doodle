# Tour of the Code

Here is a brief tour of where to find important parts of the game's code.

## Basic App Structure

**cmd/doodle/main.go** is the main function for Sketchy Maze and handles the command line flags and bootstrapping the game.

**pkg/doodle.go** holds the root Doodle (`d`) struct, the base object for the meat of the program - its setup funcs and MainLoop are here.

### Scenes

The gameplay is divided into scenes, each in a **pkg/\*_scene.go** file:

* MainScene is the title screen (default)
* PlayScene is play mode
* EditScene is the level editor, etc.

**pkg/scene.go** defines the interface for Scene itself and the `d.Goto(scene)` function.

The game changes scenes with Goto and the MainLoop calls the active scene's functions (Loop and Draw).

## Doodad JavaScript API

### pkg/scripting

This package provides the generic JavaScript API used by the developer console and doodads.

**pkg/scripting/js_api.go** defines global and generic JavaScript functions to doodads' scope, e.g.:

* console.log and friends
* Sound.Play
* RGBA, Point, and Vector
* Flash, GetTick
* time.Now() and friends
* Events
* setTimeout, setInterval..
* Self: provided externally.

### pkg/uix/scripting.go

The home of **MakeScriptAPI()** which installs various Play Mode objects into the script's environment, e.g.

* Actors and Level API
* Self: a surface of uix.Actor capabilities that point to the local actor.

## Collision Detection

### pkg/collision

The algorithmic stuff is here - the collision package focused on just the math and general types to handle logic for:

* Collision with level geometry ('pixel perfect')
* Collision between actors
* BoundingRect and collision boxes

### pkg/uix/actor_collision.go