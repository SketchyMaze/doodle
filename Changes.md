# Changes

## v0.13.2 (TBD)

This release brings some new features and optimization for the game's file
formats to improve performance and memory usage.

Some new features:

* **Doodads can be non-square!** You can now set a rectangular canvas size
  for your doodads. Many of the game's built-in doodads that used to be
  off-center before (doors, creatures) because their sprites were not squares
  now have correct rectangular shapes.
* A **Cheats Menu** has been added which enables you to enter many of the
  game's cheat codes by clicking on buttons instead. Enable it through the
  "Experimental" tab of the Settings, and the cheats menu can be opened from
  the Help menu bar during gameplay.
* The game now supports **Signed Levels** and levelpacks which will enable the
  free version of the game to play levels that have embedded doodads in them.
  The idea is that the base game in the future may come with levelpacks that
  use custom doodads (to tell a specific story) and these doodads may be so
  specific and niche that they will ride with the levelpack instead of be
  built-in to the game. To allow the free version of the game to play these
  levels, the levelpacks will be signed. It also makes it possible for
  promotional levelpacks or "free DLC" to be shipped separately and allow
  free versions of the game to play them with their attached custom doodads.

Other miscellaneous changes:

* The default Author name on your new drawings will prefer to use your
  license registration name (if the game is registered) before falling back
  on your operating system's $USER name like before.
* In the level editor, you can now use the Pan Tool to access the actor
  properties of doodads you've dropped into your level. Similar to the
  Actor Tool, when you mouse-over an actor on your level it will highlight
  in a grey box and a gear icon in the corner can be clicked to access
  its properties. Making the properties available for the Pan Tool can
  help with touchscreen devices, where it is difficult to touch the
  properties button without accidentally dragging the actor elsewhere
  on your level as might happen with the Actor Tool!
* Fix a bug wherein the gold "perfect run" icon next to the level timer
  would sometimes not appear, especially after you had been cheating
  before - if you restart the level with no cheats active the gold icon
  should now always appear.
* New cheat code: `tesla` will send a power signal to ALL actors on the
  current level in play mode - opening all electric doors and trapdoors.
  May cause fun chaos during gameplay. Probably not very useful.
* Start distributing AppImage releases for GNU/Linux (64-bit and 32-bit)

Some technical changes related to file format optimization:

* Palettes are now limited to 256 colors so that a palette index can fit
  into a uint8 on disk.
* Chunks in your level and doodad files are now encoded in a binary format
  instead of JSON for a reduction in file size. The current (and only)
  chunk implementation (the MapAccessor) encodes to a binary format involving
  trios of varints (X, Y position + a Uvarint for palette index).
* Chunk sizes in levels/doodads is now a uint8 type, meaning the maximum
  chunk size is 255x255 pixels. The game's default has always been 128x128
  but now there is a limit. This takes a step towards optimizing the game's
  file formats: large world coordinates (64-bit) are mapped to a chunk
  coordinate, and if each chunk only needs to worry about the 255 pixels
  in its territory, space can be saved in memory without chunks needing to
  theoretically support 64-bit sizes of pixels!

## v0.13.1 (Oct 10 2022)

This release brings a handful of minor new features to the game.

First, there are a couple of new Pixel Attributes available in the level editor:

* Semi-Solid: pixels with this attribute only behave as "solid" when walked on
  from above. The player can jump through the bottom of a Semi-Solid and land
  on top, and gradual slopes can be walked up and down as well, but a steep
  slope or a wall can be simply passed through as though it were just decoration.
* Slippery: the player's acceleration and friction are reduced when walking on
  a slippery floor. In the future, players and other mobile doodads may slide
  down slippery slopes automatically as well (not yet implemented).
* These attributes are available in the Level Editor by clicking the "Edit"
  button on your Palette (or the "Tools -> Edit Palette" menu). The Palette
  Editor now has small icon images for the various attributes to make room for
  the expanded arsenal of options.

Doodad/Actor Runtime Options have been added:

* In the Doodad Editor's "Doodad Properties" window, see the new "Options" tab.
* Doodad Options allow a map creator to customize certain properties about your
  doodad, on a per-instance basis (instances of doodads are called "actors" when
  placed in your level).
* In the Level Editor when the Actor Tool is selected, mousing over a doodad on
  your level will show a new gear icon in the corner. Clicking the icon will open
  the Actor Properties window, where you may toggle some of the doodad options
  (if a doodad has any options available).
* Options can be of type boolean, string, or integer and have a custom name and a
  default value at the doodad level. In the Level Editor, the map creator can
  set values for the available options which the doodad script can read using the
  `Self.GetOption()` method.
* Several of the game's built-in doodads have options you can play with, which are
  documented below.

New and updated doodads:

* "Look At Me" is a new Technical doodad that will draw the camera's attention
  to it when it receives a power signal from a linked button. For example, if
  a button would open an Electric Door far across the level, you can also place
  a "Look At Me" near the door and link the button to both doodads. When the
  button is pressed, the camera will scroll to the "Look At Me" and the player
  can see that the door has opened.
* Anvils will now attract the camera's attention while they are falling.

Several of the game's built-in doodads have new Actor Runtime Options you can
configure in your custom levels:

* Warp Doors: "locked (exit only)" will make it so the player can not enter the
  warp door - they will get a message on-screen that it is locked, similar to
  how warp doors behave when they aren't linked to another door. If it is linked
  to another door, the player may still exit from the 'locked' door -
  essentially creating a one-way warp, without needing to rely on the
  orange/blue state doors. The "Invisible Warp Door" technical doodad also
  supports this option.
* Electric Door & Electric Trapdoor: check the "opened" option and these doors
  will be opened by default when the level gameplay begins. A switch may still
  toggle the doors closed, or if the doors receive and then lose a power signal
  they will close as normal.
* Colored Doors & Small Key Door: you may mark the doors as "unlocked" at the
  start of your level, and they won't require a key to open.
* Colored Keys & Small Key: you may mark the keys as "has gravity" and they
  will be subject to the force of gravity and be considered a "mobile" doodad
  that may activate buttons or trapdoors that they fall onto.
* Gemstones: these items already had gravity by default, and now they have a
  "has gravity" option you may disable if you'd prefer gemstones not to be
  subject to gravity (and make them behave the way keys used to).
* Gemstome Totems: for cosmetic purposes you may toggle the "has gemstone"
  option and the totem will already have its stone inserted at level start.
  These gemstones will NOT emit a power signal or interact normally with
  linked totems - they should be configured this way only for the cosmetic
  appearance, e.g., to have one totem filled and some others empty; only the
  empty totems should be linked together and to a door that would open when
  they are all filled.
* Fire Region: you may pick a custom "name" for this doodad (default is "fire")
  to make it better behave as normal fire pixels do: "Watch out for (name)!"

Improvements in support of custom content:

* Add a JavaScript "Exception Catcher" window in-game. If your doodad scripts
  encounter a scripting error, a red window will pop up showing the text of
  the exception with buttons to copy the full text to your clipboard (in case
  it doesn't all fit on-screen) and to suppress any further exceptions for
  the rest of your game session (in case a broken doodad is spamming you with
  error messages). Cheat codes can invoke the Exception Catcher for testing:
  `throw <message>` to show custom text, `throw2` to test a "long" message
  and `throw3` to throw a realistic message.
* Calling `console.log()` and similar from doodad scripts will now prefix the
  log message with the doodad's filename and level ID.

There are new JavaScript API methods available to doodad scripts:

* `Self.CameraFollowMe()` will attract the game's camera viewport to center
  on your doodad, taking the camera's focus away from the player character.
  The camera will return to the player if they enter a directional input.
* `Self.Options()` returns a string array of all of the options available on
  the current doodad.
* `Self.GetOption(name)` returns the configured value for a given option.

Some improvements to the `doodad` command-line tool:

* `doodad show` will print the Options on a .doodad file and, when showing
  a .level file with the `--actors` option, will list any Options configured
  on a level's actors where they differ from the doodad's defaults.
* `doodad edit-doodad` adds a `--option` parameter to define an option on a
  doodad programmatically. The syntax is like `--option name=type=default`
  for example `--option unlocked=bool=true` or `--option unlocked=bool`; the
  default value is optional if you want it to be the "zero value" (false,
  zero, or empty string).

Minor fixes and improvements:

* Add a "Wait" modal with a progress bar. Not used yet but may be useful
  for long operations like Giant Screenshot or level saving to block input
  to the game while it's busy doing something. Can be tested using the
  cheat code "test wait screen"
* Detect the presence of a touchscreen device and automatically disable
  on-screen touch hints during gameplay if not on a touch screen.
* Mobile Linux: mark the Sketchy Maze launcher as supporting the mobile
  form-factor for the Phosh desktop shell especially.
* Fix the Crusher doodad sometimes not falling until it hits the ground
  and stopping early on slower computers.
* Small tweaks to player physics - acceleration increased from 0.025 to
  0.04 pixels per tick.

## v0.13.0 (May 7 2022)

This is a major update that brings deep architectural changes and a lot
of new content to the game.

Swimming physics have been added:

* The **water** pixels finally do something besides turn the character
  blue: swimming physics have finally been hooked up!
* While the player is touching water pixels (and is colored blue),
  your gravity and jump speed are reduced, but you can "jump"
  infinite times in order to swim higher in the water. Hold the jump
  button to climb slowly, spam it to climb quickly.
* The **Azulians** understand how to swim, too, though left to their
  own devices they will sink to the bottom of a body of water. They'll
  swim (jump) up if you're detected and above them. The Blue Azulian
  has the shortest vertical aggro radius, so it's not a very good swimmer
  but the White Azulian can traverse water with ease and track you
  from a greater distance.

New levels:

* **The Jungle** (First Quest) - a direct sequel to the Boat level, it's
  a jungle and Mayan themed platformer featuring many of the new doodads
  such as snakes, gemstones, and crushers.
* **Gems & Totems** (Tutorial, Lesson 4) - a tutorial level about the
  new Gem and Totem doodads.
* **Swimming** (Tutorial, Lesson 5) - a tutorial level to learn how
  "water pixels" work with some moderately safe platforming puzzles
  included.
* **Night Sky** (Azulian Tag) - a moderately difficult Azulian Tag level
  with relatively few enemies but plenty of tricky platforming.
* Some of the existing levels have had minor updates to take advantage
  of newer game features, such as the water being re-done for the Castle
  level.

New doodads:

* **Blue Bird:** a blue version of the Bird which flies in a sine wave
  pattern about its original altitude, and has a larger search range to
  dive at the player character.
* **Snake:** a green snake that sits coiled up and always faces the
  player. If you get nearby and try and jump over, the Snake will jump
  up and hope to catch you.
* **Crusher:** a block-headed enemy with an iron helmet which tries to
  drop on you from above. Its helmet makes a safe platform to ride
  back up like an elevator.
* **Gems and Totems:** four collectible gems (in different colors and
  shapes) that slot into Totems of a matching shape. Totems can link
  together to require multiple gemstones before they'll emit a power
  signal to other linked doodads.

New **File Formats**:

* Levels and Doodads have a new file format based on ZIP files,
  like levelpacks.
* It massively improves loading screen times and helps the
  game keep a substantially lighter memory footprint (up to 85%
  less memory used, like 1.5 GB -> 200 MB on _Azulian Tag - Forest_).
* **Your old levels and doodads still work**! The next time you save
  them, they will be converted to the new file format automatically.
* The `doodad` tool can also upgrade your levels by running:
  `doodad edit-level --touch <filename>.level`

Other new content to use in your levels:

* New wallpapers: Dotted paper (dark), Parchment paper (red, blue,
  green, and yellow).
* New palette: Neon Bright, all bright colors to pair with a dark level
  wallpaper.
* New brush pattern: Bubbles. The default "water" color of all the game's
  palettes will now use the Bubbles pattern by default.

New cheat codes:

* `super azulian`: play as the Red Azulian.
* `hyper azulian`: play as the White Azulian.
* `bluebird`: play as the Blue Bird.
* `warp whistle`: automatically win the current level, with a snarky
  cheater message in the victory dialog. It will mark the level as
  Completed but not reward a high score.
* `$ d.SetPlayerCharacter("anything.doodad")` - set your character to any
  doodad you want, besides the ones that have dedicated cheat codes. The
  ".doodad" suffix is optional. Some to try out are "key-blue", "anvil",
  or "box". If you are playing as a key, a mob might be able to collect
  you! This will softlock the level, but another call to
  `d.SetPlayerCharacter()` will fix it! Use the `pinocchio` cheat or
  restart the game to return to the default character (boy.doodad)

Updates to the JavaScript API for doodads:

* `Self.IsWet() bool` can test if your actor is currently in water.

Other changes this release:

* Editor: fancy **mouse cursors** gives some visual feedback about what
  tool is active in the editor, with a Pencil and a Flood Fill
  cursor when those tools are selected.
* Editor: your **Palette** buttons will now show their pattern with their
  color as the button face, rather than just the color.
* Editor: Auto-save is run on a background thread so that, for large
  levels, it doesn't momentarily freeze the editor on save when it runs.
* Editor: Fix the Link Tool forgetting connections when you pick up and
  drop one of the linked doodads.
* Link Tool: if you click a doodad and don't want to link it to another,
  click the first doodad again to de-select it (or change tools).
* Editor: your last scroll position on a level is saved with it, so the
  editor will be where you left it when you reopen your drawing.
* Doodad tool: `doodad edit-level --resize <int>` can re-encode a level
  file using a different chunk size (the default has been 128).
  Experimental! Very large or small chunk sizes can lead to different
  performance characteristics in game!
* Fixed the bug where characters' white eyes were showing as transparent.

## v0.12.1 (April 16 2022)

This update focuses on memory and performance improvements for the game.
Some larger levels such as "Azulian Tag - Forest" could run out of
memory on 32-bit systems. To improve on memory usage, the game more
aggressively frees SDL2 textures when no longer needed and it doesn't
try to keep the _whole_ level's chunks ready in memory (as rendered
images) -- only the chunks near the window viewport are loaded, and
chunks that leave the viewport are freed to reclaim memory. Improvements
are still a work in progress.

New fields are added to the F3 debug overlay to peek at its performance:

* "Textures:" shows the count of SDL2 textures currently loaded
  in memory. Some textures, such as toolbar buttons and the
  loadscreen wallpaper, lazy load and persist in memory. Most level
  and doodad textures should free when the level exits.
* "Sys/Heap:" shows current memory utilization in terms of MiB taken
  from the OS and MiB of active heap memory, respectively.
* "Threads:" counts the number of goroutines currently active. In
  gameplay, each actor monitors its PubSub queue on a goroutine and
  these should clean up when the level exits.
* "Chunks:" shows the count of level chunks inside and outside the
  loading viewport.

Some other changes and bug fixes in this release include:

* Fixed the bug where the player was able to "climb" vertical walls to
  their right.
* When entering a cheat code that changes the default player character
  during gameplay, you immediately become that character.

## v0.12.0 (March 27 2022)

This update adds several new features to gameplay and the Level Editor.

A **Game Rules** feature has been added to the Level Editor which allows
customizing certain gameplay features while that level is being played.
These settings are available in the Level Properties window of the editor:

* The **Difficulty** rule can modify the behavior of enemy doodads
  when the level is played. Choose between Peaceful, Normal, or Hard.
    * On Peaceful, Azulians and Birds don't attack the player, acting
      like pre-0.11.0 versions that ignored the player character.
    * On Hard difficulty, Azulians have an infinite aggro radius
      (they'll immediately hunt the player from any distance on
      level start) and they are hostile to _all_ player creatures.
* **Survival Mode** changes the definition of "high score" for levels
  where the player is very likely to die at least once.
    * The silver high score (respawned at checkpoint) will be for the
      _longest_ time survived on the level rather than the fastest time
      to complete it.
    * The gold high score (got to an Exit Flag without dying once) is
      still rewarded to the fastest time completing the level.

An update to the Level Editor's toolbar:

* New **Text Tool** to easily paste messages onto your level, selecting
  from the game's built-in fonts.
* New **Pan Tool** to be able to scroll the level safely by dragging with
  your mouse or finger.
* New **Flood Tool** (or paint bucket tool) can be used to replace
  contiguous areas of your level from one color to another.
* The toolbar buttons are smaller and rearranged. On medium-size screens
  or larger, the toolbar buttons are drawn side-by-side in two columns.
  On narrower screens with less real estate, it will use a single column
  when it fits better.

New levels:

* An **Azulian Tag** levelpack has been added featuring two levels so
  far of the Azulian Tag game mode.

New doodads:

* A technical doodad for **Reset Level Timer** resets the timer to zero,
  one time, when touched by the player. If the doodad receives a power
  signal from a linked doodad, it can reset the level timer again if the
  player touches it once more.

Updates to the JavaScript API for custom doodads:

* New global integer `Level.Difficulty` is available to doodad scripts to
  query the difficulty setting of the current level:
  * Peaceful (-1): `Level.Difficulty < 0`
  * Normal (0): `Level.Difficulty == 0`
  * Hard (1): `Level.Difficulty > 1`
* New function `Level.ResetTimer()` resets the in-game timer to zero.

New cheat codes:

* `test load screen` tests the loading screen UI for a few seconds.
* `master key` allows playing locked Story Mode levels without unlocking
  them first by completing the earlier levels.

Other changes:

* Several screens were re-worked to be responsive to mobile screen sizes,
  including the Settings window, loading screen, and the level editor.
* Fixed a bug where making the app window bigger during a loading screen
  caused the Editor to not adapt to the larger window.
* Don't show the _autosave.doodad in the Doodad Dropper.
* The Azulians have had their jump heights buffed slightly.
* Birds no longer register as solid when colliding with other birds (or
  more generally, characters unaffected by gravity).

Bugs fixed:

* When modifying your Palette to rename a color or add an additional
  color, it wasn't possible to draw with that new color without fully
  exiting and reloading the editor - this is now resolved.
* The palette editor will try and prevent the user from giving the same
  name to different colors.

## v0.11.0 (Feb 21 2022)

New features:

* **Game Controller** support has been added! The game can now be played
  with an Xbox style controller, including Nintendo Pro Controllers. The
  game supports an "X Style" and "N Style" button layout, the latter of
  which swaps the A/B and the X/Y buttons so gameplay controls match the
  button labels in your controller.
* The **JavaScript Engine** for doodad scripts has been switched from
  github.com/robertkrimen/otto to github.com/dop251/goja which helps
  "modernize" the experience of writing doodads. Goja supports many
  common ES6 features already, such as:
  * Arrow functions
  * `let` and `const` keywords
  * Promises
  * for-of loops
* The **JavaScript API** has been expanded with new functions and
  many of the built-in Creatures have gotten an A.I. update.
* For full versions of the game, the **Publish Level** function is now
  more streamlined to just a checkbox for automatically bundling your
  doodads next time you save the level.

New levels:

* **The Zoo:** part of the Tutorial levelpack, it shows off all of the
  game's built-in doodads and features a character selection room to
  play as different creatures.
* **Shapeshifter:** a new level on the First Quest where you switch
  controls between the Boy, Azulian and the Bird in order to clear the
  level.

Some of the built-in doodads have updates to their A.I. and creatures
are becoming more dangerous:

* The **Bird** now searches for the player diagonally in front of
  it for about 240px or so. If spotted it will dive toward you and
  it is dangerous when diving! When _playing_ as the bird, the dive sprite
  is used when flying diagonally downwards. The player-controlled bird
  can kill mobile doodads by diving into them.
* The **Azulians** will start to follow the player when you get
  close and they are dangerous when they touch you -- but not if
  you're the **Thief.** The red Azulian has a wider search radius,
  higher jump and faster speed than the blue Azulian.
* A new **White Azulian** has been added to the game. It is even faster
  than the Red Azulian! And it can jump higher, too!
* The **Checkpoint Flag** can now re-assign the player character when
  activated! Just link a doodad to the Checkpoint Flag like you do the
  Start Flag. When the player reaches the checkpoint, their character
  sprite is replaced with the linked doodad!
* The **Anvil** is invulnerable -- if the player character is the Anvil
  it can not die by fire or hostile enemies, and Anvils can not destroy
  other Anvils.
* The **Box** is also made invulnerable so it can't be destroyed by a
  player-controlled Anvil or Bird.

New functions are available on the JavaScript API for doodads:

* `Actors.At(Point) []*Actor`: returns actors intersecting a point
* `Actors.FindPlayer() *Actor`: returns the nearest player character
* `Actors.New(filename string)`: create a new actor (NOT TESTED YET!)
* `Self.Grounded() bool`: query the grounded status of current actor
* `Actors.SetPlayerCharacter(filename string)`: replace the nearest
  player character with the named doodad, e.g. "boy.doodad"
* `Self.Invulnerable() bool` and `Self.SetInvulnerable(bool)`: set a
  doodad is invulnerable, especially for the player character, e.g.
  if playing as the Anvil you can't be defeated by mobs or fire.

New cheat codes:

* `god mode`: toggle invincibility. When on, fire pixels and hostile
  mobs can't make you fail the level.
* `megaton weight`: play as the Anvil by default on levels that don't
  specify a player character otherwise.

Other changes:

* When respawning from a checkpoint, the player is granted 3 seconds of
  invulnerability; so if hostile mobs are spawn camping the player, you
  don't get soft-locked!
* The draw order of actors on a level is now deterministic: the most
  recently added actor will always draw on top when overlapping another,
  and the player actor is always on top.
* JavaScript exceptions raised in doodad scripts will be logged to the
  console instead of crashing the game. In the future these will be
  caught and presented nicely in an in-game popup window.
* When playing as the Bird, the flying animation now loops while the
  player is staying still rather than pausing.
* The "Level" menu in Play Mode has options to restart the level or
  retry from last checkpoint, in case a player got softlocked.
* When the game checks if there's an update available via
  <https://download.sketchymaze.com/version.json> it will send a user
  agent header like: "Sketchy Maze v0.10.2 on linux/amd64" sending only
  static data about the version and OS.

## v0.10.1 (Jan 9 2022)

New features:

* **High scores and level progression:** when playing levels out of
  a Level Pack, the game will save your progress and high scores on
  each level you play. See details on how scoring works so far, below.
* **Auto-save** for the Editor. Automatically saves your drawing every
  5 minutes. Look for e.g. the _autosave.level to recover your drawing
  if the game crashed or exited wrongly!
* **Color picker UI:** when asked to choose a color (e.g. for your level
  palette) a UI window makes picking a color easy! You can still manually
  enter a hexadecimal color value but this is no longer required!

Scoring system:

* The high score on a level is based on how quickly you complete it.
  A timer is shown in-game and there are two possible high scores
  for each level in a pack:
    * Perfect Time (gold): if you complete the level without dying and
      restarting at a checkpoint.
    * Best Time (silver): if you had used a checkpoint.
* The gold/silver icon is displayed next to the timer in-game; it
  starts gold and drops to silver if you die and restart from checkpoint.
  It gives you a preview of which high score you'll be competing with.
* If cheat codes are used, the user is not eligible for a high score
  but may still mark the level "completed."
* Level Packs may have some of their later levels locked by default,
  with only one or a few available immediately. Completing a level will
  unlock the next level until they have all been unlocked.

New and changed doodads:

* **Invisible Warp Door:** a technical doodad to create an invisible
  Warp Door (press Space/'Use' key to activate).
* All **Warp Doors** now require the player to be grounded before they
  can be opened, or else under the effects of antigravity. You shouldn't
  be able to open Warp Doors while falling or jumping off the ground
  at them.

Revised levels:

* Desert-2of2.level uses a new work-around for the unfortunate glitch
  of sometimes getting stuck on two boxes, instead of a cheat code
  being necessary to resolve.
* Revised difficulty on Tutorial 2 and Tutorial 3.

Miscellaneous changes:

* Title Screen: picks a random level from a few options, in the future
  it will pick random user levels too.
* Play Mode gets a menu bar like the Editor for easier navigation to
  other game features.
* New dev shell command: `titlescreen <level name>` to load the Title
  Screen with the named level as its background, can be used to load
  user levels _now_.
* For the doodads JavaScript API: `time.Since()` is now available (from
  the Go standard library)
* Fix a cosmetic bug where doodads scrolling off the top or left edges
  of the level were being drawn incorrectly.
* Fix the Editor's status bar where it shows your cursor position
  relative to the level and absolute to the app window to show the
  correct values of each (they were reversed before).
* The developer shell now has a chatbot in it powered by RiveScript.

## v0.10.0 (Dec 30 2021)

New features and changes:

* **Level Packs:** you can group a set of levels into a sequential
  adventure. The game's built-in levels have been migrated into Level
  Packs and users can create their own, too!
* **Crosshair Option:** in the level editor you can have a crosshair
  drawn at your cursor position, which may help align things while
  making a level. Find the option in the game's Settings window.
* **Smaller Palette Colors:** the color buttons in the Level Editor are
  smaller and fit two to a row. This allows for more colors but may be
  difficult for touch controls.
* Doodad AI updates: the **Bird** records its original altitude and will
  attempt to fly back there when it can, so in case it slid up or down a
  ramp it will correct its height when it comes back the other way.
* The "New Level" and "New Doodad" functions on the main menu are
  consolidated into a window together that can create either, bringing
  a proper UI to creating a doodad.
* Added a setting to **hide touch control hints** from Play Mode.
* The title screen is more adaptive to mobile. If the window height isn't
  tall enough to show the menu, it switches to a 'landscape mode' layout.
* Adds a custom icon to the application window.

A few notes about level packs:

* A levelpack is basically a zip file containing levels and custom
  doodads. **Note:** the game does not yet handle the doodads folder
  of a levelpack at all.
* The `doodad` command-line tool can create .levelpack files. See
  `doodad levelpack create --help`. This is the easiest way to
  generate the `index.json` file.
* In the future, a .levelpack will be able to hold custom doodads on
  behalf of the levels it contains, de-duplicating files and saving
  on space. Currently, the levels inside your levelpack should embed
  their own custom doodads each.
* Free (shareware) editions of the game can create and play custom
  level packs that use only the game's built-in doodads. Free versions
  of the game can't play levels with embedded custom doodads. You can
  always copy custom .doodad files into your profile directory though!

Bugs fixed:

* Undo/Redo now works again for the Doodad Editor.
* Fix crash when opening the Doodad Editor (v0.9.0 regression).
* The Play Level/Edit Drawing window is more responsive to small screens
  and will draw fewer columns of filenames.
* Alert and Confirm popup modals always re-center themselves, especially
  to adapt to the user switching from Portrait to Landscape orientation
  on mobile.

## v0.9.0 (Oct 9, 2021)

New features:

* **Touch controls:** the game is now fully playable on small
  touchscreen devices! For gameplay, touch the middle of the screen
  to "use" objects and touch anywhere _around_ that to move the
  player character in that direction. In the level editor, swipe
  with two fingers to pan and scroll the viewport.
* **Proper platforming physics:** the player character "jumps" by
  setting a velocity and letting gravity take care of the rest,
  rather than the old timer-based flat jump speed approach. Most
  levels play about the same but the jump reach _is_ slightly
  different than before!
* **Giant Screenshot:** in the Level Editor you can take a "Giant
  Screenshot" of your _entire_ level as it appears in the editor
  as a large PNG image in your game's screenshots folder. This is
  found in the "Level" menu in the editor.
* **Picture-in-Picture Viewport:** in the Level Editor you can open
  another viewport window into your level. This window can be resized,
  moved around, and can look at a different part of your level from
  your main editor window. Mouse over the viewport and your arrow keys
  scroll it instead of the main window.
* **"Playtest from here":** if you click and drag from the "Play (P)"
  button of the level editor, you can drop your player character
  anywhere you want on your level, and play from there instead of
  the Start Flag.
* **Middle-click to pan the level:** holding the middle click button
  and dragging your mouse cursor will scroll the level in the editor
  and on the title screen. This can be faster than the arrow keys to
  scroll quickly!
* **Zoom in/out** support has come out of "experimental" status. It's
  still a _little_ finicky but workable!
* **Bounded level limits** are now configurable in the Level Properties
  window, to set the scroll constraints for Bounded level types.

Some new "technical" doodads are added to the game. These doodads are
generally invisible during gameplay and have various effects which can
be useful to set up your level the right way:

* **Goal Region:** a 128x128 region that acts as an Exit Flag and will
  win the level if touched by the player.
* **Checkpoint Region:** a 128x128 invisible checkpoint flag that sets
  the player's spawn point.
* **Fire Region:** a 128x128 region that triggers the level fail
  condition, behaving like fire pixels.
* **Power Source:** a 64x64 invisible source of power. Link it to doodads
  like the Electric Door and it will emit a power(true) signal on level
  start, causing the door to "begin" in an opened state.
* **Stall Player (250ms):** when the player touches this doodad, control
  freezes for 250ms, one time. If this doodad receives power from a linked
  button or switch, it will reset the trap, and stall the player once more
  should they contact it again.

The Doodad Dropper window has a dedicated category for these
technical doodads.

New methods available to the Doodad JavaScript API:

* `Self.Hide()` and `Self.Show()` to turn invisible and back.
* `Self.GetVelocity()` that returns a Vector of the actor's current speed.
* New broadcast message type: `broadcast:ready` (no arguments). Subscribe
  to this to know the gameplay is ready and you can safely publish messages
  to your linked doodads or whatever, which could've ended up in deadlocks
  if done too early!
* New constants are exposed for Go time.Duration values: `time.Hour`,
  `time.Minute`, `time.Second`, `time.Millisecond` and `time.Microsecond`
  can be multiplied with a number to get a duration.

Some slight UI polish:

* The **Doodad Dropper** window of the level editor now shows a slotted
  background where there are any empty doodad slots on the current page.
* Your **scroll position** is remembered when you playtest the level; so
  coming back to the editor, your viewport into the level as where you
  left it!
* New **keybinds** are added to the Level & Doodad Editor:
    * `Backspace` to close the current popup UI modal
    * `Shift+Backspace` to close _all_ popup UI modals
    * `v` to open a new Viewport window
* Some **flashed messages** are orange instead of blue to denote an 'error'
  status. The colors of the fonts are softened a bit.
* New **cheat code**: `show all actors` to make all invisible actors shown
  during Play Mode. You would be able to see all the technical doodads
  this way!
* New command for the doodad CLI tool: `doodad edit-level --remove-actor`
  to remove actors by name or UUID from a level file.

## v0.8.1 (September 12, 2021)

New features:

* **Enable Experimental Features UI:** in the game's Settings window
  there is a tab to enable experimental features. It is equivalent to
  running the game with the `--experimental` option but the setting
  in-game is persistent across runs of the game.
* **Zoom In/Out:** the Zoom feature _mostly_ works but has a couple
  small bugs. The `+` and `-` keys on your number bar (-=) will zoom in
  or out in the Editor. Press `1` to reset zoom to 100% and press `0`
  to scroll the level back to origin. These controls are also
  available in the "View" menu of the editor.
* **Replace Palette:** with experimental features on it is possible to
  select a different palette for your already-created level. This will
  replace colors on your palette until the template palette has been
  filled in, and pixels already drawn on your level will update too.

Other changes:

* The title screen buttons are more colorful.
* This release begins to target 32-bit Windows and Linux among its builds.

## v0.8.0 (September 4, 2021)

This release brings some new features, new doodads, and new levels.

New features:

* **Checkpoints** for gameplay will ease the pain of dying to fire
  pixels or Anvils by teleporting you back to the checkpoint instead
  of resetting the whole level.
* The **Doodad Properties** window while editing a doodad grants access
  to many features which were previously only available via the
  `doodad` tool, such as:
    * Edit metadata like the Title and Author of your doodad
    * Set the default hitbox of your doodad.
    * Attach, open, and delete the JavaScript for your doodad
    * Manage tags (key/value store) on your doodads: how you can
      communicate settings to the JavaScript which can receive the
      tags via `Self.GetTag("name")`
* Some **Generic Doodad Scripts** are built in. Using only the in-game
  tools, it is possible to create custom doodads which have some basic
  in-game logic and you don't need to write any code. The generic
  scripts include:
    * Generic Solid: the hitbox is solid
    * Generic Fire: its hitbox harms the player
    * Generic Anvil: harmless, deadly when falling
    * Generic Collectible Item: it goes in your inventory
* **All Characters are Playable!** Use the Link Tool to connect your
  Start Flag with another doodad on your level, and you will play
  **as** that doodad when the level starts. The Creature doodads are
  all intended to be fully functional; playing as buttons and doors
  leads to strange results.

New doodads have been added:

* The **Anvil** is a heavy metal object which is affected by gravity.
  It is harmless to collision, but if the anvil is in freefall, it
  will destroy every mobile doodad that it touches, and is deadly
  to the player character.
* The **Electric Trapdoor** is a trapdoor that opens and closes when
  powered by a button or switch. It is a horizontal version of the
  Electric Door.
* The **Thief** is a new character which will steal items from the
  player or other mobile doodads. The Thief is able to pick up items
  and unlock doors and he walks back and forth like the Azulians.
* The **Blue Azulian** is now selectable from the Doodads menu. It
  behaves like the Red Azulian but moves at half the speed. The
  Azulians can pick up items and open doors.
* The **Checkpoint Flag** will remember the player's spot in the level.
  Dying to fire pixels or Anvils no longer forces a restart of the
  level - you can resume from your last checkpoint, or the Start Flag
  by default.

New levels have been added:

* **Castle.level:** introduces the new Thief character. Castle-themed
  level showing off various new doodads.
* **Thief 1.level:** a level where you play as the Thief! You need to
  steal Small Keys from dozens of Azulians and even steal items back
  from another Thief who has already stolen some of the keys.

Some doodads have changed behavior:

* The **Bird** can no longer pick up items, unless controlled by
  the player character.
* The **Anvil** and **Box** will reset to their original locations
  if they receive a power signal from a linked button or switch.

The user interface has been improved:

* **Tabbed windows!** The Doodad Dropper window of the level editor
  and the Settings window use new, tabbed interfaces.
* **Doodad Categories:** the Doodad Dropper's tabs are divided into
  a few major categories.
    1. Objects: Anvil, Box, Crumbly Floor, and Flags
    2. Doors: Doors, Trapdoors, and Keys
    3. Gizmos: Buttons, Switches, Electric Doors, etc.
    4. Creatures: Bird, Azulians, Thief
    5. All: a classic view paging over all doodads (and doodads
       not fitting any of the above categories).

New functions are available in the JavaScript API for custom doodads:

* FailLevel(message string): global function that kills the player
  with a custom death message.
* SetCheckpoint(Point): set the player respawn location
* Self.MoveTo(Point(x, y int))
* Self.IsPlayer() bool
* Self.SetInventory(bool): turn on or off inventory. Keys and other
  items will now only give themselves to mobile doodads which have
  inventory.
* Self.HasInventory() bool
* Self.AddItem(filename string, quantity int) - zero quantity for
  permanent items like the colored keys.
* Self.RemoveItem(filename string, quantity int)
* Self.HasItem(filename string)
* Self.Inventory() map[string]int
* Self.Hitbox() - also see Self.Hitbox.IsEmpty()

The Events.OnLeave() callback now receives a CollideEvent argument,
like OnCollide, instead of the useless actor ID string. Notable
properties on the CollideEvent will be the .Actor which is leaving
and Settled=true.

Other miscellaneous changes:

* The **Link Tool** can now un-link two doodads by clicking on
  them again.
* Actor UUIDs in your levels will now be Type 1 UUIDs (time-based)
  instead of random. This will ensure each newly added doodad gets
  a larger ID than the previous one, so in cases of draw order
  conflicts or that sort of thing, the winner can be picked
  deterministically (most recently added renders on top).
* A **death barrier** will prevent Boy from falling forever on unbounded
  maps should he somehow fall off the level. The death barrier is a
  Y value 1,000 pixels below the lowest pixel on your map.
* Mobile doodads no longer "moonwalk" when they change directions.
* A new color is added to all default palettes: "hint" (pink) for
  writing hint notes.
* A maximum scroll speed on the "follow the player character" logic
  makes for cooler animations when the character teleports around.
* Levels and Doodads are now sorted on the Open menu.

## v0.7.2 (July 19 2021)

This release brings some new features and some new content.

New features:

* **Loading screens** have been added. In previous versions, 'chunks'
  of a level were rendered on-demand as they first scrolled onto the
  screen, but busy new chunks would cause gameplay to stutter. The
  loading screen is able to _pre-render_ the entire level up front
  to ensure gameplay is smooth(er).
* **Compression for Levels and Doodads:** levels and doodads are now compressed
  with Gzip for an average 88% smaller file size on disk. The example level,
  "Tutorial 2.level" shrank from 2.2 MB to 414 KB with compression. The game
  can still read old (uncompressed) files, but will compress them on save.

Some new content:

* **New Levels:** two desert-themed levels were added. One has you
  platforming and climbing the outside of a pyramid; the next is
  inside the pyramid and has puzzles involving movable boxes.
* **New Doodad:** Box is a pushable crate that is affected by
  gravity. It is taller than Boy and can be used to gain some extra
  height and reach higher platforms.
* **New Palette:** To the "Colored Pencil" palette was added a new
  default color: Sandstone (solid).
* **New Pattern:** Perlin Noise was added as one of the available
  brush patterns. Sandstone uses this pattern by default.

Some miscellaneous changes:

* **Slowly scroll in editor:** holding down the Shift key while scrolling the
  level editor will scroll your drawing _very_ slowly.
* **'Doodads' Hotkey:** in the Level Editor, the `Q` key will open
  the Doodads window instead of the `D` key, as the `D` key is now
  part of WASD for scrolling the drawing.

## v0.7.1 (July 11 2021)

Fixes a bug on the Windows version:

* Built-in wallpapers other than the default Notebook were failing to
  load in the Windows release of v0.7.0

## v0.7.0 (June 20 2021)

This is the first release of the game where the "free version" drifts meaningfully
away from the "full version". Free versions of the game will show the label
"(shareware)" next to the game version numbers and will not support embedding
doodads inside of level files -- for creating them or playing them.
Check the website for how you can register the full version of the game.

This release brings several improvements to the game:

* **Brush Patterns** for your level palette. Instead of your colors drawing on as
  plain, solid pixels, a color swatch can _sample_ with a Pattern to create a
  textured appearance when plotted on your level. Several patterns are built in
  including Noise, Marker, Ink, and others. The idea is that your brush strokes can
  look as though they were drawn in pencil graphite or similar.
* **Title Screen:** the demo level shown on the title screen will leisurely scroll
  around the page. The arrow keys may still manually scroll the level any direction.
* **Attach Doodads to Level Files:** this is the first release that supports _truly_
  portable custom levels! By attaching your custom doodads _with_ your custom level
  file, it will "just play" on someone else's computer, and they don't need to copy
  all your custom doodads for it to work! But, free versions of the game will not
  get to enjoy this feature.
* **Settings UI**: a "Settings" button on the home screen (or the Edit->Settings
  menu in the editor) will open a settings window. Check it out!
* **Horizontal Toolbars option:** if enabled in the Settings window, the toolbar
  and palette in the Editor will be horizontal instead of vertical, along the top
  and bottom of the screen. This may be better optimized for smartphone-sized
  screens like the Pinephone. If the program is started with `-w mobile` the first
  time, it will use horizontal toolbars by default.

Some small bits of polish in the game's user interface:

* Some buttons are more colorful! The "Ok" button in alert boxes is blue and pressing
  Enter will select the blue button.
* When opening a drawing to play or edit, a blue **Browse...** button is
  added so you can more easily find downloaded custom levels and play them.
* In the Level Editor, the "Level ->  **Attached Files**" menu will let you see
  and manage files attached to your level, such as its custom wallpaper image or
  any custom doodads that were published with the level.
* The keyboard shortcut to open the developer console is now the tilde/grave key
  (`) instead of Enter.

Bugs fixed:

* The WASD keys to move the player character (as an alternative to the arrow keys)
  now works more reliably. Previously, they were affected by key-repeat so Boy would
  do a quick hop followed by a longer one when pressing W to jump. Also, his
  animation would not update correctly when moving via the WASD keys. Both bugs
  are fixed in this release.
* Shortcut keys advertised in the menu, such as Ctrl-N and Ctrl-S, now actually work.

## v0.6.0-alpha (June 6 2021)

This release brings less jank and some new features.

The new features:

* **Choice of Default Palette for New Levels:** when creating a new level, a
  "Palette:" option appears which allows you to set the default colors to start
  your level with. The options include:
  * Default: the classic default 4 colors (black, grey, red, blue).
  * Colored Pencil: a set with more earthy tones for outdoorsy levels
    (grass, dirt, stone, fire, water)
  * Blueprint: the classic Blueprint wallpaper theme, a bright version of Default
    for dark level backgrounds.
* **Custom Wallpapers:** unhappy with the default, paper-themed level background
  images? You can now use your own! They attach to your level data for easy
  transport when sharing your level with others.
* **More Wallpapers:** a couple of new default wallpapers are added: Graph paper
  and Dotted paper. They are both on light white paper and offer a light
  alternative to Blueprint.

Some bugs fixed:

* **Collision fixes:** you should be able to walk up gentle slopes to the left
  without jumping, as easily as you could to the right.
* **Debugging:** the F4 key to show collision hitboxes around all doodads in
  Play Mode now functions again, and draws boxes around _all_ doodads, not just the
  player character.
* **Hitboxes are tighter:** a doodad's declared hitbox size (from their JavaScript)
  is used when a doodad collides against level geometry or other doodads. Meaning:
  Boy, who has a narrow body but a square sprite size, now collides closer with
  objects on his right side.
* **Physics are tweaked:** Boy now moves and accelerates slightly faster.

## v0.5.0-alpha (Mar 31 2021)

Project: Doodle is renamed to Sketchy Maze in this release.

New Features:

* **New Tutorial Levels:** the bundled levels demonstrate the built-in doodads
  and how they interact and shows off several game features.
* **Level Editor:** you can now set the Title and Author of the level you're
  editing by using the Level->Page Settings window.
* The **Inventory HUD** in Play Mode now shows a small number indicator for items
  which have quantity, such as the Small Key. Colored Keys do not have quantity
  and don't show a number: those are permanent collectibles.
* **Fire Pixels:** when the player character dies by touching a "Fire" pixel
  during gameplay, the death message uses the **name** of the color instead
  of calling it "fire." For example, if you name a color "spikes" and give
  it the Fire attribute, it will say "Watch out for spikes!" if the player
  dies by touching it.
* New cheat code: `give all keys` gives all four colored keys and 99x Small Keys
  to the player character. The `drop all items` cheat clears your inventory.

New Doodads:

* **Warp Doors** allow the player character to fast travel to another location
  on the map. Drag two Warp Doors into your level and use the Link Tool to
  connect them together. Doors without an exit link will be "locked" and don't
  open.
* **Small Key Doors** are locked doors which consume the Small Keys when unlocked,
  unlike the colored doors where the key is multi-use. The player character can
  hold many small keys at once and only unlock as many doors as he has keys.

Updated Doodads:

* **Several doodads** were increased in size to better match the player character:
  Colored Locked Doors, Trapdoors, the Crumbly Floor and Electric Door, and the
  blue and orange Boolean State Blocks.
* **Colored Doors** now have a visual locked vs. unlocked state: while locked, a
  golden padlock hangs from the door, which goes away after it's been unlocked.
* **Switches** now interact differently with Electric Doors: the door will _always_
  toggle its current state regardless of the 'power' setting of the Switch.
* **Buttons** which are linked to a **Sticky Button** will press and stay down
  if the linked Sticky Button is pressed. Or in other words, the Sticky Button
  makes all linked Buttons act sticky too and stay pressed while the Sticky
  Button is pressed. If the Sticky Button is released later (e.g. by receiving
  power from a Switch) it releases its linked Buttons as well.

## v0.4.0-alpha (Nov 21 2020)

This update brings improvements to the editor; you can now fully draw all the
graphics for a custom doodad using the in-app tools!

A key difference between a Level and a Doodad is that doodads have **Layers**
where they keep multiple frames of animation or state. The editor tool now
supports working with these extra layers.

Other new features:

* The **Guidebook** has been updated with tons of good info and screenshots of
  the game's features. Press `F1` in-game to open the guidebook, or check it
  online at https://www.sketchymaze.com/guidebook/
* **Layer Selection Window for Doodads:** when you're editing a Doodad drawing,
  a "Lyr." button is added to the toolbar to access the Layers window.
* **Global UI popup modals:** when you're about to close a level with unsaved
  changes, you get an Ok/Cancel prompt confirming if you want to do that.
  Hitting the Escape key will ask you before just exiting the program. Alert
  boxes are supported too, and an `alert` command added to the developer console.

## v0.3.0-alpha (Sept 19 2020)

This update introduces the player character to the game. He doesn't have a name;
the game just refers to him as Boy.

His sprite size is bigger (33x54) than the 32x32 size of the placeholder player
character from before. So, the three example levels shipped with previous versions
of Project: Doodle no longer work. This release comes with two replacement levels
which are better decorated, anyway.

Other new features:

* **Palette Editor:** you can add and modify the colors on your level
  palette! Pick any colors you want, give a name to each swatch, and mark
  whether they behave as Solid, Water or Fire. Without any of these
  properties, colors are decorational-only by default.
* **Doodad Window:** when selecting the Actor or Link Tools in the editor,
  the Doodads window pops up. Doodads can be dragged from this window onto your
  level, instead of the Palette toolbar turning into a Doodad palette.
* The Palette toolbar on the Editor is now thinner and only shows colors.
  Mouse-over tooltips show the name and properties of each swatch.
* Added a --window option to the Doodle program to set the default window size.
  Options can be a resolution (e.g. 1024x768) or a special keyword
  "desktop", "mobile", "landscape" or "maximized"
  - Default size is desktop: 1024x768
  - Mobile and landscape mimic a smartphone at 375x812 resolution.

## v0.2.0-alpha (June 7 2020)

This release brings Sound Effects and Menus to the game.

New features:

* Added some User Documentation to ship with the game which teaches you how to
  create your own custom Doodads and program them with JavaScript. More
  documentation to come with time.
* Sound effects! Several doodads have a first pass at sound effects using some
  free sounds I found online. More doodads still need sounds and the existing
  sounds are by no means final. Buttons, switches, doors and keys have sound
  effects so far.
* The game now has a Menu Bar with pull-down menus in the Editor Mode instead
  of just a top panel with New/Save/Open buttons.

## v0.1.0-alpha (Apr 13 2020)

New doodads:

* Start Flag: drag this into your level to set where the player character will
  spawn. There should only be one per level.
* Crumbly Floor: a rocky floor that breaks and falls away after a couple
  seconds when the player (or other mobile doodad) walks onto it.
* State Blocks: blue and orange blocks that toggle between solid and passable
  when the corresponding ON/OFF button is touched.

New features:

* An inventory overlay now appears in Play Mode when the player character picks
  up one of the colored keys.
* While editing a level, you can click the new "Options" button in the top menu
  to open the level settings window (like the one you see when creating a new
  level): to change the wallpaper image or the page type.

Other changes:

* Added better platforming physics to the player character: acceleration and
  friction when walking.
* The colored Locked Door doodads have been re-designed to be shown in a
  side-view perspective and have an open and closed state in either direction.
* Tooltips added to various buttons in the Editor to show names of doodads and
  functions of various buttons.

## v0.0.10-alpha (July 18 2019)

New features:

* Added the **Eraser Tool** and support for **brush sizes**. Now you can clean
  up your mistakes and draw shapes with thicker lines!
* Added the **Ellipse Tool** for drawing elliptical shapes.
* Added a third example to the game's built-in levels.

Bug fixes:

* The Undo command now restores the original color of a pixel instead of just
  deleting it. Only works for thin lines so far.
* Improved collision detection algorithm to prevent players from clipping
  through a solid doodad, regardless of speed. This change is invisible this
  build, but opens the door to improvements in the 2D platforming physics and
  making the player character move and fall faster.
* Fix mobile non-player doodads from sometimes being able to clip through a
  solid doodad. For example, a Red Azulian could sometimes walk through a locked
  door without interacting with it.
* Sometimes hitting Undo would leave a broken "chunk" in your level, if the
  Undo operation deleted all pixels in that chunk. The broken chunk would show
  as a solid black square (non-solid) in the level. This has been fixed: empty
  chunks are now culled when the last pixel is deleted and existing level files
  will be repaired on next save.

## v0.0.9-alpha (July 9 2019)

First alpha release.
