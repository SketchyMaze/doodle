package doodle

import (
	"fmt"
	"strings"
	"time"

	"git.kirsle.net/SketchyMaze/doodle/pkg/balance"
	"git.kirsle.net/SketchyMaze/doodle/pkg/collision"
	"git.kirsle.net/SketchyMaze/doodle/pkg/cursor"
	"git.kirsle.net/SketchyMaze/doodle/pkg/doodads"
	"git.kirsle.net/SketchyMaze/doodle/pkg/gamepad"
	"git.kirsle.net/SketchyMaze/doodle/pkg/keybind"
	"git.kirsle.net/SketchyMaze/doodle/pkg/level"
	"git.kirsle.net/SketchyMaze/doodle/pkg/levelpack"
	"git.kirsle.net/SketchyMaze/doodle/pkg/log"
	"git.kirsle.net/SketchyMaze/doodle/pkg/modal"
	"git.kirsle.net/SketchyMaze/doodle/pkg/modal/loadscreen"
	"git.kirsle.net/SketchyMaze/doodle/pkg/physics"
	"git.kirsle.net/SketchyMaze/doodle/pkg/plus"
	"git.kirsle.net/SketchyMaze/doodle/pkg/plus/dpp"
	"git.kirsle.net/SketchyMaze/doodle/pkg/savegame"
	"git.kirsle.net/SketchyMaze/doodle/pkg/scripting"
	"git.kirsle.net/SketchyMaze/doodle/pkg/shmem"
	"git.kirsle.net/SketchyMaze/doodle/pkg/sprites"
	"git.kirsle.net/SketchyMaze/doodle/pkg/uix"
	"git.kirsle.net/go/render"
	"git.kirsle.net/go/render/event"
	"git.kirsle.net/go/ui"
)

// PlayScene manages the "Edit Level" game mode.
type PlayScene struct {
	// Configuration attributes.
	Filename               string
	Level                  *level.Level
	CanEdit                bool         // i.e. you came from the Editor Mode
	HasNext                bool         // has a next level to load next
	RememberScrollPosition render.Point // for the Editor quality of life
	SpawnPoint             render.Point // if not zero, overrides Start Flag

	// If this level was part of a levelpack. The Play Scene will read it
	// from the levelpack ZIP file in priority over any other location.
	LevelPack *levelpack.LevelPack

	// Private variables.
	d            *Doodle
	drawing      *uix.Canvas
	scripting    *scripting.Supervisor
	running      bool
	deathBarrier int          // Y position of death barrier in case of falling OOB.
	lastCursor   render.Point // position of cursor X,Y last tick

	// Score variables.
	startTime  time.Time // wallclock time when level begins
	perfectRun bool      // set false on first respawn
	cheated    bool      // user has entered a cheat code while playing

	// UI widgets.
	Supervisor    *ui.Supervisor
	screen        *ui.Frame // A window sized invisible frame to position UI elements.
	menubar       *ui.MenuBar
	editButton    *ui.Button
	winLevelPacks *ui.Window
	winOpenLevel  *ui.Window

	// Custom debug labels.
	debPosition   *string
	debViewport   *string
	debScroll     *string
	debWorldIndex *string
	debLoadUnload *string

	// Player character
	Player                *uix.Actor
	playerPhysics         *physics.Mover
	slipperyPhysics       *physics.Mover
	lastCheckpoint        render.Point
	playerLastDirection   float64   // player's heading last tick
	slippery              bool      // player is on a slippery surface
	antigravity           bool      // Cheat: disable player gravity
	noclip                bool      // Cheat: disable player clipping
	godMode               bool      // Cheat: player can't die
	godModeUntil          time.Time // Invulnerability timer at respawn.
	playerJumpCounter     int       // limit jump length
	jumpCooldownUntil     uint64    // future game tick for jump cooldown (swimming esp.)
	mustFollowPlayerUntil uint64    // first frames where anvils don't take focus from player

	// Inventory HUD. Impl. in play_inventory.go
	invenFrame   *ui.Frame
	invenItems   []string // item list
	invenDoodads map[string]*uix.Canvas

	// Cheats window
	cheatsWindow *ui.Window

	// Elapsed Time frame.
	timerFrame          *ui.Frame
	timerPerfectImage   *ui.Image
	timerImperfectImage *ui.Image
	timerLabel          *ui.Label

	// Touchscreen controls state.
	isTouching    bool
	playerIsIdle  bool // LoopTouchable watches for inactivity on input controls.
	idleLastStart time.Time
	idleHelpAlpha int // fade in UI hints
}

// Name of the scene.
func (s *PlayScene) Name() string {
	return "Play"
}

// Setup the play scene.
func (s *PlayScene) Setup(d *Doodle) error {
	s.d = d
	s.scripting = scripting.NewSupervisor()
	s.Supervisor = ui.NewSupervisor()

	// Show the loading screen.
	loadscreen.ShowWithProgress()
	go func() {
		if err := s.setupAsync(d); err != nil {
			log.Error("PlayScene.setupAsync: %s", err)
			return
		}

		loadscreen.Hide()
	}()

	return nil
}

// setupAsync initializes the play screen in the background, underneath
// a Loading screen.
func (s *PlayScene) setupAsync(d *Doodle) error {
	// Create an invisible 'screen' frame for UI elements to use for positioning.
	s.screen = ui.NewFrame("Screen")
	s.screen.Resize(render.NewRect(d.width, d.height))

	// Menu Bar
	s.menubar = s.setupMenuBar(d)
	s.screen.Pack(s.menubar, ui.Pack{
		Side:  ui.N,
		FillX: true,
	})

	// Level Exit handler.
	s.scripting.OnLevelExit(s.BeatLevel)
	s.scripting.OnLevelFail(s.FailLevel)
	s.scripting.OnSetCheckpoint(s.SetCheckpoint)

	// Initialize debug overlay values.
	s.debPosition = new(string)
	s.debViewport = new(string)
	s.debScroll = new(string)
	s.debWorldIndex = new(string)
	s.debLoadUnload = new(string)
	customDebugLabels = []debugLabel{
		{"Pixel:", s.debWorldIndex},
		{"Player:", s.debPosition},
		{"Viewport:", s.debViewport},
		{"Scroll:", s.debScroll},
		{"Chunks:", s.debLoadUnload},
	}

	// Initialize the "Edit Map" button.
	s.editButton = ui.NewButton("Edit", ui.NewLabel(ui.Label{
		Text: "Edit (E)",
		Font: balance.PlayButtonFont,
	}))
	s.editButton.Handle(ui.Click, func(ed ui.EventData) error {
		s.EditLevel()
		return nil
	})
	s.Supervisor.Add(s.editButton)

	// Set up the inventory HUD.
	s.setupInventoryHud()

	// Set up the elapsed time frame.
	{
		s.timerFrame = ui.NewFrame("Elapsed Timer")

		// Set the gold and silver images.
		gold, _ := sprites.LoadImage(s.d.Engine, balance.GoldCoin)
		silver, _ := sprites.LoadImage(s.d.Engine, balance.SilverCoin)
		s.timerPerfectImage = gold
		s.timerImperfectImage = silver
		s.timerLabel = ui.NewLabel(ui.Label{
			Text: "00:00",
			Font: balance.TimerFont,
		})

		if s.timerPerfectImage != nil {
			s.timerFrame.Pack(s.timerPerfectImage, ui.Pack{
				Side: ui.W,
				PadX: 2,
			})
		}
		if s.timerImperfectImage != nil {
			s.timerFrame.Pack(s.timerImperfectImage, ui.Pack{
				Side: ui.W,
				PadX: 2,
			})
			s.timerImperfectImage.Hide()
		}

		s.timerFrame.Pack(s.timerLabel, ui.Pack{
			Side: ui.W,
			PadX: 2,
		})

		s.screen.Place(s.timerFrame, ui.Place{
			Top:  40,
			Left: 40,
		})
	}

	// Initialize the drawing canvas.
	s.drawing = uix.NewCanvas(balance.ChunkSize, false)
	s.drawing.Name = "play-canvas"
	s.drawing.Compute(d.Engine)

	// Handler when an actor touches water or fire.
	s.drawing.OnLevelCollision = func(a *uix.Actor, col *collision.Collide) {
		a.SetWet(col.InWater)

		if col.InFire != "" {
			a.Canvas.MaskColor = render.Black
			if a.ID() == "PLAYER" { // only the player dies in fire.
				s.DieByFire(col.InFire)
			}
		} else if col.InWater {
			a.Canvas.MaskColor = render.DarkBlue
		} else {
			a.Canvas.MaskColor = render.Invisible
		}

		// Slippery floor?
		s.slippery = col.IsSlippery
	}

	// Handle a doodad changing the player character.
	s.drawing.OnSetPlayerCharacter = s.SetPlayerCharacter
	s.drawing.OnResetTimer = s.ResetTimer

	// If this level game from a signed LevelPack, inform the canvas.
	if s.LevelPack != nil && dpp.Driver.IsLevelPackSigned(s.LevelPack) {
		s.drawing.IsSignedLevelPack = s.LevelPack
	}

	// Given a filename or map data to play?
	if s.Level != nil {
		log.Debug("PlayScene.Setup: received level from scene caller")
		s.drawing.LoadLevel(s.Level)
		if err := s.installActors(); err != nil {
			log.Error("InstallActors: %s", err)
		}
	} else if s.Filename != "" {
		loadscreen.SetSubtitle("Opening: " + s.Filename)
		log.Debug("PlayScene.Setup: loading map from file %s", s.Filename)
		// NOTE: s.LoadLevel also calls s.drawing.InstallActors
		s.LoadLevel(s.Filename)
	}

	if s.Level == nil {
		log.Debug("PlayScene.Setup: no grid given, initializing empty grid")
		s.Level = level.New()
		s.drawing.LoadLevel(s.Level)
		if err := s.installActors(); err != nil {
			log.Error("InstallActors: %s", err)
		}
	}

	// Choose a death barrier in case the user falls off the map,
	// so they don't fall forever.
	worldSize := s.Level.Chunker.WorldSize()
	s.deathBarrier = worldSize.H + 1000
	log.Debug("Death barrier at %d", s.deathBarrier)

	// Set the loading screen text with the level metadata.
	loadscreen.SetSubtitle(
		s.Level.Title,
		"by "+s.Level.Author,
	)

	// Load all actor scripts.
	s.drawing.SetScriptSupervisor(s.scripting)
	if err := s.scripting.InstallScripts(s.Level); err != nil {
		log.Error("PlayScene.Setup: failed to InstallScripts: %s", err)
	}

	// Load in the player character.
	s.setupPlayer(balance.PlayerCharacterDoodad)

	if s.CanEdit {
		d.Flash("Entered Play Mode. Press 'E' to edit this map.")
	} else {
		d.FlashError("%s", s.Level.Title)
	}

	// Pre-cache all bitmap images from the level chunks.
	// Note: we are not running on the main thread, so SDL2 Textures
	// don't get created yet, but we do the full work of caching bitmap
	// images which later get fed directly into SDL2 saving speed at
	// runtime, + the bitmap generation is pretty wicked fast anyway.
	loadscreen.PreloadAllChunkBitmaps(s.Level.Chunker)
	s.drawing.LoadUnloadChunks(true)

	// Gamepad: put into GameplayMode.
	gamepad.SetMode(gamepad.GameplayMode)

	// Run all the actor scripts' main() functions.
	if err := s.drawing.InstallScripts(); err != nil {
		log.Error("PlayScene.Setup: failed to drawing.InstallScripts: %s", err)
	}

	s.startTime = time.Now()
	s.perfectRun = true
	s.running = true

	// // Cap the canvas size in case the user has an ultra HD monitor that's bigger
	// // than a bounded level's limits.
	// s.screen.Compute(d.Engine)
	// s.PlaceResizeCanvas()

	return nil
}

// Common function to install the actors into the level.
//
// InstallActors may return an error if doodads were not found - because the
// player is on the free version and can't load attached doodads from nonsigned
// files.
func (s *PlayScene) installActors() error {
	if err := s.drawing.InstallActors(s.Level.Actors); err != nil {
		summary := "This level references some doodads that were not found:"
		if strings.Contains(err.Error(), plus.ErrRegisteredFeature.Error()) {
			summary = "This level contains embedded doodads, but this is not\n" +
				"available in the free version of the game. The following\n" +
				"doodads could not be loaded:"
		}
		modal.Alert("%s\n\n%s", summary, err).WithTitle("Level Errors")
		return fmt.Errorf("EditorScene.LoadLevel: InstallActors: %s", err)
	}
	return nil
}

// PlaceResizeCanvas updates the Canvas size and placement on the screen,
// e.g. if an ultra HD monitor plays a Bounded level where the entirety of a
// level bounds is on-screen, the drawing should be cut there and the
// canvas centered.
func (s *PlayScene) PlaceResizeCanvas() {
	var (
		width   = s.d.width
		height  = s.d.height
		menubar = 0
	)

	if s.menubar != nil {
		menubar = s.menubar.Size().H
		height -= menubar
	}

	if s.Level != nil && s.Level.PageType >= level.Bounded {
		if s.Level.MaxWidth < int64(width) {
			width = int(s.Level.MaxWidth)
		}
		if s.Level.MaxHeight < int64(height) {
			height = int(s.Level.MaxHeight)
		}
	}

	s.drawing.Resize(render.NewRect(width, height))
	s.drawing.MoveTo(render.Point{
		X: (s.d.width / 2) - (width / 2),
		Y: menubar,
	})
}

// Canvas returns the main level canvas - useful to call from the debug console as `d.Scene.Canvas()`
func (s *PlayScene) Canvas() *uix.Canvas {
	return s.drawing
}

// ScriptSupervisor returns the scripting supervisor for play mode.
func (s *PlayScene) ScriptSupervisor() *scripting.Supervisor {
	return s.scripting
}

// SetPlayerCharacter changes the doodad used for the player, by destroying the
// current player character and making it from scratch.
func (s *PlayScene) SetPlayerCharacter(filename string) {
	// Record the player position and size and back up their inventory.
	var (
		spawn     = s.Player.Position()
		inventory = s.Player.Inventory()
	)

	// TODO: to account for different height players, the position ought to be
	// adjusted so the player doesn't clip and fall thru the floor.
	spawn.Y -= 20 // work-around

	s.Player.Destroy()
	s.drawing.RemoveActor(s.Player)

	log.Info("SetPlayerCharacter: %s", filename)
	s.installPlayerDoodad(filename, spawn, render.Rect{})
	if err := s.drawing.InstallScripts(); err != nil {
		log.Error("SetPlayerCharacter: InstallScripts: %s", err)
	}

	// Restore their inventory.
	for item, qty := range inventory {
		s.Player.AddItem(item, qty)
	}
}

// ResetTimer sets the level elapsed timer back to zero.
func (s *PlayScene) ResetTimer() {
	s.startTime = time.Now()
}

// setupPlayer creates and configures the Player Character in the level.
func (s *PlayScene) setupPlayer(playerCharacterFilename string) {
	// Find the spawn point of the player. Search the level for the
	// "start-flag.doodad"
	var (
		isStartFlagCharacter bool

		spawn     render.Point
		centerIn  render.Rect
		flag      = &level.Actor{}
		flagSize  = render.NewRect(86, 86) // TODO: start-flag.doodad is 86x86 px
		flagCount int
	)
	for actorID, actor := range s.Level.Actors {
		if actor.Filename == "start-flag.doodad" {
			// Support alternative player characters: if the Start Flag is linked
			// to another actor, that actor becomes the player character.
			for _, linkID := range actor.Links {
				if linkedActor, ok := s.Level.Actors[linkID]; ok {
					playerCharacterFilename = linkedActor.Filename
					isStartFlagCharacter = true
					log.Info("Playing as: %s", playerCharacterFilename)
					break
				}
			}

			// TODO: start-flag.doodad is 86x86 pixels but we can't tell that
			// from right here.
			log.Info("Found start-flag.doodad at %s (ID %s)", actor.Point, actorID)
			flag = actor
			flagCount++
			break
		}
	}

	// If the user is cheating for the player character, mark the
	// session cheated already. e.g. "Play as Bird" cheat would let
	// them just fly to the goal in levels that don't link their
	// Start Flag to a specific character.
	if !isStartFlagCharacter && !balance.IsPlayerCharacterDefault() {
		log.Warn("Mark session as cheated: the player spawned as %s instead of default", playerCharacterFilename)
		s.SetCheated()
	}

	// The Start Flag becomes the player's initial checkpoint.
	s.lastCheckpoint = flag.Point

	if !s.SpawnPoint.IsZero() {
		spawn = s.SpawnPoint
	} else {
		spawn = flag.Point
		centerIn = render.Rect{
			W: flagSize.W,
			H: flagSize.H,
		}
	}

	// Surface warnings around the spawn flag.
	if flagCount == 0 {
		s.d.FlashError("Warning: this level contained no Start Flag.")
	} else if flagCount > 1 {
		s.d.FlashError("Warning: this level contains multiple Start Flags. Player spawn point is ambiguous.")
	}

	s.installPlayerDoodad(playerCharacterFilename, spawn, centerIn)

	// Scroll the level canvas to center on the start point.
	scroll := render.Point{
		X: -(spawn.X - (s.d.width / 2)),
		Y: -(spawn.Y - (s.d.height / 2)),
	}
	log.Info("Scrolling level viewport to spawn (%s) location: %s", spawn, scroll)
	s.drawing.ScrollTo(scroll)
}

// Load and install the player doodad onto the level.
// Make sure the previous PLAYER was removed.
// If spawn is zero, uses the player's last spawn point.
// centerIn is optional, ignored if zero.
func (s *PlayScene) installPlayerDoodad(filename string, spawn render.Point, centerIn render.Rect) {
	// Load in the player character.
	player, err := dpp.Driver.LoadFromEmbeddable(filename, s.Level, false)
	if err != nil {
		log.Error("PlayScene.Setup: failed to load player doodad: %s", err)
		player = doodads.NewDummy(32)
	}

	// Center the player within the box of the doodad, for the Start Flag especially.
	if !centerIn.IsZero() {
		spawn = render.NewPoint(
			spawn.X+(centerIn.W/2)-(player.ChunkSize()/2),

			// Y: the bottom of the flag, 4 pixels from the floor.
			spawn.Y+centerIn.H-4-(player.ChunkSize()),
		)
	} else if spawn.IsZero() && !s.SpawnPoint.IsZero() {
		spawn = s.SpawnPoint
	}

	s.Player = uix.NewActor("PLAYER", &level.Actor{Filename: filename}, player)
	s.Player.SetInventory(true) // player always can pick up items
	s.Player.MoveTo(spawn)
	s.drawing.AddActor(s.Player)
	s.drawing.FollowActor = s.Player.ID()

	// Set up the movement physics for the player.
	s.playerPhysics = &physics.Mover{
		MaxSpeed:     physics.NewVector(balance.PlayerMaxVelocity, balance.PlayerMaxVelocity),
		Acceleration: balance.PlayerAcceleration,
		Friction:     balance.PlayerFriction,
	}
	s.slipperyPhysics = &physics.Mover{
		MaxSpeed:     s.playerPhysics.MaxSpeed,
		Acceleration: balance.SlipperyAcceleration,
		Friction:     balance.SlipperyFriction,
	}

	// Set up the player character's script in the VM.
	if err := s.scripting.AddLevelScript(s.Player.ID(), s.Player.Actor.Filename); err != nil {
		log.Error("PlayScene.Setup: scripting.InstallActor(player) failed: %s", err)
	}
}

// EditLevel toggles out of Play Mode to edit the level.
func (s *PlayScene) EditLevel() {
	log.Info("Edit Mode, Go!")

	// If they didn't come from the Level Editor originally, e.g. they are in Story Mode,
	// confirm they want the editor in case they accidentally hit the "E" key due to
	// its proximity to the WASD keys.
	if !s.CanEdit {
		modal.Confirm("Open this level in the editor?").Then(s.doEditLevel)
	} else {
		s.doEditLevel()
	}
}

// Common logic to transition into the Editor.
func (s *PlayScene) doEditLevel() {
	gamepad.SetMode(gamepad.MouseMode)
	s.d.Goto(&EditorScene{
		Filename:               s.Filename,
		Level:                  s.Level,
		RememberScrollPosition: s.RememberScrollPosition,
	})
}

// RestartLevel starts the level over again.
func (s *PlayScene) RestartLevel() {
	log.Info("Restart Level")
	s.d.Goto(&PlayScene{
		LevelPack: s.LevelPack,
		Filename:  s.Filename,
		Level:     s.Level,
		CanEdit:   s.CanEdit,
	})
}

// SetCheckpoint sets the player's checkpoint.
func (s *PlayScene) SetCheckpoint(where render.Point) {
	s.lastCheckpoint = where
}

// RetryCheckpoint moves the player back to their last checkpoint.
func (s *PlayScene) RetryCheckpoint() {
	// Grant the player invulnerability for 5 seconds
	s.godModeUntil = time.Now().Add(balance.RespawnGodModeTimer)

	log.Info("Move player back to last checkpoint")
	s.Player.MoveTo(s.lastCheckpoint)
	s.running = true
}

// BeatLevel handles the level success condition.
func (s *PlayScene) BeatLevel() {
	s.d.Flash("Hurray!")
	s.ShowEndLevelModal(
		true,
		"Level Completed",
		"Congratulations on clearing the level!",
	)
}

/*
FailLevel handles a level failure triggered by a doodad or fire pixel.

If the Survival GameRule is set, this ends the level with a note on how long the
player had survived for and they get a silver rating.
*/
func (s *PlayScene) FailLevel(message string) {
	if s.Player.Invulnerable() || s.godMode || s.godModeUntil.After(time.Now()) {
		return
	}
	s.SetImperfect()
	s.d.FlashError(message)

	if s.Level.GameRule.Survival {
		s.ShowEndLevelModal(
			false,
			"You've died!",
			fmt.Sprintf(
				"%s\nCongrats on surviving for %s!",
				message,
				savegame.FormatDuration(time.Since(s.startTime)),
			),
		)
		return
	}

	s.ShowEndLevelModal(
		false,
		"You've died!",
		message,
	)
}

// DieByFire ends the level by "fire", or w/e the swatch is named.
func (s *PlayScene) DieByFire(name string) {
	s.FailLevel(fmt.Sprintf("Watch out for %s!", name))
}

// SetImperfect sets the perfectRun flag to false and changes the icon for the timer.
func (s *PlayScene) SetImperfect() {
	if s.cheated {
		return
	}

	s.perfectRun = false
	if s.timerPerfectImage != nil {
		s.timerPerfectImage.Hide()
	}
	if s.timerImperfectImage != nil {
		s.timerImperfectImage.Show()
	}
}

// SetCheated marks the level as having been cheated. The developer shell will call
// this if the user enters a cheat code during gameplay.
func (s *PlayScene) SetCheated() {
	s.cheated = true
	s.perfectRun = false

	// Hide both timer icons.
	if s.timerPerfectImage != nil {
		s.timerPerfectImage.Hide()
	}
	if s.timerImperfectImage != nil {
		s.timerImperfectImage.Hide()
	}
}

// GetCheated gives read-only access to tell if you have been cheating. However, by
// querying this in the dev console during gameplay, you would be marked as cheating. ;)
func (s *PlayScene) GetCheated() bool {
	return s.cheated
}

// GetPerfect gives read-only access to the perfectRun flag.
func (s *PlayScene) GetPerfect() bool {
	return s.perfectRun
}

// ShowEndLevelModal centralizes the EndLevel modal config.
// This is the common handler function between easy methods such as
// BeatLevel, FailLevel, and DieByFire.
func (s *PlayScene) ShowEndLevelModal(success bool, title, message string) {
	// Always restore the cursor.
	cursor.Current = cursor.NewPointer(s.d.Engine)

	config := modal.ConfigEndLevel{
		Engine:            s.d.Engine,
		Success:           success,
		OnRestartLevel:    s.RestartLevel,
		OnRetryCheckpoint: s.RetryCheckpoint,
		OnExitToMenu: func() {
			gamepad.SetMode(gamepad.MouseMode)
			s.d.Goto(&MainScene{})
		},
	}

	if s.CanEdit {
		config.OnEditLevel = s.EditLevel
	}

	// Survival Mode failure? The level is considered completed even if you
	// die (silver high score) but the default button should be Retry rather
	// than Next Level.
	var survivalFailure bool
	if !success && s.Level.GameRule.Survival {
		survivalFailure = true
		success = true // level is completed
	}

	// Beaten the level?
	if success {
		config.OnRetryCheckpoint = nil

		// Are we in a levelpack?
		if s.LevelPack != nil {
			// Update the savegame to mark the level completed.
			save, err := savegame.GetOrCreate()
			if err != nil {
				log.Warn("Load savegame file: %s", err)
			}

			log.Info("Mark level '%s' from pack '%s' as completed", s.Filename, s.LevelPack.Filename)
			if !s.cheated {
				elapsed := time.Since(s.startTime)
				highscore := save.NewHighScore(s.LevelPack.Filename, s.Filename, s.Level.UUID, s.perfectRun, elapsed, s.Level.GameRule)
				if highscore {
					s.d.Flash("New record!")
					config.NewRecord = true
					config.IsPerfect = s.perfectRun
					config.TimeElapsed = elapsed
				}
			} else {
				// Player has cheated! Mark the level completed but grant no high score.
				save.MarkCompleted(s.LevelPack.Filename, s.Filename, s.Level.UUID)
			}

			// Save the player's scores file.
			if err = save.Save(); err != nil {
				log.Error("Couldn't save game: %s", err)
			}

			// Show the "Next Level" button if there is a sequel to this level.
			for i, level := range s.LevelPack.Levels {
				i := i
				level := level

				if level.Filename == s.Filename && i < len(s.LevelPack.Levels)-1 {
					// Show "Next" button!
					config.OnNextLevel = func() {
						nextLevel := s.LevelPack.Levels[i+1]
						log.Info("Advance to next level: %s", nextLevel.Filename)
						s.d.PlayFromLevelpack(s.LevelPack, nextLevel)
					}
				}
			}
		}
	}

	// Survival Mode failures: the Retry buttons should be higher
	// priority than Next Level but they still get the (pity)
	// Next Level button.
	if survivalFailure {
		config.OnPityNextLevel = config.OnNextLevel
		config.OnNextLevel = nil
	}

	// Show the modal.
	modal.EndLevel(config, title, message)

	// Stop the simulation.
	s.running = false
}

// Loop the editor scene.
func (s *PlayScene) Loop(d *Doodle, ev *event.State) error {
	// Skip if still loading.
	if loadscreen.IsActive() {
		return nil
	}

	// Inform the gamepad controller whether we have antigravity controls.
	gamepad.PlayModeAntigravity = s.antigravity || !s.Player.HasGravity()

	// Update debug overlay values.
	*s.debWorldIndex = s.drawing.WorldIndexAt(render.NewPoint(ev.CursorX, ev.CursorY)).String()
	*s.debPosition = s.Player.Position().String() + " vel " + s.Player.Velocity().String()
	*s.debViewport = s.drawing.Viewport().String()
	*s.debScroll = s.drawing.Scroll.String()
	inside, outside := s.drawing.LoadUnloadMetrics()
	*s.debLoadUnload = fmt.Sprintf("%d in %d out %d cached %d gc", inside, outside, s.drawing.Chunker().CacheSize(), s.drawing.Chunker().GCSize())

	// Update the timer.
	s.timerLabel.Text = savegame.FormatDuration(time.Since(s.startTime))

	s.Supervisor.Loop(ev)

	// Has the window been resized?
	if ev.WindowResized || s.drawing.Point().IsZero() {
		s.PlaceResizeCanvas()
		s.screen.Resize(render.NewRect(d.width, d.height))
		return nil
	}

	// Switching to Edit Mode?
	if keybind.GotoEdit(ev) {
		s.EditLevel()
		return nil
	}

	// Is the simulation still running?
	if s.running {
		// Loop the script supervisor so timeouts/intervals can fire in scripts.
		if err := s.scripting.Loop(); err != nil {
			log.Error("PlayScene.Loop: scripting.Loop: %s", err)
		}

		// Touch regions.
		s.LoopTouchable(ev)

		// Hide the mouse cursor if a gameplay input was received.
		if keybind.Right(ev) || keybind.Left(ev) || keybind.Up(ev) || keybind.Down(ev) ||
			keybind.Use(ev) {
			cursor.Current = cursor.NoCursor
		} else if s.lastCursor != shmem.Cursor {
			cursor.Current = cursor.NewPointer(s.d.Engine)
		}
		s.lastCursor = shmem.Cursor

		s.movePlayer(ev)
		if err := s.drawing.Loop(ev); err != nil {
			log.Error("Drawing loop error: %s", err.Error())
		}

		// Check if the player hit the death barrier.
		if s.Player.Position().Y > s.deathBarrier {
			// The player must die to avoid the softlock of falling forever.
			s.godMode = false
			s.Player.SetInvulnerable(false)
			s.DieByFire("falling off the map")
		}

		// Update the inventory HUD.
		s.computeInventory()
	}

	return nil
}

// Draw the pixels on this frame.
func (s *PlayScene) Draw(d *Doodle) error {
	// Skip if still loading.
	if loadscreen.IsActive() {
		return nil
	}

	// Clear the canvas and fill it with white.
	d.Engine.Clear(balance.WindowBackground)

	// Draw the canvas widget.
	s.drawing.Present(d.Engine, s.drawing.Point())

	// Draw out bounding boxes.
	if DebugCollision {
		for _, actor := range s.drawing.Actors() {
			d.DrawCollisionBox(s.drawing, actor)
		}
	}

	// Bug: sometimes (especially after cheating) if you restart a level
	// properly, cheated=false perfectRun=true but the perfectRunIcon
	// would not be showing.
	if !s.cheated && s.perfectRun && s.timerPerfectImage.Hidden() {
		s.timerPerfectImage.Show()
	}

	// Draw the UI screen and any widgets that attached to it.
	s.screen.Compute(d.Engine)
	s.screen.Present(d.Engine, render.Origin)

	// Draw the Edit button.
	var (
		canSize = s.drawing.Size()
		size    = s.editButton.Size()
		padding = 8
	)
	s.editButton.MoveTo(render.Point{
		X: canSize.W - size.W - padding,
		Y: canSize.H - size.H - padding,
	})
	s.editButton.Present(d.Engine, s.editButton.Point())

	// Visualize the touch regions?
	s.DrawTouchable()

	// Let Supervisor draw menus
	s.Supervisor.Present(d.Engine)

	return nil
}

// Drawing returns the private world drawing, for debugging with the console.
func (s *PlayScene) Drawing() *uix.Canvas {
	return s.drawing
}

// LoadLevel loads a level from disk.
//
// If the PlayScene was called with a LevelPack, it will check there
// first before the usual locations.
//
// The usual locations are: embedded bindata, ./assets folder on disk,
// and user content finally.
func (s *PlayScene) LoadLevel(filename string) error {
	s.Filename = filename

	var (
		lvl *level.Level
		err error
	)

	// Are we playing out of a levelpack?
	if s.LevelPack != nil {
		levelbin, err := s.LevelPack.GetFile("levels/" + filename)
		if err != nil {
			log.Error("Error reading levels/%s from zip: %s", filename, err)
		}

		lvl, err = level.FromJSON(filename, levelbin)
		if err != nil {
			log.Error("PlayScene.LoadLevel(%s) from zipfile: %s", filename, err)
		}

		log.Info("PlayScene.LoadLevel: found %s in LevelPack zip data", filename)
	}

	// Try the usual suspects.
	if lvl == nil {
		log.Info("PlayScene.LoadLevel: trying the usual places")
		lvl, err = level.LoadFile(filename)
		if err != nil {
			return fmt.Errorf("PlayScene.LoadLevel(%s): %s", filename, err)
		}
	}

	s.Level = lvl
	s.drawing.LoadLevel(s.Level)
	if err := s.installActors(); err != nil {
		return err
	}

	return nil
}

// Destroy the scene.
func (s *PlayScene) Destroy() error {
	// Free SDL2 textures. Note: if they are switching to the Editor, the chunks still have
	// their bitmaps cached and will regen the textures as needed.
	s.drawing.Destroy()

	// Free inventory doodad textures.
	for _, can := range s.invenDoodads {
		log.Info("Destroy inventory doodad: %s", can)
		can.Destroy()
	}

	return nil
}
