# TODO

## Alpha Launch Minimum Checklist

- [ ] Open Source Licenses
- [x] Doodad Scripts: an "end level" function for a level goalpost.

**Blocker Bugs:**

- [ ] Sometimes the red Azulians don't interact with other doodads
  properly, but sometimes they do. (i.e. they phase thru doors, don't
  interact with buttons or keys).

**UI Cleanup:**

- Doodads Palette:
  - [x] Hide some doodads like the player character.
  - [x] Pagination or scrolling UI for long lists of doodads.

**Nice to haves:**

## Release Launch Checklist

**Features:**

- [ ] Single-player "campaign mode" of built-in levels.
  - campaign.json file format configuring the level order
- [ ] Level Editor Improvements
  - [x] Undo/Redo Function
  - [x] Lines and Boxes
  - [ ] Eraser Tool
  - [ ] Brush size and/or shape
- [ ] Doodad CLI Tool Features
  - [x] `doodad show` to display information about a level or doodad.
  - [ ] `doodad init` or some such to generate a default JS script.
  - [x] Options to toggle various states (hidden, hasInventory?)

**Shareware Version:**

- [x] Can't draw or edit doodads.
- [ ] Can only create Bounded maps, not infinite ones.
- [ ] Can play custom maps but only ones using built-in doodads.
- [ ] Can not place custom doodads in maps.

**Built-in Doodads:**

- [x] Buttons
  - [x] Press Button
  - [x] Sticky Button
- [x] Switches (4 varieties)
- [x] Doors
  - [x] Locked Doors and Keys
  - [x] Electric Doors
  - [x] Trapdoors (all 4 directions)

## Doodad Ideas

In addition to those listed above:

- [ ] Crumbly floor: Tomb Raider inspired cracked stone floor that
  crumbles under the player a moment after being touched.
- [ ] Firepit: decorative, painful
- [ ] Gravity Boots: flip player's gravity upside down.
- [ ] Warp Doors that lead to other linked maps.
  - For campaign levels only. If used in a normal player level, acts
    as a level goal and ends the level.
  - Doodads "Warp Door A" through "Warp Door D"
  - The campaign.json would link levels together.
- [ ] Conveyor Belt

## New Ideas

- [ ] New Doodad struct fields:
  - [ ] `Hidden bool`: skip showing this doodad in the palette UI.
  - [ ] `HasInventory bool`: for player characters and maybe thieves. This way
    keys only get picked up by player characters and not "any doodad that
    touches them"
  - [ ] ``

## Path to Multiplayer

* Add a Player abstraction between events and player characters.
  * Keyboard keys would update PlayerOne's state with actions (move left, right, jump, etc)
  * Possible to have multiple local players (i.e. bound to different keyboard keys, bound to joypads, etc.)
* A NetworkPlayer provides a Player's inputs from over a network.
* Client/server negotiation, protocol
  * Client can request chunks from server for local rendering.
  * Players send inputs over network sockets.
